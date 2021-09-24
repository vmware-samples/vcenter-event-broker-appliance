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
      $jsonSecrets = ${env:TEAMS_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:TEAMS_SECRET does not look to be defined"
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: K8s Secrets:`n${env:TEAMS_SECRET}`n"

      Write-Host "$(Get-Date) - DEBUG: CloudEventData`n $(${cloudEventData} | Out-String)`n"
   }

   # Construct Teams message object
   $payload = [PSCustomObject][Ordered]@{
      "@type"      = "MessageCard"
      "@context"   = "http://schema.org/extensions"
      "themeColor" = '0078D7'
      "summary"      = "New VMC SDDC Provisioned"
      "sections"   = @(
         @{
            "activityTitle" = "&#x1F973; **New VMC SDDC Provisioned** &#x1F973;";
            "activitySubtitle" = "In $($cloudEventData.org_name) Organization";
            "activityImage" = "https://blogs.vmware.com/vsphere/files/2019/07/Icon-2019-VMWonAWS-Primary-354-x-256.png";
            "facts" = @(
               @{
                     "name" = "SDDC:";
                     "value" = $cloudEventData.resource_name;
               },
               @{
                     "name" = "User:";
                     "value" = $cloudEventData.message_username;
               },
               @{
                     "name" = "URL:";
                     "value" = "[Click to open SDDC]($($cloudEvent.source))";
               }
            );
            "markdown" = $true
         }
      )
   }

   # Convert Teams message object into JSON
   $body = $payload | ConvertTo-Json -Depth 5

   if($env:function_debug -eq "true") {
         Write-Host "DEBUG: body=$body"
   }

   Write-Host "$(Get-Date) - Sending notification to Microsoft Teams ..."
   $ProgressPreference = "SilentlyContinue"

   try {
      Invoke-WebRequest -Uri $(${jsonSecrets}.TEAMS_WEBHOOK_URL) -Method POST -ContentType "application/json" -Body $body
   } catch {
      throw "$(Get-Date) - Failed to send notification to Microsoft Teams: $($_)"
   }

   Write-Host "$(Get-Date) - Successfully sent notification to Microsoft Teams ..."
}
