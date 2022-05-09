Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   try {
      $jsonSecrets = ${env:DATASTORE_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:DATASTORE_SECRET does not look to be defined"
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
      $jsonSecrets = ${env:DATASTORE_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:DATASTORE_SECRET does not look to be defined"
   }

   $alarmName = $($cloudEventData.Alarm.Name -replace "\n"," ")
   $datastoreName = $($cloudEventData.Ds.Name)
   $alarmStatus = $($cloudEventData.To)
   $vcenter = $($cloudEvent.source -replace "/sdk","")
   $datacenter = $($cloudEventData.Datacenter.Name)
   $alarmToMonitor = ${jsonSecrets}.VC_ALARM_NAME
   $datastoresToMonitor = ${jsonSecrets}.DATASTORE_NAMES

   if (${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: Alarm Name: $alarmName"
      Write-Host "$(Get-Date) - DEBUG: DS Name: $datastoreName"
      Write-Host "$(Get-Date) - DEBUG: Alarm Status: $alarmStatus"
      Write-Host "$(Get-Date) - DEBUG: vCenter: $vcenter"
      Write-Host "$(Get-Date) - DEBUG: Data Center:  $datacenter"
      Write-Host "$(Get-Date) - DEBUG: Alarm to Monitor: $alarmToMonitor"
      Write-Host "$(Get-Date) - DEBUG: Datastores to Monitor: $datastoresToMonitor"
   }

   if( ("$alarmName" -match $($alarmToMonitor)) -and ([bool]($datastoresToMonitor -match "$datastoreName")) -and ($alarmStatus -eq "yellow" -or $alarmStatus -eq "red" -or $alarmStatus -eq "green") ) {
      # Warning Email Body
      if($alarmStatus -eq "yellow") {
         $subject = "⚠️ $(${jsonSecrets}.EMAIL_SUBJECT) ⚠️ "
         $threshold = "warning"
      } elseif($alarmStatus -eq "red") {
         $subject = "☢️ $(${jsonSecrets}.EMAIL_SUBJECT) ☢️ "
         $threshold = "error"
      } elseif($alarmStatus -eq "green") {
         $subject = "$(${jsonSecrets}.EMAIL_SUBJECT)"
         $threshold = "normal"
      }

      $Body = "$alarmName $datastoreName has reached $threshold threshold.`r`n"

      if ( $threshold -ne "normal") {
         $Body = $Body + "Please log in to $($vcenter) and ensure that everything is operating as expected.`r`n"
      }

      $Body = $Body + @"
      vCenter Server: $vCenter
       Datacenter: $datacenter
       Datastore: $datastoreName
"@

      if (${env:FUNCTION_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - DEBUG: Message Subject: $($Subject)"
         Write-Host "$(Get-Date) - DEBUG: Message Body: $($Body)"
      }

      $emailTo = ${jsonSecrets}.EMAIL_TO

      # If the JSON file has a custom property email field defined, find the value
      # This is used to allow admins to add an email address as a custom property on a datastore for storage alarms independent of the EMAIL_TO value
      if (${jsonSecrets}.DATASTORE_CUSTOM_PROP_EMAIL_TO.length -gt 0)
      {
         # This object has all defined custom fields in the vCenter
         try {
            $customFieldMgr = Get-View ($global:DefaultVIServer.ExtensionData.Content.CustomFieldsManager)
         }
         catch {
            Write-Host "$(Get-Date) - ERROR: unable to retrieve CustomFieldsManager view`n"
            throw $_
         }

         try {
            $datastoreView = Get-View -ViewType Datastore -Property Name, Value -Filter @{"name"=$datastoreName}
         }
         catch {
            Write-Host "$(Get-Date) - ERROR: unable to retrieve Datastore view view`n"
            throw $_
         }

         # Build 2 hash tables for the key-value pairs in the Custom Fields Manager, one to search by Custom Field ID and one to search by Name
         $customKeyLookup = @{}
         $customNameLookup = @{}
         $customFieldMgr.Field | ForEach-Object {
               $customKeyLookup.Add($_.Key, $_.Name)
               $customNameLookup.Add($_.Name, $_.Key)
         }

         # This is the custom field that we're looking to pull an email address out of
         $emailKey = $customNameLookup[$(${jsonSecrets}.DATASTORE_CUSTOM_PROP_EMAIL_TO)]
         if(${env:FUNCTION_DEBUG} -eq "true") {
            Write-Host "$(Get-Date) - DEBUG: custom prop: $(${jsonSecrets}.DATASTORE_CUSTOM_PROP_EMAIL_TO)"
            Write-Host "$(Get-Date) - DEBUG: email Key: $emailKey"
         }

         #If we find one, this is the email address we will add to the "To" field in the email
         $addEmailAddress = ""
         foreach ($row in $datastoreView.Value) {
               Write-Host "$(Get-Date) - INFO: Datastore" $datastoreName "has Custom Field:" $customKeyLookup[$row.Key] "with value:" $row.Value "`n"
               if ($row.Key -eq $emailKey)
               {
                  if(${env:FUNCTION_DEBUG} -eq "true") {
                     write-host "$(Get-Date) - DEBUG: Found key" $emailKey "with value" $row.value
                  }
                  $addEmailAddress = $row.value
                  break
               }
         }

         if ($addEmailAddress.length -gt 0){
               $emailTo = $emailTo + $addEmailAddress
         }
         else {
               Write-Host "$(Get-Date) - WARN: DATASTORE_CUSTOM_PROP_EMAIL_TO value '"${jsonSecrets}.DATASTORE_CUSTOM_PROP_EMAIL_TO "' found in JSON config but not found on datastore"
         }

      }

      Write-Host "$(Get-Date) - Sending notification to $($emailTo)  ...`n"
      # If defined in the config file, send via authenticated SMTP, otherwise use standard SMTP
      if (${jsonSecrets}.SMTP_PASSWORD.length -gt 0 -and ${jsonSecrets}.SMTP_USERNAME.length -gt 0)
      {
         $password = ConvertTo-SecureString "$(${jsonSecrets}.SMTP_PASSWORD)" -AsPlainText -Force
         $credential = New-Object System.Management.Automation.PSCredential($(${jsonSecrets}.SMTP_USERNAME), $password)
         try {
            Send-MailMessage -From $(${jsonSecrets}.EMAIL_FROM) -to $($emailTo)  -Subject $Subject -Body $Body -SmtpServer $(${jsonSecrets}.SMTP_SERVER) -port $(${jsonSecrets}.SMTP_PORT) -UseSsl -Credential $credential -Encoding UTF32
         }
         catch {
            Write-Host "$(Get-Date) - ERROR: Unable to send email message`n"
            throw $_
         }
      }
      else
      {
         try {
            Send-MailMessage -From $(${jsonSecrets}.EMAIL_FROM) -to $($emailTo) -Subject $Subject -Body $Body -SmtpServer $(${jsonSecrets}.SMTP_SERVER) -port $(${jsonSecrets}.SMTP_PORT) -Encoding UTF32
         }
         catch {
            Write-Host "$(Get-Date) - ERROR: Unable to send email message`n"
            throw $_
         }
      }
   }

   Write-Host "$(Get-Date) - datastore-usage-email operation complete ...`n"

   Write-Host "$(Get-Date) - Handler Processing Completed ...`n"
}
