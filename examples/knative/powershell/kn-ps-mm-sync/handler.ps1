Function Process-Init {
   Write-Host "$(Get-Date) - Processing Init"
   Write-Host "$(Get-Date) - Init Processing Completed"
   Write-Host "---------------------------"
}

Function Process-Shutdown {
   Write-Host "$(Get-Date) - Processing Shutdown"
   Write-Host "$(Get-Date) - Shutdown Processing Completed"
   Write-Host "---------------------------"
}

Function Process-Handler {
   param(
      [Parameter(Position=0,Mandatory=$true)][CloudNative.CloudEvents.CloudEvent]$CloudEvent
   )

## Extract subject from CloudEvent object
$eventSubject=$cloudEvent.subject
Write-Host "$(Get-Date) - Received event from vCenter with subject" $eventSubject

## Form CloudEventData object
$cloudEventData = $cloudEvent | Read-CloudEventJsonData -ErrorAction SilentlyContinue -Depth 10
if($cloudEventData -eq $null) {
   $cloudEventData = $cloudEvent | Read-CloudEventData
   }

## Output CloudEventData to console for debugging
# Write-Host "$(Get-Date) - Full contents of CloudEventData`n $(${cloudEventData} | ConvertTo-Json)`n"

## Extract hostname from CloudEventData object
$esxiHost=$cloudEventData.Host.Name
Write-Host "$(Get-Date) - The event relates to host named" $esxiHost

## Check secret in place which supplies vROps environment variables for debugging
# Write-Host "$(Get-Date) - vropsFqdn:" ${env:vropsFqdn}
# Write-Host "$(Get-Date) - vropsUser:" ${env:vropsUser}
# Write-Host "$(Get-Date) - vropsPassword:" ${env:vropsPassword}

## Form unauthorized headers payload
$headers = @{
   "Content-Type" = "application/json";
   "Accept"  = "application/json"
   }

## Acquire bearer token
$uri = "https://" + $env:vropsFqdn + "/suite-api/api/auth/token/acquire"
$basicAuthBody = @{
   username = $env:vropsUser;
   password = $env:vropsPassword;
   }
$basicAuthBodyJson = $basicAuthBody | ConvertTo-Json -Depth 5
# Write-Host "$(Get-Date) - Acquiring vROps API bearer token ..."
$bearer = Invoke-WebRequest -Uri $uri -Method POST -Headers $headers -Body $basicAuthBodyJson -SkipCertificateCheck | ConvertFrom-Json

## Output vROps API bearer token to console for debugging
# Write-Host "$(Get-Date) - vROps API bearer token is" $bearer.token

## Form authorized headers payload
$authedHeaders = @{
   "Content-Type" = "application/json";
   "Accept"  = "application/json";
   "Authorization" = "vRealizeOpsToken " + $bearer.token
   }

## Get host ResourceID
$uri = "https://" + $env:vropsFqdn + "/suite-api/api/adapterkinds/VMWARE/resourcekinds/HostSystem/resources?name=" + $esxiHost
# Write-Host "$(Get-Date) - Acquiring host ResourceID ..."
$resource = Invoke-WebRequest -Uri $uri -Method GET -Headers $authedHeaders -SkipCertificateCheck
$resourceJson = $resource.Content | ConvertFrom-Json

## Output vROps ResourceID of host to console for debugging
# Write-Host "$(Get-Date) - ResourceID of host is " $resourceJson.resourceList[0].identifier

## Call API to mark host maintenance mode state
$uri = "https://" + $env:vropsFqdn + "/suite-api/api/resources/" + $resourceJson.resourceList[0].identifier + "/maintained"


If ( $eventSubject -eq 'EnteredMaintenanceModeEvent'){
   Write-Host "$(Get-Date) - Attempting to mark host in vROps as being in maintenance mode ..."
   Invoke-WebRequest -Uri $uri -Method PUT -Headers $authedHeaders -SkipCertificateCheck
   }
If ( $eventSubject -eq 'ExitMaintenanceModeEvent'){
   Write-Host "$(Get-Date) - Attempting to mark host in vROps as not being in maintenance mode ..."
   Invoke-WebRequest -Uri $uri -Method DELETE -Headers $authedHeaders -SkipCertificateCheck
   }

## Get host maintenance mode state
$uri = "https://" + $env:vropsFqdn + "/suite-api/api/adapterkinds/VMWARE/resourcekinds/HostSystem/resources?name=" + $esxiHost
# Write-Host "Checking vROps host maintenance mode state ..."
$resource = Invoke-WebRequest -Uri $uri -Method GET -Headers $authedHeaders -SkipCertificateCheck
$resourceJson = $resource.Content | ConvertFrom-Json
Write-Host "$(Get-Date) -" $esxiHost "vROps maintenence mode state is now " $resourceJson.resourceList[0].resourceStatusStates[0].resourceState "( STARTED=Not In Maintenance | MAINTAINED_MANUAL=In Maintenance )"
Write-Host "---------------------------"
}
