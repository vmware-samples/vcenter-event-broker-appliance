# Process function Secrets passed in
$VC_CONFIG_FILE = "/var/openfaas/secrets/vcconfig"
$VC_CONFIG = (Get-Content -Raw -Path $VC_CONFIG_FILE | ConvertFrom-Json)
if($env:function_debug -eq "true") {
    Write-host "DEBUG: `"$VC_CONFIG`""
}

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: `"$json`""
}

$eventObjectName = $json.objectName

Set-PowerCLIConfiguration -InvalidCertificateAction Ignore  -DisplayDeprecationWarnings $false -ParticipateInCeip $false -Confirm:$false | Out-Null

# Connect to vCenter Server
Write-Host "Connecting to vCenter Server ..."
Connect-VIServer -Server $($VC_CONFIG.VC) -User $($VC_CONFIG.VC_USERNAME) -Password $($VC_CONFIG.VC_PASSWORD)

# Retrieve VM and apply vSphere Tag
Write-Host "Applying vSphere Tag `"$($VC_CONFIG.TAG_NAME)`" to $eventObjectName ..."
Get-VM $eventObjectName | New-TagAssignment -Tag (Get-Tag -Name $($VC_CONFIG.TAG_NAME)) -Confirm:$false

Write-Host "Disconnecting from vCenter Server ..."
Disconnect-VIServer * -Confirm:$false
