Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   try {
      $jsonSecrets = ${env:TELEGRAM_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:TELEGRAM_SECRET does not look to be defined"
   }

   # Extract all telegram secrets for ease of use in function
   $VCENTER_SERVER = ${jsonSecrets}.VCENTER_SERVER
   $VCENTER_USERNAME = ${jsonSecrets}.VCENTER_USERNAME
   $VCENTER_PASSWORD = ${jsonSecrets}.VCENTER_PASSWORD
   $VCENTER_CERTIFICATE_ACTION = ${jsonSecrets}.VCENTER_CERTIFICATE_ACTION

   # Configure TLS 1.2/1.3 support as this is required for latest vSphere release
   [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor [System.Net.SecurityProtocolType]::Tls12 -bor [System.Net.SecurityProtocolType]::Tls13

   Write-Host "$(Get-Date) - Configuring PowerCLI Configuration Settings`n"
   Set-PowerCLIConfiguration -InvalidCertificateAction:${VCENTER_CERTIFICATE_ACTION} -ParticipateInCeip:$true -Confirm:$false

   Write-Host "$(Get-Date) - Connecting to vCenter Server $VCENTER_SERVER`n"

   try {
      Connect-VIServer -Server $VCENTER_SERVER -User $VCENTER_USERNAME -Password $VCENTER_PASSWORD
   } catch {
      Write-Error "$(Get-Date) - Failed to connect to vCenter Server"
      throw $_
   }

   Write-Host "$(Get-Date) - Successfully connected to $VCENTER_SERVER`n"

   Write-Host "$(Get-Date) - Init Processing Completed`n"
}

Function Process-Shutdown {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Shutdown`n"

   Write-Host "$(Get-Date) - Disconnecting from vCenter Server`n"

   try {
      Disconnect-VIServer * -Confirm:$false
   } catch {
      Write-Error "$(Get-Date) - Failed to Disconnect from vCenter Server"
   }

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
      $jsonSecrets = ${env:TELEGRAM_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:TELEGRAM_SECRET does not look to be defined"
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: K8s Secrets:`n${env:TELEGRAM_SECRET}`n"

      Write-Host "$(Get-Date) - DEBUG: CloudEventData`n $(${cloudEventData} | Out-String)`n"
   }

   $sendTelegram = $true

   # Ignore ResourcePool filter if not provided
   if(${jsonSecrets}.VCENTER_RESOURCE_POOL_FILTER -ne $NULL) {

      try {
         Write-Host "$(Get-Date) - Creating VM MoReF"

         # Construct VM Object given the MoRef ID from CloudEvent
         $moRef = New-Object VMware.Vim.ManagedObjectReference
         $moRef.Type = "VirtualMachine"
         $moRef.Value = ${cloudEventData}.Vm.Vm.Value

         $vm = Get-View $moRef
      } catch {
         throw "`Unable to construct VM Object using MoRef ID: $(${cloudEventData}.Vm.Vm.Value)"
      }

      # Retrieve the parent ResourcePool and only request the Name property
      try {
         Write-Host "$(Get-Date) - Retreiving parent Resource Pool"
         $rp = Get-View $vm.ResourcePool -Property Name
      } catch {
         throw "`Unable to retrieve parent Resource Pool"
      }

      # Check whether RP name matches filter, do not send telegram message if it does not match
      if($rp.name -ne ${jsonSecrets}.VCENTER_RESOURCE_POOL_FILTER) {
         $sendTelegram = $false
      }

   }

   $telegramMessage = "$(${cloudEventData}.Vm.Name) has been successfully migrated at $(${cloudEventData}.CreatedTime)"

   if($sendTelegram) {
      Write-Host "$(Get-Date) - Sending message to Telegram ..."
      $ProgressPreference = "SilentlyContinue"

      $telegramUrl = "https://api.telegram.org/bot$(${jsonSecrets}.TELEGRAM_BOT_API_KEY)/sendMessage"

      if(${env:FUNCTION_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - DEBUG: Telegram URL:`n`t${telegramUrl}`n"

         Write-Host "$(Get-Date) - DEBUG: Telegram Message:`n`t${telegramMessage}`n"
      }

      try {
         Invoke-WebRequest -Method POST -Uri $telegramUrl -ContentType "application/json;charset=utf-8" -Body (ConvertTo-Json -Compress -InputObject @{chat_id="$(${jsonSecrets}.TELEGRAM_GROUP_CHAT_ID)";text=$telegramMessage})
      } catch {
         throw "$(Get-Date) - Failed to send SMS: $($_)"
      }

      Write-Host "$(Get-Date) - Successfully sent Telegram message ..."
   }
}
