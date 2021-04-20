Function Process-Handler {
   param(
      [Parameter(Position=0,Mandatory=$true)][CloudNative.CloudEvents.CloudEvent]$CloudEvent
   )

   # Decode CloudEvent
   $cloudEventData = $cloudEvent | Read-CloudEventJsonData -ErrorAction SilentlyContinue -Depth 10
   if($cloudEventData -eq $null) {
      $cloudEventData = $cloudEvent | Read-CloudEventData
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "DEBUG: K8s Secrets:`n${env:SLACK_SECRET}`n"

      Write-Host "DEBUG: CloudEventData`n $(${cloudEventData} | Out-String)`n"
   }

   if(${env:SLACK_SECRET}) {
      $jsonSecrets = ${env:SLACK_SECRET} | ConvertFrom-Json
   } else {
      Write-Host "K8s secrets `$env:SLACK_SECRET does not look to be defined"
      break
   }

   # Send VM changes
   Write-Host "Detected change to $($cloudEvent.Subject) ..."

   $payload = @{
      attachments = @(
         @{
            pretext = ":rotating_light: Virtual Machine PoweredOff Alert from :veba: Knative Function :rotating_light:";
            fields = @(
               @{
                     title = "VM";
                     value = $cloudEventData.Vm.Name;
                     short = "false";
               }
               @{
                     title = "Username";
                     value = $cloudEventData.UserName;
                     short = "false";
               }
               @{
                     title = "DateTime";
                     value = $cloudEventData.CreatedTime;
                     short = "false";
               }
            )
         }
      )
   }

   $body = $payload | ConvertTo-Json -Depth 5

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "DEBUG: `"$body`""
   }

   Write-Host "Sending Webhook payload to Slack ..."
   $ProgressPreference = "SilentlyContinue"
   Invoke-WebRequest -Uri $(${jsonSecrets}.SLACK_WEBHOOK_URL) -Method POST -ContentType "application/json" -Body $body
   Write-Host "Successfully sent Webhook ..."
}
