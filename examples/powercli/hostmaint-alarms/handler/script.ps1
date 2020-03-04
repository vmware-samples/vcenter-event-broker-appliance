# Process function Secrets passed in
$VC_CONFIG_FILE = "/var/openfaas/secrets/vc-hostmaint-config"
$VC_CONFIG = (Get-Content -Raw -Path $VC_CONFIG_FILE | ConvertFrom-Json)
if($env:function_debug -eq "true") {
    Write-host "DEBUG: `"$VC_CONFIG`""
}

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

$eventObjectName = $json.data.host.name

Set-PowerCLIConfiguration -InvalidCertificateAction Ignore  -DisplayDeprecationWarnings $false -ParticipateInCeip $false -Confirm:$false | Out-Null

# Connect to vCenter Server
Write-Host "Connecting to vCenter Server ..."
Connect-VIServer -Server $($VC_CONFIG.VC) -User $($VC_CONFIG.VC_USERNAME) -Password $($VC_CONFIG.VC_PASSWORD)

# Construct MoRef from Type/Value
$moRef = New-Object VMware.Vim.ManagedObjectReference
$moRef.Type = $json.data.host.host.type
$moRef.Value = $json.data.host.host.Value
$hostMoRef = Get-View $moRef

# Get the vCenter AlarmManager
$alarmManager = Get-View AlarmManager

if ($json.subject -eq "EnteredMaintenanceModeEvent") {
    # Disable alarm actions on the host
    Write-Host "Disabling alarm actions on host: $eventObjectName"
    $alarmManager.EnableAlarmActions($hostMoRef.MoRef, $false)
}

if ($json.subject -eq "ExitMaintenanceModeEvent") {
    # Enable alarm actions on the host
    Write-Host "Enabling alarm actions on host: $eventObjectName"
    $alarmManager.EnableAlarmActions($hostMoRef.MoRef, $true)
}

Write-Host "Disconnecting from vCenter Server ..."
Disconnect-VIServer * -Confirm:$false
