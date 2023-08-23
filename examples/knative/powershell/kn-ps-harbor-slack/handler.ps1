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

   try {
      $jsonSecrets = ${env:SLACK_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:SLACK_SECRET does not look to be defined"
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: K8s Secrets:`n${env:SLACK_SECRET}`n"

      Write-Host "$(Get-Date) - DEBUG: CloudEventData`n $(${cloudEventData} | Out-String)`n"
   }

   # Construct Slack message object
   $payload = @{
      attachments = @(
         @{
            pretext = $(${jsonSecrets}.SLACK_MESSAGE_PRETEXT);
            fields = @(
               @{
                  title = "Event Type";
                  value = $cloudEvent.type;
                  short = "false";
               }
               @{
                  title = "DateTime in UTC";
                  value = $cloudEvent.time;
                  short = "false";
               }
               @{
                  title = "Unique Identifier";
                  value = $cloudEvent.id;
                  short = "false";
               }
               @{
                  title = "Username";
                  value = $cloudEvent.extensions.operator; # WIP
                  short = "false";
               }
               @{
                  title = "Repository Name";
                  value = $cloudEventData.repository.repo_full_name;
                  short = "false";
               }
               @{
                  title = "Repository Type";
                  value = $cloudEventData.repository.repo_type;
                  short = "false";
               }
               @{
                  title = "Image Tag";
                  value = $cloudEventData.resources[0].tag;
                  short = "false";
               }
               @{
                  title = "Image Resource Data";
                  value = $cloudEventData.resources[0].resource_url;
                  short = "false";
               }
               @{
                  title = "Image Digest";
                  value = $cloudEventData.resources[0].digest;
                  short = "false";
               }
            )
            footer = "Powered by VEBA";
            footer_icon = "https://raw.githubusercontent.com/vmware-samples/vcenter-event-broker-appliance/development/logo/veba_icon_only.png";
         }
      )
   }

   # Convert Slack message object into JSON
   $body = $payload | ConvertTo-Json -Depth 5

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: `"$body`""
   }

   Write-Host "$(Get-Date) - Sending Webhook payload to Slack ..."
   $ProgressPreference = "SilentlyContinue"

   try {
      Invoke-WebRequest -Uri $(${jsonSecrets}.SLACK_WEBHOOK_URL) -Method POST -ContentType "application/json" -Body $body
   } catch {
      throw "$(Get-Date) - Failed to send Slack Message: $($_)"
   }

   Write-Host "$(Get-Date) - Successfully sent Webhook ..."

}
