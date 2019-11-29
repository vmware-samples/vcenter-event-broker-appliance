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
$managedObjectReference = $json.managedObjectReference

Set-PowerCLIConfiguration -InvalidCertificateAction Ignore  -DisplayDeprecationWarnings $false -ParticipateInCeip $false -Confirm:$false | Out-Null

# Connect to vCenter Server
Write-Host "Connecting to vCenter Server ..."
Connect-VIServer -Server $($VC_CONFIG.VC) -User $($VC_CONFIG.VC_USERNAME) -Password $($VC_CONFIG.VC_PASSWORD)

# Get the vCenter AlarmManager
$alarmManager = Get-View AlarmManager
if ($json.topic -eq "entered.maintenance.mode") {
    # Disable alarm actions on the host
    Write-Host "Disabling alarm actions on host: $eventObjectName"
    $alarmManager.EnableAlarmActions($managedObjectReference, $false)
}
else {
    # Enable alarm actions on the host
    Write-Host "Enabling alarm actions on host: $eventObjectName"
    $alarmManager.EnableAlarmActions($managedObjectReference, $true)
}

Write-Host "Disconnecting from vCenter Server ..."
Disconnect-VIServer * -Confirm:$false
