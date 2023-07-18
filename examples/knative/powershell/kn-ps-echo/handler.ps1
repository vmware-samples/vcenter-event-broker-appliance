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

   Write-Host $(Get-Date) - "Cloud Event"
   Write-Host $(Get-Date) - "  Source: $($cloudEvent.Source)"
   Write-Host $(Get-Date) - "  Type: $($cloudEvent.Type)"
   Write-Host $(Get-Date) - "  EventClass: $($cloudEvent.EventClass)"
   Write-Host $(Get-Date) - "  Id: $($cloudEvent.Id)"

   # Decode CloudEvent
   try {
      $cloudEventData = $cloudEvent | Read-CloudEventJsonData -Depth 10
   } catch {
      throw "`nPayload must be JSON encoded"
   }

   Write-Host $(Get-Date) - "CloudEvent Data:"
   Write-Host $(Get-Date) - "`n$($cloudEventData | ConvertTo-Json)"
}
