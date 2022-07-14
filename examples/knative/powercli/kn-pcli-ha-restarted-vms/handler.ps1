Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   try {
      $jsonSecrets = ${env:HA_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:HA_SECRET does not look to be defined"
   }

   # Extract all tag secrets for ease of use in function
   $VCENTER_SERVER = ${jsonSecrets}.VCENTER_SERVER
   $VCENTER_USERNAME = ${jsonSecrets}.VCENTER_USERNAME
   $VCENTER_PASSWORD = ${jsonSecrets}.VCENTER_PASSWORD
   $VCENTER_CERTIFICATE_ACTION = ${jsonSecrets}.VCENTER_CERTIFICATE_ACTION

   # Configure TLS 1.2/1.3 support as this is required for latest vSphere release
   [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor [System.Net.SecurityProtocolType]::Tls12 -bor [System.Net.SecurityProtocolType]::Tls13

   Write-Host "$(Get-Date) - Configuring PowerCLI Configuration Settings`n"
   Set-PowerCLIConfiguration -InvalidCertificateAction:${VCENTER_CERTIFICATE_ACTION} -ParticipateInCeip:$true -Confirm:$false

   Write-Host "$(Get-Date) - Connecting to vCenter Server $VCENTER_SERVER`n"

   try {
      Connect-VIServer -Server $VCENTER_SERVER -User $VCENTER_USERNAME -Password $VCENTER_PASSWORD
   }
   catch {
      Write-Error "$(Get-Date) - ERROR: Failed to connect to vCenter Server"
      throw $_
   }

   Write-Host "$(Get-Date) - Successfully connected to $VCENTER_SERVER`n"

   Write-Host "$(Get-Date) - Init Processing Completed`n"
}

Function Process-Shutdown {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Shutdown`n"

   Write-Host "$(Get-Date) - Disconnecting from vCenter Server`n"

   try {
      Disconnect-VIServer * -Confirm:$false
   }
   catch {
      Write-Error "$(Get-Date) - Error: Failed to Disconnect from vCenter Server"
   }

   Write-Host "$(Get-Date) - Shutdown Processing Completed`n"
}

Function Process-Handler {
   [CmdletBinding()]
   param(
      [Parameter(Position=0,Mandatory=$true)][CloudNative.CloudEvents.CloudEvent]$CloudEvent
   )

   # Decode CloudEvent
   try {
      $cloudEventData = $cloudEvent | Read-CloudEventJsonData -Depth 10
   }
   catch {
      throw "`nPayload must be JSON encoded"
   }

   try {
      $jsonSecrets = ${env:HA_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:HA_SECRET does not look to be defined"
   }

   # Main processing of gathering specific events
   $report = @()
   $eventnumber = 1000
   # get vCenter EventManager
   try {
      $si = get-view ServiceInstance
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: Could not retrieve Service Instance"
      throw $_
   }

   try {
      $em = get-view $si.Content.EventManager
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: Could not retrieve Event Manager"
      throw $_
   }

   # Create Event Filter Spec
   $EventFilterSpec = New-Object VMware.Vim.EventFilterSpec
   # Get specific Events based on EventTypeIDs for EventFilterSpec
   $tgtEvtTypeIDs = "com.vmware.vc.ha.VmRestartedByHAEvent", "com.vmware.vc.HA.DasHostFailedEvent"
   $EventFilterSpec.EventTypeId = $tgtEvtTypeIDs
   # Time range to look for specific IDs (current day specified) for EventFilterSpec
   $EventFilterSpecByTime = New-Object VMware.Vim.EventFilterSpecByTime
   # [datetime]::Today is midnight on the date the ClusterFailoverActionCompletedEvent fired
   $EventFilterSpecByTime.BeginTime = [datetime]::Today
   # Searches the entire 24 hour period on the date the ClusterFailoverActionCompletedEvent fired
   $EventFilterSpecByTime.EndTime = ([datetime]::Today).AddDays(+1)
   $EventFilterSpec.Time = $EventFilterSpecByTime

   # Create an Event Collector and loop through events storing them into an array
   try {
      $eCollector = Get-View ($em.CreateCollectorForEvents($EventFilterSpec))
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: Could not create event collector"
      throw $_
   }

   try {
      $events = $eCollector.ReadNextEvents($eventnumber)
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: Could not retrieve events with event collector"
      throw $_
   }

   While ($events) {
      $events | %{
         $report += $_
      }
      try {
         $events = $eCollector.ReadNextEvents($eventnumber)
      }
      catch {
         Write-Host "$(Get-Date) - ERROR: Could not retrieve events with event collector"
         throw $_
      }
   }

   $eCollector.DestroyCollector()

   $SMTP_SERVER = ${jsonSecrets}.SMTP_SERVER
   $SMTP_PORT = ${jsonSecrets}.SMTP_PORT
   $SMTP_USERNAME = ${jsonSecrets}.SMTP_USERNAME
   $SMTP_PASSWORD = ${jsonSecrets}.SMTP_PASSWORD
   $EMAIL_TO = ${jsonSecrets}.EMAIL_TO
   $EMAIL_FROM = ${jsonSecrets}.EMAIL_FROM
   $DISPLAY_HOST_FQDN = ${jsonSecrets}.DISPLAY_HOST_FQDN

   # Retrieve Host name of ESXi host which crashed - used for email report
   if ($report | Where-Object{$_.EventTypeId -eq "com.vmware.vc.HA.DasHostFailedEvent"}) {
      if ($DISPLAY_HOST_FQDN -eq $false){
         # Displays just hostname - strips FQDN
         $vmHost = ($report | Where-Object {$_.EventTypeId -eq "com.vmware.vc.HA.DasHostFailedEvent"} | Select-Object ObjectName -ExpandProperty ObjectName -First 1).split(".")[0]
      } else {
         # Displays host FQDN
         $vmHost = $report | Where-Object {$_.EventTypeId -eq "com.vmware.vc.HA.DasHostFailedEvent"} | Select-Object ObjectName -ExpandProperty ObjectName -First 1
      }
   }

   # Set up fields for email body and send email - only sending VMname, time VM restarted on another host, and VM description
   $output = $report | Where-Object {$_.ObjectType -eq "VirtualMachine" } | Select-Object ObjectName, @{N = "Date"; E = { $_.CreatedTime } }, @{N = "Description"; E = { (" - " + (Get-view -id $_.vm.vm).config.annotation | Out-String) } } | Sort-Object ObjectName
   $msgBody = $output | ConvertTo-Html | Out-String
   $subject = "** $vmHost Failure - VMs Restarted **"

   if (${env:HA_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: From - $($EMAIL_FROM)"
      Write-Host "$(Get-Date) - DEBUG: To - $($EMAIL_TO)"
   }

   Write-Host "$(Get-Date) - Sending notification for host $($vmHost)"

   # If defined in the config file, send via authenticated SMTP, otherwise use standard SMTP
   if ($SMTP_PASSWORD.length -gt 0 -and $SMTP_USERNAME.length -gt 0)
   {
      $password = ConvertTo-SecureString "$($SMTP_PASSWORD)" -AsPlainText -Force
      $credential = New-Object System.Management.Automation.PSCredential($($SMTP_USERNAME), $password)
      try {
         Send-MailMessage -From $($EMAIL_FROM) -to $($EMAIL_TO)  -Subject $subject -Body $msgBody -BodyAsHtml -SmtpServer $($SMTP_SERVER) -port $($SMTP_PORT) -UseSsl -Credential $credential #-Encoding UTF32
      }
      catch {
         Write-Host "$(Get-Date) - ERROR: Could not send authenticated email"
         throw $_
      }
   }
   else
   {
      try {
         Send-MailMessage -From $($EMAIL_FROM) -to $($EMAIL_TO) -Subject $subject -Body $msgBody -BodyAsHtml -SmtpServer $($SMTP_SERVER) -port $($SMTP_PORT) -Encoding UTF32
      }
      catch {
         Write-Host "$(Get-Date) - ERROR: Could not send email"
         throw $_
      }
   }

   Write-Host "$(Get-Date) - Handler Processing Completed ...`n"
}