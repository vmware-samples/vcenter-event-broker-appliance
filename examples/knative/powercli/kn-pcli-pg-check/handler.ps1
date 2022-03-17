Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   try {
      $jsonSecrets = ${env:PG_CHECK_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:PG_CHECK_SECRET does not look to be defined"
   }

   # Extract all tag secrets for ease of use in function
   $VCENTER_SERVER = ${jsonSecrets}.VCENTER_SERVER
   $VCENTER_USERNAME = ${jsonSecrets}.VCENTER_USERNAME
   $VCENTER_PASSWORD = ${jsonSecrets}.VCENTER_PASSWORD
   $VCENTER_CERTIFICATE_ACTION = ${jsonSecrets}.VCENTER_CERTIFICATE_ACTION

   # Configure TLS 1.2/1.3 support as this is required for latest vSphere release
   [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor [System.Net.SecurityProtocolType]::Tls12 -bor [System.Net.SecurityProtocolType]::Tls13

   Write-Host "$(Get-Date) - Configuring PowerCLI Configuration Settings`n"
   Set-PowerCLIConfiguration -InvalidCertificateAction:${VCENTER_CERTIFICATE_ACTION} -ParticipateInCeip:$true -Confirm:$false

   Write-Host "$(Get-Date) - Connecting to vCenter Server $VCENTER_SERVER`n"

   try {
      Connect-VIServer -Server $VCENTER_SERVER -User $VCENTER_USERNAME -Password $VCENTER_PASSWORD
   }
   catch {
      Write-Error "$(Get-Date) - ERROR: Failed to connect to vCenter Server"
      throw $_
   }

   Write-Host "$(Get-Date) - Successfully connected to $VCENTER_SERVER`n"

   Write-Host "$(Get-Date) - Init Processing Completed`n"
}

Function Process-Shutdown {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Shutdown`n"

   Write-Host "$(Get-Date) - Disconnecting from vCenter Server`n"

   try {
      Disconnect-VIServer * -Confirm:$false
   }
   catch {
      Write-Error "$(Get-Date) - Error: Failed to Disconnect from vCenter Server"
   }

   Write-Host "$(Get-Date) - Shutdown Processing Completed`n"
}

Function Process-Handler {
   [CmdletBinding()]
   param(
      [Parameter(Position=0,Mandatory=$true)][CloudNative.CloudEvents.CloudEvent]$CloudEvent
   )

   # Decode CloudEvent
   try {
      $cloudEventData = $cloudEvent | Read-CloudEventJsonData -Depth 10
   }
   catch {
      throw "`nPayload must be JSON encoded"
   }

   try {
      $jsonSecrets = ${env:PG_CHECK_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:PG_CHECK_SECRET does not look to be defined"
   }

   $VM_WATCH_TAGS = ${jsonSecrets}.VM_WATCH_TAGS
   $PG_WATCH_TAGS = ${jsonSecrets}.PG_WATCH_TAGS

   $deviceUnchanged = ($NULL -eq $cloudEventData.ConfigSpec.DeviceChange )
   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: ConfigSpec.DeviceChange Is Null? $($deviceUnchanged)`n"
   }

   # If no devices changed, then the NIC could not have been updated to a different portgroup
   if ($deviceUnchanged)
   {
      Write-Host "$(Get-Date) - No devices changed.`n"
      return
   }

   # Build the MoRef ID
   $vmID = $cloudEventData.Vm.Vm.Type + "-" + $cloudEventData.Vm.Vm.Value
   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: Vm.Vm.Type Is Null? $($NULL -eq $cloudEventData.Vm.Vm.Type)`n"
      Write-Host "$(Get-Date) - DEBUG: Vm.Vm.Value Is Null? $($NULL -eq $cloudEventData.Vm.Vm.Value)`n"
      Write-Host "$(Get-Date) - DEBUG: vmID is $vmID`n"
   }

   # Retrieve the VM object by MoRef ID
   try {
      $vm = Get-VM -id $vmID
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: unable to retrieve VM ID $vmID`n"
      throw $_
   }

   # Retrieve all tags on the VM
   Write-Host "$(Get-Date) - Retrieving tags on $($vm.Name)`n"
   try {
      $vmTags = $vm | Get-TagAssignment
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: unable to retrieve tags for $($vm.Name)`n"
      throw $_
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: Tags found on VM: $($vmTags.tag)`n"
      Write-Host "$(Get-Date) - DEBUG: Tags to monitor: $($VM_WATCH_TAGS)`n"
   }

   # Search through the tags found on the VM and compare them to the VM tags specified in the secret. 
   # If any match is found, break out of the loop - finding any match means the VM is tagged for further inspection
   $checkVM = $false
   :outer foreach ($tag in $vmTags) {
      if(${env:FUNCTION_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - DEBUG: Comparing VM tag: $($tag.Tag) on VM $($vm.Name)`n"
      }
      foreach ($watchTag in $VM_WATCH_TAGS) {
         if(${env:FUNCTION_DEBUG} -eq "true") {
            Write-Host "$(Get-Date) - DEBUG: Comparing watch tag: $($watchTag)`n"
         }
         if ($watchTag -eq $tag.Tag) {
            if(${env:FUNCTION_DEBUG} -eq "true") {
               Write-Host "$(Get-Date) - DEBUG: Match found for: $($watchTag), breaking outer loop`n"
            }
            $checkVM = $true
            break outer
         }
      }
   }

   # If the VM isn't tagged, no further inspection is necessary
   if ($checkVM -eq $false) {
      Write-Host "$(Get-Date) - VM not tagged for monitoring, no inspection required.`n"
      return
   }

   Write-Host "$(Get-Date) - Match found for $($vm.Name), checking portgroups`n"

   # Try retrieving the VM's network adapters
   try {
      $networkAdapters = $vm | Get-NetworkAdapter
   }
   catch {
      Write-Host "$(Get-Date) - ERROR: unable to retrieve retrieve network adapters for $($vm.Name)`n"
      throw $_
   }

   # This hash will store any network adapters that are attached to unapproved portgroups
   $networkAdapterHash = @{}

   foreach ($nic in $networkAdapters) {
      if(${env:FUNCTION_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - DEBUG: VM $($vm.Name) - NIC $($nic.Name) - PortGroup $($nic.NetworkName)`n"
      }
      # If the portgroup is a VSS portgroup, it will have a backing network property
      $vssBackingNetwork = $nic.extensiondata.Backing.Network
      # If the portgroup is a DVS portgroup, it will have a PortgroupKey property
      $dvPortGroupKey = $nic.extensiondata.Backing.Port.PortgroupKey
      $vNetworkID = $null

      if ($NULL -ne $vssBackingNetwork) {
         if(${env:FUNCTION_DEBUG} -eq "true") {
            Write-Host "$(Get-Date) - DEBUG: VSS Backing network is $($vssBackingNetwork)`n"
         }
         # vssBackingNetwork is already in MoRef ID format
         $vNetworkID = $vssBackingNetwork
      }
      elseif ($NULL -ne $dvPortGroupKey) {
         if (${env:FUNCTION_DEBUG} -eq "true") {
            Write-Host "$(Get-Date) - DEBUG: DVS Port Group Key is $($dvPortGroupKey)`n"
         }
         # Unlike the VSS above, We have to build the MoRef ID for VDS
         $vNetworkID = "DistributedVirtualPortgroup-" + $dvPortGroupKey
      } else {
         Write-Host "$(Get-Date) - ERROR: Could not determine network type`n"
         throw "$(Get-Date) - ERROR: Could not determine network type for $($nic.Name)`n"
      }

      if(${env:FUNCTION_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - DEBUG: Virtual Network ID: $($vNetworkID)`n"
      }

      # Now that we've determined the ID, we can try retrieving the portgroup virtual network information
      try {
         $pg = Get-VirtualNetwork -Id $vNetworkID
      }
      catch {
         Write-Host "$(Get-Date) - ERROR: unable to retrieve retrieve virtual network information for $($vNetworkID)`n"
         throw $_
      }

      # Retrieve the tag assignments for the port group
      try {
         $pgTags = $pg | Get-TagAssignment
      }
      catch {
         Write-Host "$(Get-Date) - ERROR: unable to retrieve retrieve tag assignments for $($pg.Name)`n"
         throw $_
      }

      # If $null, the NIC is on a pg with no tags - add it to the hash table
      if ($null -eq $pgTags) {
         $networkAdapterHash[$nic.id] = "$($pg.Name)"
      } else {
         $pgMatch = $false
         # Search through all tags on the portgroup, looking for match on any of the watch tags 
         # provided in the secret. If there is a match on any tag, stop processing - we only need
         # one tag match for the portgroup to be marked as OK
         :outer foreach ($pgTag in $pgTags) {
            $fullTag = $pgTag.Tag.Category.ToString() + "/" + $pgTag.Tag.Name.ToString()
            if(${env:FUNCTION_DEBUG} -eq "true") {
               Write-Host "$(Get-Date) - DEBUG: PortGroup Tag: $($fullTag)`n"
            }

            foreach ($watchTag in $PG_WATCH_TAGS) {
               if ($fullTag -eq $watchTag) {
                  Write-Host "$(Get-Date) - INFO: Found a match on $($watchTag)`n"
                  $pgMatch = $true
                  break outer
               }
            }
         }

         # If none of the portgroup's tags are a match, the vNIC is added to the hash table
         if ($pgMatch -eq $false ) {
            Write-Host "$(Get-Date) - INFO: No permitted tags were found on the portgroup`n"
            $networkAdapterHash[$nic.id] = "$($pg.Name)"
         }
      }
   }

   # Check to see if the hash is empty
   if ( $networkAdapterHash.Count -eq 0 ) {
      Write-Host "$(Get-Date) - INFO: All NICs are on approved portgroups"
      return
   }

   $msg = "NICs using unapproved portgroups:`n"
   Write-Host "$(Get-Date) - $($msg)"
   # Build a list of NICs and unapproved portgroups
   foreach ($nic in $networkAdapters) {
      if ($networkAdapterHash.ContainsKey($nic.Id)) {
         Write-Host "$(Get-Date) - $($nic.Name): on unapproved portgroup $($networkAdapterHash[$nic.Id])"
         $msg += $nic.Name + " - " + $networkAdapterHash[$nic.Id] + "`n"
      }
   }

   # Payload for Slack
   $payload = @{
      attachments = @(
         @{
            pretext = $(${jsonSecrets}.SLACK_MESSAGE_PRETEXT);
            fields = @(
               @{
                     title = "EventType";
                     value = $cloudEvent.Subject;
                     short = "false";
               }
               @{
                     title = "Username";
                     value = $cloudEventData.UserName;
                     short = "false";
               }
               @{
                     title = "DateTime";
                     value = $cloudEventData.CreatedTime;
                     short = "false";
               }
               @{
                  title = "Full Message";
                  value = $msg + "`n`n" + $cloudEventData.FullFormattedMessage ;
                  short = "false";
               }
            )
         }
      )
   }

   # Convert Slack message object into JSON
   $body = $payload | ConvertTo-Json -Depth 5

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: `"$body`""
   }

   Write-Host "$(Get-Date) - Sending Webhook payload to Slack ..."
   $ProgressPreference = "SilentlyContinue"

   try {
      Invoke-WebRequest -Uri $(${jsonSecrets}.SLACK_WEBHOOK_URL) -Method POST -ContentType "application/json" -Body $body
   } catch {
      throw "$(Get-Date) - Failed to send Slack Message: $($_)"
   }

   Write-Host "$(Get-Date) - Successfully sent Webhook ..."

   Write-Host "$(Get-Date) - PG Check operation complete ...`n"

   Write-Host "$(Get-Date) - Handler Processing Completed ...`n"
}
