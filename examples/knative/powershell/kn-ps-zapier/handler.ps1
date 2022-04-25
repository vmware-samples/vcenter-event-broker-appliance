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
      $jsonSecrets = ${env:ZAPIER_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:ZAPIER_SECRET does not look to be defined"
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: K8s Secrets:`n${env:ZAPIER_SECRET}`n"

      Write-Host "$(Get-Date) - DEBUG: CloudEventData`n $(${cloudEventData} | Out-String)`n"
   }

   # Arguments is returned as an array of Key/Value objects
   $arguments = ${cloudEventData}.Arguments
   foreach ($argument in $arguments) {
      if($argument.Key -eq "userIp") {
         $userIp = $argument.Value
         break
      }
   }

   # Construct Zapier payload
   $payload = @{
      Username = $cloudEventData.UserName
      UserIP = $userIp
      TimeStamp = $cloudEventData.CreatedTime
   }

   # Convert Zapier payload into JSON
   $body = $payload | ConvertTo-Json -Depth 5

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: `"$body`""
   }

   Write-Host "$(Get-Date) - Sending Webhook payload to Zapier ..."
   $ProgressPreference = "SilentlyContinue"

   try {
      Invoke-WebRequest -Uri $(${jsonSecrets}.ZAPIER_WEBHOOK_URL) -Method POST -ContentType "application/json" -Body $body
   } catch {
      throw "$(Get-Date) - Failed to send Zapier Message: $($_)"
   }

   Write-Host "$(Get-Date) - Successfully sent Webhook ..."
}
