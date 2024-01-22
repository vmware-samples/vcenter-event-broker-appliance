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
      $jsonSecrets = ${env:GOOGLE_CHAT_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:GOOGLE_CHAT_SECRET does not look to be defined"
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: K8s Secrets:`n${env:GOOGLE_CHAT_SECRET}`n"

      Write-Host "$(Get-Date) - DEBUG: CloudEventData`n $(${cloudEventData} | Out-String)`n"
   }

   $vc = (${cloudEvent}.Source).toString().replace("/sdk","")
   $dateTime = ${cloudEvent}.time

   $payload = @{
      "text" = "`VCSA Backup Failure Alert` - `\n\t*DateTime*: ${dateTime}`\n\t*VCSA*: ${vc}/ui`\n\t*VCSA VAMI*: ${vc}:5480"
   }

   # Convert payload to JSON and un-escape "\n" which is used to annotate line break for Google Chat messages
   $body = ($payload | ConvertTo-Json -Depth 2).Replace('\\n','\n').Replace('\\t','\t')

   $headers = @{
      "Content-Type" = "application/json";
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: `"$($headers | Out-String)`""
      Write-Host "$(Get-Date) - DEBUG: `"$body`""
   }

   Write-Host "$(Get-Date) - Sending message to Google Chat Webhook ..."
   $ProgressPreference = "SilentlyContinue"

   try {
      Invoke-WebRequest -Uri $(${jsonSecrets}.GOOGLE_CHAT_WEBHOOK_URL) -Method POST -Headers $headers -Body $body
   } catch {
      throw "$(Get-Date) - Failed to send Google Chat message: $($_)"
   }

   Write-Host "$(Get-Date) - Successfully sent Google Chat message ..."
}
