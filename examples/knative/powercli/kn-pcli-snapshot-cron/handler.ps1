Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   try {
      $jsonSecrets = ${env:SNAPSHOT_SECRET} | ConvertFrom-Json
   } catch {
      throw "`nK8s secrets `$env:SNAPSHOT_SECRET does not look to be defined"
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
   } catch {
      Write-Error "$(Get-Date) - Failed to connect to vCenter Server"
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
   } catch {
      Write-Error "$(Get-Date) - Failed to Disconnect from vCenter Server"
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
   } catch {
      throw "`nPayload must be JSON encoded"
   }

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: CloudEvent`n $(${cloudEvent} | Out-String)`n"

      Write-Host "$(Get-Date) - DEBUG: CloudEventData`n $(${cloudEventData} | ConvertTo-Json)`n"
   }

   $ProgressPreference = "SilentlyContinue"

   # Extract snapshot retention configuration & VM list from 'data' payload
   $maxSnapshotSize = ${cloudEventData}.retentionConfig.sizeGB
   $maxSnapshotDays = ${cloudEventData}.retentionConfig.days
   $vmList = ${cloudEventData}.virtualMachines
   $dryRun = ${cloudEventData}.dryRun

   # Check to ensure at least one snapshot policy is defined
   if($maxSnapshotSize -eq "" -and $maxSnapshotDays -eq "") {
      throw "No snapshot retentation policy defined"
   }

   # Check to ensure we have a list of VMs
   if($vmList -eq "" -or $vmList.getType().BaseType.Name -ne "Array") {
      throw "Invalid or missing input found for VM list"
   }

   # Retrieve list of VMs to run snapshot management against
   try {
      $vms = Get-Vm -NoRecursion -Name $vmList
   } catch {
      throw "Failed to retrieve VM list"
   }

   $maxVM = 20
   $vmCount = 0
   foreach ($vm in $vms) {
      if($vmCount -gt $maxVM) {
         Write-Host "$(Get-Date) - Maximum number of VM count ($maxVM) has been exceeded ...`n"
         break
      }

      $vmName = ${vm}.name

      Write-Host "$(Get-Date) - Checking VM: $vmName "
      try {
         $snapshots = $vm | Get-Snapshot
      } catch {
         throw "Failed to retrieve snapshots from VM $vmName with error $($Error[0])"
      }

      $currentDate = Get-Date

      foreach ($snapshot in $snapshots) {
         $snapshotName = ${snapshot}.name
         $removeSnapshot = $false

         # Get snapshot size (GB) and check against max size
         if($maxSnapshotSize -ne "") {
            if($snapshot.sizeGB -gt $maxSnapshotSize) {
               Write-Host "$(Get-Date) - `tSnapshot $snapshotName consumes $(${snapshot}.SizeGB) GB and exceeds maximum size (${maxSnapshotSize} GB)"

               $removeSnapshot = $true
            }
         }

         # Get snapshot creation date & check against current date
         if($maxSnapshotDays -ne "") {
            $snapshotDate = $snapshot.Created
            $snapshotAge = ($currentDate - $snapshotDate).Days

            if($snapshotAge -gt $maxSnapshotDays) {
               Write-Host "$(Get-Date) - `tSnapshot $snapshotName is $snapshotAge days old and exceeds maximum number of days (${maxSnapshotDays})"

               $removeSnapshot = $true
            }
         }

         # Remove snapshot if flagged and dryRun is disabled
         if($dryRun -eq $false -and $removeSnapshot -eq $true) {
            Write-Host "$(Get-Date) - `tSnapshot removal started for $snapshotName"

            try {
               $snapshot | Remove-Snapshot -RunAsync -Confirm:$false
            } catch {
               throw "Failed to remove snapshot $snapshotName from VM $vmName with error $($Error[0])"
            }
         }
      }
      $vmCount++
   }

   Write-Host "$(Get-Date) - Handler Processing Completed ...`n"
}
