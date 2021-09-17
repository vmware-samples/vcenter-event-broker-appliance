
$headers = @{
    "Content-Type" = "application/json";
    "ce-specversion" = "1.0";
    "ce-id" = "166419";
    "ce-source" = "https://hz-01.vmware.corp";
    "ce-type" = "com.vmware.event.router/horizon.vlsi_userlogin_rest_failed.v0";
    "ce-time" = "2021-09-03T16:00:28Z";
}

$body = Get-Content -Raw -Path "./test-payload.json"

Write-Host "Testing Function ..."
Invoke-WebRequest -Uri http://localhost:8080 -Method POST -Headers $headers -Body $body

Write-host "See docker container console for output"