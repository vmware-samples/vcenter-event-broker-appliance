# Credit to Craig Straka (craig.straka@it-partners.com) for writing the original version of this

Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init"

   try {
      $jsonSecrets = ${env:TAG_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "K8s secrets `$env:TAG_SECRET does not look to be defined"
   }

   # Extract all tag secrets for ease of use in function
   $VCENTER_SERVER = ${jsonSecrets}.VCENTER_SERVER
   $VCENTER_USERNAME = ${jsonSecrets}.VCENTER_USERNAME
   $VCENTER_PASSWORD = ${jsonSecrets}.VCENTER_PASSWORD
   $VCENTER_CERTIFICATE_ACTION = ${jsonSecrets}.VCENTER_CERTIFICATE_ACTION

   # Configure TLS 1.2/1.3 support as this is required for latest vSphere release
   [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor [System.Net.SecurityProtocolType]::Tls12 -bor [System.Net.SecurityProtocolType]::Tls13

   Write-Host "$(Get-Date) - Configuring PowerCLI Configuration Settings"
   Set-PowerCLIConfiguration -InvalidCertificateAction:${VCENTER_CERTIFICATE_ACTION} -ParticipateInCeip:$true -Confirm:$false

   Write-Host "$(Get-Date) - Connecting to vCenter Server $VCENTER_SERVER"

   try {
      Connect-VIServer -Server $VCENTER_SERVER -User $VCENTER_USERNAME -Password $VCENTER_PASSWORD
   }
   catch {
      Write-Error "$(Get-Date) - Failed to connect to vCenter Server"
      throw $_
   }

   Write-Host "$(Get-Date) - Successfully connected to $VCENTER_SERVER"

   Write-Host "$(Get-Date) - Init Processing Completed"
}

Function Process-Shutdown {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Shutdown"

   Write-Host "$(Get-Date) - Disconnecting from vCenter Server"

   try {
      Disconnect-VIServer * -Confirm:$false
   }
   catch {
      Write-Error "$(Get-Date) - Failed to Disconnect from vCenter Server"
   }

   Write-Host "$(Get-Date) - Shutdown Processing Completed"
}

Function Process-Handler {
   [CmdletBinding()]
   param(
      [Parameter(Position = 0, Mandatory = $true)][CloudNative.CloudEvents.CloudEvent]$CloudEvent
   )

   # Decode CloudEvent
   try {
      $cloudEventData = $cloudEvent | Read-CloudEventJsonData -Depth 10
   }
   catch {
      throw "Payload must be JSON encoded"
   }

   try {
      $jsonSecrets = ${env:TAG_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "K8s secrets `$env:TAG_SECRET does not look to be defined"
   }

   # enable/disable DEBUG mode
   if (${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: K8s Secrets:`n${env:TAG_SECRET}"
      Write-Host "$(Get-Date) - DEBUG: CloudEventData:`n $(${cloudEventData} | Out-String)"
   }
   
   # NSX settings
   $NSX_SERVER = ${jsonSecrets}.NSX_SERVER
   $NSX_USERNAME = ${jsonSecrets}.NSX_USERNAME
   $NSX_PASSWORD = ${jsonSecrets}.NSX_PASSWORD
   $NSX_SKIP_CERT_CHECK = ${jsonSecrets}.NSX_SKIP_CERT_CHECK
   
   # Pull VM name from event
   $vmname = ($cloudEventData.Arguments | where-object { $_.Key -eq "Object" }).Value

   # Test for existince of content in $vmname variable and exit handler early if vm value is missing
   if (!$vmname) {
      throw "$(Get-Date) - ERROR: unable to retrieve VM Object from Event payload"
   } 

   $arguments = $cloudEventData.Arguments | Out-String
   Write-Host "$(Get-Date) - DEBUG: CloudEventDataArguments:`n $arguments"
   Write-Host "$(Get-Date) - DEBUG: VM name: $vmname"

   # Get VM object from vCenter
   try {
      $vm = Get-VM -name $vmname | Select-Object Name, PersistentId
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: unable to retrieve VM object"
      throw $_
   }

   # until uniquely identifiable VM data is provided in a vSphere event this is the only option to maintain a safe NSX-T operating environment
   if ($vm.PersistentID -is [array]) {
      throw "$(Get-Date) - ERROR: Machine $($vm.name[0]) is not unique in the vSphere instance: update NSX tags manually" 
   }

   if (${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: VM Persistence ID: $($vm.PersistentID)"
   }

   # Get VM objects tags from vCenter and write them to a JSON object
   # Create the JSON Tagging structure for NSX
   $nsxTagList = New-Object System.Collections.ArrayList
   try {
      $tags = Get-VM -name $vm.name | Get-TagAssignment
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: unable to retrieve tags for VM object"
      throw $_
   }

   foreach ($tag in $tags) {
      $tagString = $tag.tag.ToString()
      $tagArray = $tagString.split('/')
      $nsxTagList.add(@{"tag" = $tagArray[1]; "scope" = $tagArray[0] })

      if (${env:FUNCTION_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - DEBUG: Tag: ${tagstring}"
      }
   }

   # Create the JSON Tagging structure for NSX
   $nsxJSON = @{}
   $nsxJSON.add("external_id", $vm.PersistentId)
   $nsxJSON.add("tags", $nsxTagList)

   # Write nsxJSON string to JSON for the NSX REST call payload
   try {
      $nsxBody = $nsxJSON | ConvertTo-Json -depth 10
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: convert NSX tag assignments to JSON"
      throw $_
   }

   # Create Basic Auth string for NSX authentication
   $pair = "$($NSX_USERNAME):$($NSX_PASSWORD)"
   $bytes = [System.Text.Encoding]::ASCII.GetBytes($pair)
   $base64 = [System.Convert]::ToBase64String($bytes)
   $basicAuthValue = "Basic $base64"

   # Render the NSX URL to POST VM Tag update
   $nsxUrl = "https://$($NSX_SERVER)/api/v1/fabric/virtual-machines?action=update_tags"

   #URL Headers
   $headers = @{
      "Authorization" = "$basicAuthValue";
      "Accept="       = "application/json";
      "Content-Type"  = "application/json";
   }

   if (${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: nsxURL=`"$($nsxUrl | Format-List | Out-String)`""
      Write-Host "$(Get-Date) - DEBUG: headers=`"$($headers | Format-List | Out-String)`""
      Write-Host "$(Get-Date) - DEBUG: nsxbody=`"$($nsxBody | Format-List | Out-String)`""
      Write-Host "$(Get-Date) - DEBUG: Applying vSphere Tags for "$vm.name "to NSX-T"
   }

   # POST to NSX
   try {
      $response = ""
      if ($NSX_SKIP_CERT_CHECK -eq "true") {
         $response = Invoke-Webrequest -Uri $nsxUrl -Method POST -Headers $headers -SkipHeaderValidation -Body $nsxbody -SkipCertificateCheck
      }
      else {
         $response = Invoke-Webrequest -Uri $nsxUrl -Method POST -Headers $headers -SkipHeaderValidation -Body $nsxbody
      }
   
      if (${env:FUNCTION_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - DEBUG: Invoke-WebRequest response=$($response)"
      }
   
      Write-Host "$(Get-Date) - vSphere Tag to NSX Operation complete"
      Write-Host "$(Get-Date) - Handler Processing complete"
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: send NSX tag assignments web request"
      throw $_
   }
}
