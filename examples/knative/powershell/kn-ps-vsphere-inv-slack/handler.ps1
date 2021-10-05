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

   # List of vSphere Inventory operations from a TaskEvent to process
   $inventoryOperationsProcess = @(
      "Folder.unregisterAndDestroy",
      "ResourcePool.destroy"
   )

   # Only process TaskEvent where DescriptionId matches $inventoryOperationsProcess list
   if($inventoryOperationsProcess.contains(${cloudEventData}.Info.DescriptionId)) {
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
               pretext = ":rotating_light: vSphere Inventory Deletion Alert from :veba: Knative Function :rotating_light:";
               fields = @(
                  @{
                        title = "EntityType";
                        value = $cloudEventData.Info.Entity.Type;
                        short = "false";
                  }
                  @{
                     title = "EntityName";
                     value = $cloudEventData.Info.EntityName;
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
}
