
# Process function Secrets passed in
$SECRETS_FILE = "/var/openfaas/secrets/vro-secrets"
$SECRETS_CONFIG = (Get-Content -Raw -Path $SECRETS_FILE | ConvertFrom-Json)

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

$vcenter = ($json.source -replace "https://","" -replace "/sdk","");
$vmMoRef = $json.data.vm.vm.value;
$vm = $json.data.vm.name;

if($vmMoRef -eq "" -or $vm -eq "") {
    Write-Host "Unable to retrieve VM Object from Event payload, please ensure Event contains VM result"
    exit
}

# e.g. mgmt-vcsa-01.cpbu.corp/vm-2660
$vroVmId = "$vcenter/$vmMoRef"

# assumes following vRO Workflow https://github.com/kclinden/vro-vsphere-tagging as an example
$body = @"
{
    "parameters":
	[
        {
            "value": {
                "sdk-object":{
                    "type": "VC:VirtualMachine",
                    "id": "$($vroVmId)"}
                },
            "type": "VC:VirtualMachine",
            "name": "vm",
            "scope": "local"
        },
        {
            "value": {
                "string":{
                    "value": "$($SECRETS_CONFIG.TAG_CATEGORY_NAME)"
                }
            },
            "type": "string",
            "name": "categoryName",
            "scope": "local"
        },
        {
            "value": {
                "string":{
                    "value": "$($SECRETS_CONFIG.TAG_NAME)"
                }
            },
            "type": "string",
            "name": "tagToFind",
            "scope": "local"
        }
	]
}
"@

# Basic Auth for vRO execution
$pair = "$($SECRETS_CONFIG.VRO_USERNAME):$($SECRETS_CONFIG.VRO_PASSWORD)"
$bytes = [System.Text.Encoding]::ASCII.GetBytes($pair)
$base64 = [System.Convert]::ToBase64String($bytes)
$basicAuthValue = "Basic $base64"

$headers = @{
    "Authorization"="$basicAuthValue";
    "Accept="="application/json";
    "Content-Type"="application/json";
}

$vroUrl = "https://$($SECRETS_CONFIG.VRO_SERVER):443/vco/api/workflows/$($SECRETS_CONFIG.VRO_WORKFLOW_ID)/executions"

if($env:function_debug -eq "true") {
    Write-Host "DEBUG: vRoVmID=$vroVmId"
    Write-Host "DEBUG: TagCategory=$($SECRETS_CONFIG.TAG_CATEGORY_NAME)"
    Write-Host "DEBUG: TagName=$($SECRETS_CONFIG.TAG_NAME)"
    Write-Host "DEBUG: vRoURL=`"$($vroUrl | Format-List | Out-String)`""
    Write-Host "DEBUG: headers=`"$($headers | Format-List | Out-String)`""
    Write-Host "DEBUG: body=$body"
}

Write-Host "Applying vSphere Tag: $($SECRETS_CONFIG.TAG_NAME) to VM: $vm ..."
if($env:skip_vro_cert_check -eq "true") {
    Invoke-Webrequest -Uri $vroUrl -Method POST -Headers $headers -SkipHeaderValidation -Body $body -SkipCertificateCheck
} else {
    Invoke-Webrequest -Uri $vroUrl -Method POST -Headers $headers -SkipHeaderValidation -Body $body
}
