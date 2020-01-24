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
$vmToTag = Get-VM $eventObjectName
$searchTag = $NULL
$lenTagFilter = $($VC_CONFIG.TAG_FILTER).length
if ($lenTagFilter -gt 0)  # If the TAG_FILTER property is populated, we only want to add TAG_NAME to VMs that are already tagged with TAG_FILTER
{
	$searchTag = $vmToTag | Get-TagAssignment | Where { $_.Tag.Name -eq $($VC_CONFIG.TAG_FILTER) }
	if ($searchTag -eq $NULL)
	{
		Write-Host "TAG_FILTER"$($VC_CONFIG.TAG_FILTER)"not found on VM"$eventObjectName
	}
}

if ($lenTagFilter -eq 0 -Or $searchTag -ne $NULL) #If TAG_FILTER is empty, add TAG_NAME to all VMs. If TAG_FILTER is populated and the VM is tagged with the filter, add TAG_NAME
{
	Write-Host "Applying vSphere Tag `"$($VC_CONFIG.TAG_NAME)`" to $eventObjectName ..."
	$vmToTag | New-TagAssignment -Tag (Get-Tag -Name $($VC_CONFIG.TAG_NAME)) -Confirm:$false
}

Write-Host "Disconnecting from vCenter Server ..."
Disconnect-VIServer * -Confirm:$false
