Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   try {
      $jsonSecrets = ${env:HOSTMAINT_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:HOSTMAINT_SECRET does not look to be defined"
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
      $jsonSecrets = ${env:HOSTMAINT_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:HOSTMAINT_SECRET does not look to be defined"
   }

   if (${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: Event - $($cloudEvent.type)"
   }
   $hostName = $cloudEventData.Host.Name
   $moRef = New-Object VMware.Vim.ManagedObjectReference
   $moRef.Type = $cloudEventData.Host.Host.Type
   $moRef.Value = $cloudEventData.Host.Host.Value

   try {
      $hostMoRef = Get-View $moRef
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: Could not retrieve host view by moRef"
      throw $_
   }

   try {
      $alarmManager = Get-View AlarmManager
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: Could not retrieve AlarmManager object"
      throw $_
   }

   if ($cloudEvent.type -eq "com.vmware.vsphere.EnteredMaintenanceModeEvent.v0") {
      # Disable alarm actions on the host
      Write-Host "$(Get-Date) - Disabling alarm actions on host: $hostName"
      try {
         $alarmManager.EnableAlarmActions($hostMoRef.MoRef, $false)
      }
      catch {
         Write-Host "$(Get-Date) - ERROR: Could not disable alarm actions"
         throw $_
      }
   }

   if ($cloudEvent.type -eq "com.vmware.vsphere.ExitMaintenanceModeEvent.v0") {
      # Enable alarm actions on the host
      Write-Host "$(Get-Date) - Enabling alarm actions on host: $hostName"
      try {
         $alarmManager.EnableAlarmActions($hostMoRef.MoRef, $true)
      }
      catch {
         Write-Host "$(Get-Date) - ERROR: Could not enable alarm actions"
         throw $_
      }
  }

   Write-Host "$(Get-Date) - kn-pcli-hostmaint-alarms operation complete ...`n"

   Write-Host "$(Get-Date) - Handler Processing Completed ...`n"
}
