Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   try {
      $jsonSecrets = ${env:VDS_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:VDS_SECRET does not look to be defined"
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
   }
   catch {
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
   }
   catch {
      throw "`nPayload must be JSON encoded"
   }

   try {
      $jsonSecrets = ${env:VDS_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:VDS_SECRET does not look to be defined"
   }

   $ENFORCED_MTU = ${jsonSecrets}.ENFORCED_MTU

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "$(Get-Date) - DEBUG: K8s secret:`n${env:VDS_SECRET}`n"

      Write-Host "$(Get-Date) - DEBUG: Enforced MTU:`n$ENFORCED_MTU`n"

      Write-Host "$(Get-Date) - DEBUG: CloudEventData`n $(${cloudEventData} | Out-String)`n"
   }

   # Extract VM Name from event
   $vdsName = $cloudEventData.Dvs['Name']
   $currentMTU = $cloudEventData.ConfigSpec['MaxMtu']

   if(${env:FUNCTION_DEBUG} -eq "true") {
      Write-Host "Found VDS from event: $vdsName"
      Write-Host "Current MTU from event: $currentMTU"
   }

   if ( $currentMTU -ne $ENFORCED_MTU) {
      try {
         $vswitch = Get-VDSwitch "$vdsName"
      }
      catch {
         Write-Host "$(Get-Date) - ERROR: unable to retrieve VDS"
         throw $_
      }

      if(${env:FUNCTION_DEBUG} -eq "true") {
         Write-Host "$(Get-Date) - DEBUG: MTU on the switch currently set to"$vswitch.mtu
      }

      Write-Host "Resetting MTU to"$ENFORCED_MTU
      try {
         Set-VDSwitch $vswitch -Mtu "$ENFORCED_MTU"
      }
      catch {
         Write-Host "$(Get-Date) - ERROR: unable to set MTU on"$vswitch.Name
         throw $_
      }
   }
   else {
      Write-Host "$(Get-Date) - INFO: Current MTU of $currentMTU matches the enforced value - no reconfiguration was done."
   }


   Write-Host "$(Get-Date) - VDS reconfig operation complete ...`n"

   Write-Host "$(Get-Date) - Handler Processing Completed ...`n"
}
