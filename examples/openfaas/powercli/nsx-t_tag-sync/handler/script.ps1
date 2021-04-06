# Author: Craig Straka (craig.straka@it-partners.com)

# Version .5 (aligned with VEBA .5)

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
#           Intention is to refine this as vSphere event messages evolve
#
# Planned enhancements (script today is minimally viable):
#   Updates based on vSphere event changes
#   Hopefully, better vSphere and NSX-T integration will render this script obsolete

# Process function Secrets passed in
$SECRETS_FILE = "/var/openfaas/secrets/nsx-secrets"
$SECRETS_CONFIG = (Get-Content -Raw -Path $SECRETS_FILE | ConvertFrom-Json)

Set-PowerCLIConfiguration -InvalidCertificateAction Ignore  -DisplayDeprecationWarnings $false -ParticipateInCeip $false -Confirm:$false | Out-Null

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json

if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

# Process payload sent from vCenter Server Event
$vcenter = ($json.source -replace "https://","" -replace "/sdk","")

# Pull VM name from event message and set it to variable. 
$vm = ($json.data.Arguments | where-object {$_.key -eq "Object"}).Value

# Test for existince of content in $vm variable and exit script early if test results false
if($vm -eq "") {
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

# Create the JSON Tagging structure for NSX
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