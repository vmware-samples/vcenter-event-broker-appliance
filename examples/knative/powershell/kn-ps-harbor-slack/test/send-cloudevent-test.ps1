
$headers = @{
    "Content-Type" = "application/json";
    "ce-specversion" = "1.0";
    "ce-id" = "291ee129-1d27-415c-bbe1-3ca45d5f230a";
    "ce-source" = "/projects/2/webhook/policies/1";
    "ce-type" = "harbor.artifact.pushed";
    "ce-operator" = "admin";
    "ce-time" = "2023-08-22T15:56:41Z";
}

$body = Get-Content -Raw -Path "./test-payload.json"

Write-Host "Testing Function ..."
Invoke-WebRequest -Uri http://localhost:8080 -Method POST -Headers $headers -Body $body

Write-host "See docker container console for output"