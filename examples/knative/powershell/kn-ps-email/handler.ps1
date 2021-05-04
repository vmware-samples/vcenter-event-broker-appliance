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
      Write-Host "DEBUG: K8s Secrets:`n${env:EMAIL_SECRET}`n"

      Write-Host "DEBUG: CloudEventData`n $(${cloudEventData} | ConvertTo-Json)`n"
   }

   if(${env:EMAIL_SECRET}) {
      $jsonSecrets = ${env:EMAIL_SECRET} | ConvertFrom-Json
   } else {
      Write-Host "K8s secrets `$env:EMAIL_SECRET does not look to be defined"
      break
   }

   ### BEGIN BUSINESS LOGIC CODE ###

   Import-Module Send-MailKitMessage

   # Extract all Email secrets for ease of use in function
   $EMAIL_SERVER=${jsonSecrets}.SMTP_SERVER
   $EMAIL_SERVER_PORT=${jsonSecrets}.SMTP_PORT
   $EMAIL_SERVER_USERNAME=${jsonSecrets}.SMTP_USERNAME
   $EMAIL_SERVER_PASSWORD=${jsonSecrets}.SMTP_PASSWORD
   $EMAIL_SUBJECT=${jsonSecrets}.EMAIL_SUBJECT
   $EMAIL_TO=${jsonSecrets}.EMAIL_TO
   $EMAIL_FROM=${jsonSecrets}.EMAIL_FROM

   # Extract VM Deleted Info from event for inclusion in email
   $VmDeletedName = $cloudEventData.Vm.Name
   $VmDeletedByUser = $cloudEventData.UserName
   $VmDeletedTime = $cloudEventData.CreatedTime

   # Create Email Body
   $EmailBody = "Virtual Machine ${VmDeletedName} was deleted by ${VmDeletedByUser} on ${VmDeletedTime}"

   if(${env:FUNCTION_DEBUG} -eq "true") {
      $debugOutput = @"
      EMAIL_SERVER=$EMAIL_SERVER
      EMAIL_SERVER_PORT=$EMAIL_SERVER_PORT
      EMAIL_SERVER_USERNAME=$EMAIL_SERVER_USERNAME
      EMAIL_SERVER_PASSWORD=$EMAIL_SERVER_PASSWORD
      EMAIL_SUBJECT=$EMAIL_SUBJECT
      EMAIL_TO=$EMAIL_TO
      EMAIL_FROM=$EMAIL_FROM
      EmailBody=$EmailBody
"@

      Write-Host "DEBUG: `n$debugOutput"
   }

   # Secure Email
   if($EMAIL_SERVER_USERNAME.length -gt 0 -and $EMAIL_SERVER_PASSWORD.length -gt 0) {
      $SecurePasswordString = ConvertTo-SecureString "$EMAIL_SERVER_PASSWORD" -AsPlainText -Force
      $Credential = New-Object System.Management.Automation.PSCredential($EMAIL_SERVER_USERNAME, $SecurePasswordString)

      $EmailParams = @{
         "RecipientList" = $EMAIL_TO
         "From"          = $EMAIL_FROM
         "Subject"       = $EMAIL_SUBJECT
         "TextBody"      = $EmailBody
         "SmtpServer"    = $EMAIL_SERVER
         "Credential"    = $Credential
         "Port"          = $EMAIL_SERVER_PORT
      }

      Write-Host "Sending Secure Email ..."
      Send-MailkitMessage @EmailParams

   } else {
   # Non-Secure Email
      $EmailParams = @{
         "RecipientList" = $EMAIL_TO
         "From"          = $EMAIL_FROM
         "Subject"       = $EMAIL_SUBJECT
         "TextBody"      = $EmailBody
         "SmtpServer"    = $EMAIL_SERVER
         "Port"          = $EMAIL_SERVER_PORT
   }

      Write-Host "Sending Non-Secure Email ..."
      Send-MailkitMessage @EmailParams
   }

   ### END BUSINESS LOGIC CODE ###
}
