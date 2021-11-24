Import-Module CloudEvents.Sdk

. ./vrni-functions.ps1

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
      [Parameter(Position = 0, Mandatory = $true)]$Source,
      [Parameter(Position = 1, Mandatory = $true)]$Body
   )

   # Retrieve webhook sink URL, default to internal VEBA URL if it's not in ENV
   $webhookSinkUrl = "http://default-broker-ingress.vmware-functions.svc.cluster.local"
   if ($null -ne $env:WEBHOOK_SINK_URL) {
      $webhookSinkUrl = ${env:WEBHOOK_SINK_URL}
   }

   Write-host "$(Get-Date) - Creating CloudEvents...`n"
   try {
      $cloudEvents = New-vRNICloudEventsFromDatabus -Data $body -cloudEventSource $Source
   }
   catch {
      $ErrorMessage = $_.Exception.Message
      throw "`n$(Get-Date) - Unable to run New-vRNICloudEventsFromDatabus:`n$($ErrorMessage)"
   }

   if ($cloudEvents.Count -eq 0) {
      Write-Host "$(Get-Date) - No cloud events generated: not sending to VMware Event Broker"
      return
   }

   # only send to broker if SERVICE_TEST is not "true"
   if (${env:SERVICE_TEST} -ne "true") {
      $ProgressPreference = "SilentlyContinue"

      # run through all generated events and send them towards the event broker
      foreach ($cloudEvent in $cloudEvents) {
         try {
            $cloudEventBinaryHttpMessage = $cloudEvent | ConvertTo-HttpMessage -ContentMode Binary
         }
         catch {
            throw "`nFailed to convert CloudEvent into binary HTTP message"
         }

         if (${env:SERVICE_DEBUG} -eq "true") {
            # Only used for debugging and printing to console
            Write-Host "$(Get-Date) - [CloudEvent]:"
            Write-Host $($cloudEvent | out-string)
            Write-Host
         }

         Write-Host "$(Get-Date) - Sending CloudEvent to VMware Event Broker ...`n"
         try {
            Invoke-WebRequest -Method POST -Uri $webhookSinkUrl -Headers $cloudEventBinaryHttpMessage.Headers -Body $cloudEventBinaryHttpMessage.Body
         }
         catch {
            throw "`nFailed to forward CloudEvent to VMware Event Broker"
         }

         Write-Host "$(Get-Date) - Successfully forwarded CloudEvent ...`n"
      }
   }
   else {
      Write-Host "$(Get-Date) - Running in local development: not sending CloudEvent to VMware Event Broker"
   }
}
