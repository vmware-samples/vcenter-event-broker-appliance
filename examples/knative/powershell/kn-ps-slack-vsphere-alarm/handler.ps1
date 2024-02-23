Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   Write-Host "$(Get-Date) - Init Processing Completed`n"
}

Function Process-Shutdown {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Shutdown`n"

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
   } catch {
      throw "`nPayload must be JSON encoded"
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: K8s Secrets:`n${env:SLACK_SECRET}`n"

      Write-Host "$(Get-Date) - DEBUG: CloudEventData`n $(${cloudEventData} | Out-String)`n"
   }

   # This currently will process ALL enriched vSphere Alarms based on AlarmStatusChangedEvent
   # If you wish to limit this to a specific vSphere Alarm add $CloudEvent.AlarmInfo.Name -eq "alarm-name"
   if($CloudEvent.Type -eq "com.vmware.vsphere.AlarmStatusChangedEvent.v0") {

      if($cloudEventData.To -eq "red" -or $cloudEventData.To -eq "yellow") {

         try {
            $jsonSecrets = ${env:SLACK_SECRET} | ConvertFrom-Json
         } catch {
            throw "`nK8s secrets `$env:SLACK_SECRET does not look to be defined"
         }

         # Retrieve the vSphere Object Type
         # Additional vSphere Types can be mapped to custom Slack icons
         switch($cloudEventData.Entity.Entity.Type) {
            "datastore" {
               $objectType = ":database-6178:"
            }
            default {
               $objectType = ":unknown:"
            }
         }

         # Retrieve the vSphere Alarm Name
         $alarmName = $cloudEventData.Alarm.Name

         # Retrieve the vSphere Alarm Name
         $alarmDateTime = $cloudEventData.CreatedTime

         # Retreive the vSphere Object Name
         $objectName = $cloudEventData.Entity.Name

         # Retrieve vSphere Alarm Color: yellow or red
         switch($cloudEventData.To) {
            "yellow" {
               $objectStatus = "Warning"
               $objectColor = ":warning:"
               $objectPercentage = [int]$($cloudEventData.AlarmInfo.Expression.Expression[0].Yellow)/100
            }
            "red" {
               $objectStatus = "Error"
               $objectColor = ":fail:"
               $objectPercentage = [int]$($cloudEventData.AlarmInfo.Expression.Expression[0].Red)/100
            }
         }

         # Retrieve the Alarm operator type: isAbove or isBelow
         $operatorType = $cloudEventData.AlarmInfo.Expression.Expression[0].Operator

         if($operatorType -eq "isAbove") {
            $objectOperator = ":gt:"
         } elseif ($operatorType -eq "isBelow") {
            $objectOperator = ":less-than:"
         } else {
            $objectOperator = ":none:"
         }

         if(${env:FUNCTION_DEBUG} -eq "true") {
            Write-Host -ForegroundColor yellow "DEBUG:"
            Write-Host -ForegroundColor yellow "AlarmName: $($alarmName)"
            Write-Host -ForegroundColor yellow "AlarmDataTime: $($alarmDateTime)"
            Write-Host -ForegroundColor yellow "ObjectType: $($objectType)"
            Write-Host -ForegroundColor yellow "ObjectStatus: $($objectStatus)"
            Write-Host -ForegroundColor yellow "ObjectName: $($objectName)"
            Write-Host -ForegroundColor yellow "ObjectColor: $($objectColor)"
            Write-Host -ForegroundColor yellow "ObjectOperator: $($objectOperator)"
            Write-Host -ForegroundColor yellow "ObjectPercentage: $($objectPercentage)"
         }

         $payload = @{
            attachments = @(
               @{
                  pretext = ":vsphere_icon: vSphere Alarm :alert:";
                  fields = @(
                     @{
                        title = "Alarm Name";
                        value = ":bell: ${alarmName}";
                        short = "false";
                     },
                     @{
                        title = "Object Type/Name";
                        value = "${objectType} ${objectName}";
                        short = "false";
                     },
                     @{
                        title = "Alert Type";
                        value = "${objectColor} (${objectStatus})";
                        short = "false";
                     },
                     @{
                        title = "Threshold";
                        value = "${objectOperator} ${objectPercentage}%";
                        short = "false";
                     },
                     @{
                        title = "Triggered DateTime";
                        value = ":clock2: ${alarmDateTime}";
                        short = "false";
                     }
                  )
               }
            )
         }

         $body = $payload | ConvertTo-Json -Depth 5

         if(${env:FUNCTION_DEBUG} -eq "true") {
            Write-Host -ForegroundColor yellow "DEBUG: `"$body`""
         }

         Write-Host "Sending Webhook payload to Slack ..."
         $ProgressPreference = "SilentlyContinue"
         Invoke-WebRequest -Uri $(${jsonSecrets}.SLACK_WEBHOOK_URL) -Method POST -ContentType "application/json" -Body $body
         Write-Host "Successfully sent Webhook ..."
      }
   }
}
