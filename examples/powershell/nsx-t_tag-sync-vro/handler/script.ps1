# Author: Craig Straka (craig.straka@it-partners.com)
# Version .1 (Beta)
# Master stored at: https://github.com/cstraka/NSX-T_Tag-Sync (/vro)
# Intention:
#   Synchronize vSphere tags with NSX-T tags unidiretionally (vSphere is the master)
#   Script is used as a OpenFaas VMware Event Broker powercli function to intercept and parse vSphere events from event types:
#       - com.vmware.cis.tagging.attach
#       - com.vmware.cis.tagging.detach
#   Script sends parsed results to a VRO instance for further data inut validation and to discover vSphere Tags and apply them to NSX.
# 
# Tested with:
#   VEBA .5 (OpenFaaS)
#   vSphere 7.0.1
#   NSX-T 3.0.2
#   VMware vRealize Orchestrator 8.2 - Standalone (no VRA)
#
# Notes:
#   Machine names in a vCenter instance must be unique: 
#       Script has no way, based on limited specificity of vSphere event message of the types above, to discern correct VM other than by name.
#           Lots of work to accomodate spaces, new lines, and other anomalies in the event data to get the VM name.
#               accomodation efforts to handle naming are ongoing as errata is reported.
#           
#       Script does NOT detect duplicate named VM's, VRO workflow does.  
#           Overridding intention is to not use 'Get-VM' cmdlet, and required modules, in this script which would be required to detect dupes.
#
# Planned enhancements (script today is minimally viable):
#   Updates based on vSphere event changes.
#   Send in NSX FQDN so vRO can dynamically determine the correct NSX-T Local Manager
#       Possibly part of the Secrets file, but further vSphere and NSX integration may alter the enhancement direction.
#   Hopefully, better vSphere and NSX-T integration will render this script obsolete.

# Process function Secrets passed in
$SECRETS_FILE = "/var/openfaas/secrets/vro-secrets"
$SECRETS_CONFIG = (Get-Content -Raw -Path $SECRETS_FILE | ConvertFrom-Json)

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

# Set vCenter server name to a variable from event message text
$vcenter = ($json.source -replace "https://","" -replace "/sdk","");

# Pull VM name from event message text and set it to variable.  
$separator = "object"
$FullFormattedMessage = $json.data.FullFormattedMessage
if($env:function_debug -eq "true") {
    write-host "FullFormattedMessage RAW="$FullFormattedMessage
}
$FullFormattedMessage.replace([Environment]::NewLine," ")
if($env:function_debug -eq "true") {
    write-host "FullFormattedMessage NewLine="$FullFormattedMessage
}
$pos = $FullFormattedMessage.IndexOf($separator)
$rightPart = $FullFormattedMessage.Substring($pos+1)
if($env:function_debug -eq "true") {
    $leftPart = $FullFormattedMessage.Substring(0, $pos)
    write-host "FullFormattedMessage leftPart="$leftPart
    write-host "FullFormattedMessage rightPart="$rightPart
}
$pos = $rightPart.replace("bject","")
$FormattedMessage = $pos.replace([Environment]::NewLine," ")
if($env:function_debug -eq "true") {
    write-host "FullFormattedMessage Split="$FullFormattedMessage
}
$FormattedMessage = $FormattedMessage.trim()
if($env:function_debug -eq "true") {
    write-host "FullFormattedMessage Complete="$FullFormattedMessage
}
$vm = $FormattedMessage

# Test for existince of content in $vm variable and exit script early if test results false
if($vm -eq "") {
    Write-Host "Unable to retrieve VM Object from Event payload, please ensure Event contains VM result"
    exit
}

# This syntax is very specific.  
# The 'name' element (e.g. "name": "virtualMachineName" & "name": "vcenterName") MUST be a same named input to the VRO workflow with a matching type (e.g. 'string')
# CASE SENSITIVE.
$vroBody = @"
{
    "parameters": [
        {
            "type": "string",
            "name": "virtualMachineName",
            "scope": "local",
            "value": {
                "string": {
                    "value": "$vm"
                }
            }
        },
        {
            "type": "string",
            "name": "vcenterName",
            "scope": "local",
            "value": {
                "string": {
                    "value": "$vcenter"
                }
            }
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

#Setting VRO URL
$vroUrl = "https://$($SECRETS_CONFIG.VRO_SERVER):443/vco/api/workflows/$($SECRETS_CONFIG.VRO_WORKFLOW_ID)/executions"

#writing variables to console if 'function_debug' is 'true' in stack.yml
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: VM=$vm"
    Write-Host "DEBUG: vRoURL=`"$($vroUrl | Format-List | Out-String)`""
    Write-Host "DEBUG: headers=`"$($headers | Format-List | Out-String)`""
    Write-Host "DEBUG: body=$vroBody"
}

#calling VRO
$response = ""
if($env:skip_vro_cert_check -eq "true") {
    $response = Invoke-Webrequest -Uri $vroUrl -Method POST -Body $vroBody -Headers $headers -SkipHeaderValidation -SkipCertificateCheck
} else {
    $response = Invoke-Webrequest -Uri $vroUrl -Method POST -Body $vroBody -Headers $headers -SkipHeaderValidation 
}
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: Invoke-WebRequest=$response"
}