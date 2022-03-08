Function Process-Init {
   [CmdletBinding()]
   param()
   Write-Host "$(Get-Date) - Processing Init`n"

   try {
      $jsonSecrets = ${env:FUNCTION_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:FUNCTION_SECRET does not look to be defined"
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
      $jsonSecrets = ${env:FUNCTION_SECRET} | ConvertFrom-Json
   }
   catch {
      throw "`nK8s secret `$env:FUNCTION_SECRET does not look to be defined"
   }

   #
   # Your custom code goes here 
   #
   # When sending messages back to the console, please conform to the following standards: 
   # Write-Host "$(Get-Date) - DEBUG: "
   # Write-Host "$(Get-Date) - WARN: "
   # Write-Host "$(Get-Date) - ERROR: "

   # This is the final line of your custom code.
   # Replace #REPLACE-FN-NAME# with a meaningful message showing the end of your custom code
   #Write-Host "$(Get-Date) - #REPLACE-FN-NAME# operation complete ...`n"

   Write-Host "$(Get-Date) - Handler Processing Completed ...`n"
}
