# Process function Secrets passed in
$SECRETS_FILE = "/var/openfaas/secrets/eventbridge-secrets"
$SECRETS_CONFIG = (Get-Content -Raw -Path $SECRETS_FILE | ConvertFrom-Json)

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

Import-Module AWS.Tools.EventBridge

$details = [pscustomobject] @{
    CreatedTime = $json.CreatedTime;
    UserName = $json.UserName;
    VMName = $json.objectName;
}

$data = ($details | convertTo-Json).toString()

$payload = New-Object Amazon.EventBridge.Model.PutEventsRequestEntry
$payload.EventBusName = $SECRETS_CONFIG.AWS_EVENTBRIDGE_BUS
$payload.Source = $json.source
$payload.Detail = $data
$payload.DetailType = $json.topic

if($env:function_debug -eq "true") {
    Write-Host "DEBUG: payload=`"$($payload | Format-List | Out-String)`""
}

Write-Host "Publishing custom event to EventBridge Bus ..."
Write-EVBEvent -Entry @($payload) -AccessKey $SECRETS_CONFIG.AWS_ACCESS_KEY -SecretKey $SECRETS_CONFIG.AWS_SECRET_KEY -Region $SECRETS_CONFIG.AWS_REGION
