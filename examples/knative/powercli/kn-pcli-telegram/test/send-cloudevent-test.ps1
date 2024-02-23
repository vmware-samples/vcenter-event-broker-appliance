
$headers = @{
    "Content-Type" = "application/json";
    "ce-specversion" = "1.0";
    "ce-id" = "41289fef-0727-46f7-b1a9-b8145972c734";
    "ce-source" = "https://vcenter.local/sdk";
    "ce-type" = "com.vmware.vsphere.VmMigratedEvent.v0";
}

$body = Get-Content -Raw -Path "./test-payload.json"

Write-Host "Testing Function ..."
Invoke-WebRequest -Uri http://localhost:8080 -Method POST -Headers $headers -Body $body

Write-host "See docker container console for output"