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

# import and configure Slack
Import-Module PSSlack | Out-Null


Set-PowerCLIConfiguration -InvalidCertificateAction Ignore  -DisplayDeprecationWarnings $false -ParticipateInCeip $false -Confirm:$false | Out-Null

# Connect to vCenter Server
Write-Host "Connecting to vCenter Server ..."
Connect-VIServer -Server $($VC_CONFIG.VC) -User $($VC_CONFIG.VC_USERNAME) -Password $($VC_CONFIG.VC_PASSWORD)

# Retrieve VM changes
$Message = (Get-VM $eventObjectName | Get-ViEvent -MaxSamples 1).FullFormattedMessage

# Bold format for titles
[string]$Message = $Message -replace "Modified","*Modified*" -replace "Added","*Added*" -replace "Deleted","*Deleted*"

# Send VM changes
Write-Host "Detected change to $eventObjectName ..."

New-SlackMessageAttachment -Color $([System.Drawing.Color]::red) `
                           -Title 'VM Change detected' `
                           -Text "$Message" `
                           -Fallback 'ouch' |
    New-SlackMessage -Channel $($VC_CONFIG.SLACK_CHANNEL) `
                     -IconEmoji :fire: |
    Send-SlackMessage -Uri $($VC_CONFIG.SLACK_URL)


Write-Host "Disconnecting from vCenter Server ..."
Disconnect-VIServer * -Confirm:$false
