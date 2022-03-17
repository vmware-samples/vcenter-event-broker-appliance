# The ce-subject value should match the event router subject in function.yaml
$headers = @{
    "Content-Type" = "application/json";
    "ce-specversion" = "1.0";
    "ce-id" = "id-123";
    "ce-source" = "source-123";
    "ce-type" = "com.vmware.event.router/event";
    "ce-subject" = "VmReconfiguredEvent";
}

$payloadPath = "./test-payload.json"
if ( $args.Count -gt 0 ) {
    if ( Test-Path $args[0] ) {
        $payloadPath = $args[0]
    }
    else {
        Write-Host "$(Get-Date) - ERROR: Invalid path"$args[0]"`n"
        exit
    }
}
$body = Get-Content -Raw -Path $payloadPath

Write-Host "Testing Function ..."
Invoke-WebRequest -Uri http://localhost:8080 -Method POST -Headers $headers -Body $body

Write-host "See docker container console for output"