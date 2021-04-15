
$headers = @{
    "Content-Type" = "application/json";
    "ce-specversion" = "1.0";
    "ce-id" = "id-123";
    "ce-source" = "source-123";
    "ce-type" = "binary";
    "ce-subject" = "subject-123";
}

$body = Get-Content -Raw -Path "./binary-payload"

Write-Host "Testing Function ..."
Invoke-WebRequest -Uri http://localhost:8080 -Method POST -Headers $headers -Body $body

Write-host "See docker container console for output"