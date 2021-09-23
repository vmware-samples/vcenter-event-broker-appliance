Import-Module CloudEvents.Sdk

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
      [Parameter(Position=0,Mandatory=$true)]$Headers,
      [Parameter(Position=1,Mandatory=$true)]$Body
   )

   try {
      $jsonSecrets = ${env:WEBHOOK_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:WEBHOOK_SECRET does not look to be defined"
   }

   Write-host "$(Get-Date) - Processing Body ...`n"
   try {
      if ($Body -is [string]) {
         $Body = [System.Text.Encoding]::UTF8.GetBytes($Body)
      }
   } catch {
      throw "`nFailed to encode `$Body input - `$Error[0]"
   }

   if(${env:SERVICE_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - [Function Secrets]:`n${env:WEBHOOK_SECRET}`n"
   }

   # Check to see if basic auth has been configured for webhook
   if(${jsonSecrets}.WEBHOOK_USERNAME -ne $NULL -and ${jsonSecrets}.WEBHOOK_PASSWORD -ne $NULL) {
      Write-host "$(Get-Date) - Checking for authorization ...`n"

      # Expected base64 value for the configured webhook username/pass
      $expectedPair = "$(${jsonSecrets}.WEBHOOK_USERNAME):$(${jsonSecrets}.WEBHOOK_PASSWORD)"
      $expectedBytes = [System.Text.Encoding]::ASCII.GetBytes($expectedPair)
      $expectedBase64 = "Basic " + [System.Convert]::ToBase64String($expectedBytes)

      if(${env:SERVICE_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - [Expected Authorization Header]:`n$expectedBase64`n"
      }

      # Check for the supplied webhook username/pass
      if(${Headers}.Authorization) {
         if($(${Headers}.Authorization) -ne "$expectedBase64") {
            Write-Error "$(Get-Date) - Invalid Authorization`n"
            $exception = New-Object Exception
            $httpException = New-Object System.Net.Http.HttpRequestException -ArgumentList "Invalid Authorization",$exception,401
            throw $httpException
         }
      } else {
         Write-Error "$(Get-Date) - Authorization header was not provided`n"
         $exception = New-Object Exception
         $httpException = New-Object System.Net.Http.HttpRequestException -ArgumentList "Authorization header was not provided",$exception,401
         throw $httpException
      }
   }

   # Retrieve webhook sink URL
   $webhookSinkUrl = $(${jsonSecrets}.WEBHOOK_SINK_URL)

   # Retrieve the CloudEvent Type
   $cloudEventType = $(${jsonSecrets}.WEBHOOK_CE_EVENT_TYPE)

   # Retrieve forwarded IP[0] (client address), if that is empty, set to function hostname
   try {
      $cloudEventSource = ((${headers}.'X-Forwarded-For') -split ',')[0]

      if($cloudEventSource -eq "") {
         $cloudEventSource = [System.Net.Dns]::GetHostName()
      }
   } catch {
      Write-Host "$(Get-Date) - Unable to get remote client, setting CloudEvent source to function hostname"
      throw "`nFailed to set source for CloudEvent"
   }


   Write-host "$(Get-Date) - Creating CloudEvent ...`n"
   try {
      $cloudEvent = New-CloudEvent -Type $cloudEventType -Source $cloudEventSource -Id (New-Guid).Guid -Time (Get-Date)
      $cloudEvent = $cloudEvent | Set-CloudEventData -DataContentType "application/json" -Data $body
   } catch {
      throw "`nFailed to construct CloudEvent"
   }

   if(${env:SERVICE_DEBUG} -eq "true") {
      # Only used for debugging and printing to console
      $webhookBody = [System.Text.Encoding]::UTF8.GetString($Body) | ConvertFrom-Json

      Write-Host "$(Get-Date) - [Headers]:"
      Write-Host $($Headers | Out-String)
      Write-Host "$(Get-Date) - [Body]:"
      Write-Host
      Write-Host $($webhookBody)
      Write-Host
      Write-Host "$(Get-Date) - [CloudEvent]:"
      Write-Host $($cloudEvent|out-string)
   }

   # Disable forwarding to broker if $webhookSinkUrl is null or ""
   if($webhookSinkUrl -ne $null -and $webhookSinkUrl -ne "") {
      $ProgressPreference = "SilentlyContinue"

      try {
         $cloudEventBinaryHttpMessage = $cloudEvent | ConvertTo-HttpMessage -ContentMode Binary
      } catch {
         throw "`nFailed to convert CloudEvent into binary HTTP message"
      }

      Write-Host "$(Get-Date) - Sending CloudEvent to VMware Event Broker ...`n"
      try {
         Invoke-WebRequest -Method POST -Uri $webhookSinkUrl -Headers $cloudEventBinaryHttpMessage.Headers -Body $cloudEventBinaryHttpMessage.Body
      } catch {
         throw "`nFailed to send CloudEvent to VMware Event Broker"
      }

      Write-Host "$(Get-Date) - Successfully sent CloudEvent ...`n"
   } else {
      Write-Host "$(Get-Date) - Running in local development: not sending CloudEvent to VMware Event Broker"
   }
}
