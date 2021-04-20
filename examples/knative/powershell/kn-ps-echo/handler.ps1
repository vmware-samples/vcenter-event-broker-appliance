Function Process-Handler {
   param(
      [Parameter(Position=0,Mandatory=$true)][CloudNative.CloudEvents.CloudEvent]$CloudEvent
   )

   Write-Host "Cloud Event"
   Write-Host "  Source: $($cloudEvent.Source)"
   Write-Host "  Type: $($cloudEvent.Type)"
   Write-Host "  Subject: $($cloudEvent.Subject)"
   Write-Host "  Id: $($cloudEvent.Id)"

   # Decode CloudEvent
   $cloudEventData = $cloudEvent | Read-CloudEventJsonData -ErrorAction SilentlyContinue -Depth 10
   if($cloudEventData -eq $null) {
      $cloudEventData = $cloudEvent | Read-CloudEventData
   }

   Write-Host "CloudEvent Data:"
   Write-Host "$($cloudEventData | Out-String)"
}
