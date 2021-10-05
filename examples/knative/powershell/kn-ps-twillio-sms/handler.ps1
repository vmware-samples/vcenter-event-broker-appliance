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
      $jsonSecrets = ${env:SMS_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:SMS_SECRET does not look to be defined"
   }

   # Check TaskEvent to make sure its VM Create Snapshot Event
   # See https://williamlam.com/2019/02/creating-vcenter-alarms-based-on-task-events-such-as-folder-creation.html for more details on finding descriptionId from TaskEvent
   if($cloudEventData.Info.DescriptionId -eq "VirtualMachine.createSnapshot") {
      if(${env:FUNCTION_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - [Function Secrets]:`n${env:SMS_SECRET}`n"

         Write-Host "$(Get-Date) - [CloudEventData]:`n $(${cloudEventData} | Out-String)`n"
      }

      # Construct Custom SMS Message
      $smsMessage = @"
Hi,

Snapshot for VM: $($cloudEventData.Vm.Name) was just created by $($cloudEventData.UserName)

-From VEBA
"@

      # The following was adapted from https://www.twilio.com/docs/usage/tutorials/how-to-make-http-basic-request-twilio-powershell

      $TWILLO_BASE_API_URL_ENDPOINT = ${jsonSecrets}.TWILLO_BASE_API_URL_ENDPOINT
      $TWILLO_SID = ${jsonSecrets}.TWILLO_SID
      $TWILLO_AUTH_TOKEN = ${jsonSecrets}.TWILLO_AUTH_TOKEN
      $TWILLO_NUMBER = ${jsonSecrets}.TWILLO_NUMBER
      $SMS_DESTINATION_NUMBER = ${jsonSecrets}.SMS_DESTINATION_NUMBER

      # Twilio API endpoint and POST params
      $url = "${TWILLO_BASE_API_URL_ENDPOINT}/Accounts/${TWILLO_SID}/Messages.json"
      $params = @{ To = ${SMS_DESTINATION_NUMBER}; From = ${TWILLO_NUMBER}; Body = $smsMessage }

      # Create a credential object for HTTP basic auth
      $p = ${TWILLO_AUTH_TOKEN} | ConvertTo-SecureString -asPlainText -Force
      $credential = New-Object System.Management.Automation.PSCredential($TWILLO_SID, $p)

      $ProgressPreference = "SilentlyContinue"

      # Make API request, selecting JSON properties from response
      Write-Host "$(Get-Date) - Sending SMS Message"
      try {
         $response = Invoke-WebRequest $url -Method Post -Credential $credential -Body $params
         $StatusCode = $Response.StatusCode
      } catch {
         throw "$(Get-Date) - Failed to send SMS: $($_)"
      }

      if($StatusCode -eq "201") {
         Write-Host "$(Get-Date) - Successfully sent SMS message"
      } else {
         Write-Host "$(Get-Date) - Failed to send SMS message with status code: $StatusCode"
      }
   } else {
      Write-Host "$(Get-Date) - Skipped TaskEvent type $($cloudEventData.Info.DescriptionId)"
   }
}
