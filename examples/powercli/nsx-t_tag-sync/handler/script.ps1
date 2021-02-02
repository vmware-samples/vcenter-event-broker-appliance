# Author: Craig Straka (craig.straka@it-partners.com)
# Version .1 (Beta)
# Master stored at: https://github.com/cstraka/NSX-T_Tag-Sync (/nsx)
# Intention:
#   Synchronize vSphere tags with NSX-T tags unidiretionally (vSphere is the master)
#   Script is used as a OpenFaas VMware Event Broker powercli function to intercept and parse vSphere events from event types:
#       - com.vmware.cis.tagging.attach
#       - com.vmware.cis.tagging.detach
#   Script gets vSphere Tags from vSphere, transforms the data, and updates the Virtual machine object in NSX-T.
# 
# Tested with:
#   VEBA .5 (OpenFaaS)
#   vSphere 7.0.1
#   NSX-T 3.0.2
#
# Notes:
#   Machine names in a vCenter instance must be unique: 
#       Script has no way, based on limited specificity of vSphere event message of the types above, to discern correct machine other than by name.
#           Lots of issues here as the name may not be unique in the cluster
#           Intention is to fix this as vSphere event messages evolve.
#       Event data has a number of odd charachters, such as new lines, that must be accomodated to get the machine name.  
#           accomodation efforts to handle naming are ongoing as errata is reported.
#
# Planned enhancements (script today is minimally viable):
#   Updates based on vSphere event changes
#   Hopefully, better vSphere and NSX-T integration will render this script obsolete.

# Process function Secrets passed in
$SECRETS_FILE = "/var/openfaas/secrets/nsx-secrets"
$SECRETS_CONFIG = (Get-Content -Raw -Path $SECRETS_FILE | ConvertFrom-Json)

# Test if PowerCLI Module is installed, install if not
if(Get-Module -ListAvailable -Name VMware.VimAutomation.Core) {
    Write-Host "Module exists"
} else {
    Write-Host "Module does not exist"
    Install-Package -Name VMware.VimAutomation.Core
}
Set-PowerCLIConfiguration -InvalidCertificateAction Ignore  -DisplayDeprecationWarnings $false -ParticipateInCeip $false -Confirm:$false | Out-Null

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json

if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

# Process payload sent from vCenter Server Event
$vcenter = ($json.source -replace "https://","" -replace "/sdk","")

# Pull VM name from event message text and set it to variable.  
$separator = "object"
$FullFormattedMessage = $json.data.FullFormattedMessage
if($env:function_debug -eq "true") {
    write-host "FullFormattedMessage RAW="$FullFormattedMessage
}
$FullFormattedMessage.replace([Environment]::NewLine," ")
if($env:function_debug -eq "true") {
    write-host "FullFormattedMessage minus NewLine="$FullFormattedMessage
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

if($vmMoRef -eq "" -or $vm -eq "") {
    Write-Host "Unable to retrieve VM Object from Event payload, please ensure Event contains VM result"
    exit
}

#Assigning credentials securely
$userName = $SECRETS_CONFIG.vCenter_USERNAME
$password = convertto-securestring $SECRETS_CONFIG.vCenter_PASSWORD -AsPlainText -Force
$Credentials = New-Object System.Management.Automation.PSCredential $userName,$password

#connecting to VI server
Write-Host "Connecting to VI Server..."
Connect-VIServer -Server $vcenter -Protocol https -Credential $credentials

# Get VM object from vCenter
$vm = Get-VM -name $vm | Select-Object Name,PersistentId

# until uniquely identifiable VM data is provided in a vSphere event this is the only option to maintain a safe NSX-T operating environment
if($vm.PersistentID -is [array]) {
    Write-host "Machine" $vm.name[0] "is not unique in the vSphere instance.  Update NSX tags manually" 
    exit
} else {
    if($env:function_debug -eq "true") {
        write-host $vm.PersistentID
    }
}

# Get VM objects tags from vCenter and write them to a JSON object
# Create the JSON Tagging structure for NSX
$nsxList = New-Object System.Collections.ArrayList
$tags = Get-VM -name $vm.name | Get-TagAssignment
foreach ($tag in $tags)
{
    $tagString = $tag.tag.ToString()
    $tagArray = $tagString.split('/')
    $nsxList.add(@{"tag"=$tagArray[1];"scope"=$tagArray[0]})
    if($env:function_debug -eq "true") {
        write-host $tagString
    }
}
$nsxJSON = @{}
$nsxJSON.add("external_id",$vm.PersistentId)
$nsxJSON.add("tags",$nsxList)
# Write nsxJSON string to JSON for the NSX REST call payload
$nsxBody = $nsxJSON | ConvertTo-Json -depth 10

# disconnect from VI server
Write-Host "Disconnecting from vCenter Server ..."
Disconnect-VIServer * -Confirm:$false

# Create Basic Auth string for NSX authentication
$pair = "$($SECRETS_CONFIG.NSX_USERNAME):$($SECRETS_CONFIG.NSX_PASSWORD)"
$bytes = [System.Text.Encoding]::ASCII.GetBytes($pair)
$base64 = [System.Convert]::ToBase64String($bytes)
$basicAuthValue = "Basic $base64"

# Render the NSX URL to POST VM Tag update
$nsxUrl = "https://$($SECRETS_CONFIG.NSX_SERVER)/api/v1/fabric/virtual-machines?action=update_tags"

#URL Headers
$headers = @{
    "Authorization"="$basicAuthValue";
    "Accept="="application/json";
    "Content-Type"="application/json";
}
if($env:debug_writehost -eq "true") {
    Write-Host "DEBUG: nsxURL=`"$($nsxUrl | Format-List | Out-String)`""
    Write-Host "DEBUG: headers=`"$($headers | Format-List | Out-String)`""
    Write-Host "DEBUG: nsxbody=`"$($nsxBody | Format-List | Out-String)`""
    Write-Host "DEBUG: Applying vSphere Tags for "$vm.name "to NSX-T"
}

# POST to NSX
$response = ""
if($env:skip_nsx_cert_check = "true") {
    $response = Invoke-Webrequest -Uri $nsxUrl -Method POST -Headers $headers -SkipHeaderValidation -Body $nsxbody -SkipCertificateCheck
} else {
    $response = Invoke-Webrequest -Uri $nsxUrl -Method POST -Headers $headers -SkipHeaderValidation -Body $nsxbody
}
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: Invoke-WebRequest=$response"
}