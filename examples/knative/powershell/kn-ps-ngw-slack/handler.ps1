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
               pretext = ":tada: New VMC SDDC Provisioned :tada:";
               fields = @(
                  @{
                     title = "SDDC";
                     value = $cloudEventData.resource_name;
                     short = "false";
                  }
                  @{
                     title = "Org";
                     value = $cloudEventData.org_name;
                     short = "false";
                  }
                  @{
                     title = "User";
                     value = $cloudEventData.message_username;
                     short = "false";
                  }
                  @{
                     title = "URL";
                     value ="<$($cloudEvent.source)|*Click to open SDDC*>";
                     short = "false";
                  }
               )
               footer_icon = "https://raw.githubusercontent.com/vmware-samples/vcenter-event-broker-appliance/development/logo/veba_otto_the_orca_320x320.png";
               footer = "Powered by VEBA";
         }
      )
   }

   # Convert Slack message object into JSON
   $body = $payload | ConvertTo-Json -Depth 5

   if($env:function_debug -eq "true") {
         Write-Host "DEBUG: body=$body"
   }

   Write-Host "$(Get-Date) - Sending notification to Slack ..."
   $ProgressPreference = "SilentlyContinue"

   try {
      Invoke-WebRequest -Uri $(${jsonSecrets}.SLACK_WEBHOOK_URL) -Method POST -ContentType "application/json" -Body $body
   } catch {
      throw "$(Get-Date) - Failed to send notification to Slack: $($_)"
   }

   Write-Host "$(Get-Date) - Successfully sent notification to Slack ..."
}
