
$headers = @{
    "Content-Type" = "application/json";
    "ce-specversion" = "1.0";
    "ce-id" = "df8cbd23-01f0-4003-9a6b-d9f16a59f6be";
    "ce-source" = "https://vmc.vmware.com/console/sddcs/b8f349e8-48f1-4517-99fe-0bddc753e899";
    "ce-type" = "vmware.vmc.SDDC-PROVISION.v0";
}

$body = Get-Content -Raw -Path "./test-payload.json"

Write-Host "Testing Function ..."
Invoke-WebRequest -Uri http://localhost:8080 -Method POST -Headers $headers -Body $body

Write-host "See docker container console for output"