# Process function Secrets passed in
$VC_CONFIG_FILE = "/var/openfaas/secrets/vc-sdrs-config"
$VC_CONFIG = (Get-Content -Raw -Path $VC_CONFIG_FILE | ConvertFrom-Json)
if($env:function_debug -eq "true") {
    Write-host "DEBUG: `"$VC_CONFIG`""
}

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

$vcenter = ($json.source -replace "/sdk","")
$datacenter = $json.data.datacenter.name
$DatastoreCluster = $json.data.objectname

$Body = "A new storage DRS recommendation has been generated `r`n"

$Body = $Body + @"
    vCenter Server: $vcenter
    Datacenter: $datacenter
    Datastore cluster: $DatastoreCluster	
"@


$bodytemp = @{
    roomId=$($VC_CONFIG.WEBEX_ROOM_ID)
    markdown=$Body
}
$json = $bodytemp | ConvertTo-Json
Invoke-RestMethod -Method Post `
    -Headers @{"Authorization"="Bearer $($VC_CONFIG.CISCO_BOT_TOKEN)"} `
    -ContentType "application/json" -Body $json `
    -Uri "https://api.ciscospark.com/v1/messages"


