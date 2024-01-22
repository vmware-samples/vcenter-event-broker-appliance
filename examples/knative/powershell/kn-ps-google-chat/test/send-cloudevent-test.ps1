
$headers = @{
    "Content-Type" = "application/json";
    "ce-specversion" = "1.0";
    "ce-id" = "2112913";
    "ce-source" = "https://vcenter.primp-industries.local/sdk";
    "ce-type" = "com.vmware.vsphere.com.vmware.applmgmt.backup.job.failed.event.v0";
}

$body = Get-Content -Raw -Path "./test-payload.json"

Write-Host "Testing Function ..."
Invoke-WebRequest -Uri http://localhost:8080 -Method POST -Headers $headers -Body $body

Write-host "See docker container console for output"