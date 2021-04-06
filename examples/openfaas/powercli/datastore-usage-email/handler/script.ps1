# Process function Secrets passed in
$VC_CONFIG_FILE = "/var/openfaas/secrets/vc-datastore-config"
$VC_CONFIG = (Get-Content -Raw -Path $VC_CONFIG_FILE | ConvertFrom-Json)
if($env:function_debug -eq "true") {
    Write-host "DEBUG: `"$VC_CONFIG`""
}

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

$alarmName = ($json.data.alarm.name -replace "\n"," ")
$datastoreName = $json.data.ds.name
$alarmStatus = $json.data.to
$vcenter = ($json.source -replace "/sdk","")
$datacenter = $json.data.datacenter.name

if($env:function_debug -eq "true") {
    Write-Host "DEBUG: alarmName: `"$alarmName`""
    Write-host "DEBUG: datastoreName: `"$datastoreName`""
    Write-Host "DEBUG: alarmStatus: `"$alarmStatus`""
    Write-Host "DEBUG: vcenter: `"$vcenter`""
}

if( ("$alarmName" -match "$($VC_CONFIG.VC_ALARM_NAME)") -and ([bool]($VC_CONFIG.DATASTORE_NAMES -match "$datastoreName")) -and ($alarmStatus -eq "yellow" -or $alarmStatus -eq "red" -or $alarmStatus -eq "green") ) {

    # Warning Email Body
    if($alarmStatus -eq "yellow") {
        $subject = "⚠️ $($VC_CONFIG.EMAIL_SUBJECT) ⚠️ "
        $threshold = "warning"
    } elseif($alarmStatus -eq "red") {
        $subject = "☢️ $($VC_CONFIG.EMAIL_SUBJECT) ☢️ "
        $threshold = "error"
    } elseif($alarmStatus -eq "green") {
        $subject = "$($VC_CONFIG.EMAIL_SUBJECT)"
        $threshold = "normal"
	}

    $Body = "$alarmName $datastoreName has reached $threshold threshold.`r`n"
	
	if ( $threshold -ne "normal" )
	{
		$Body = $Body + "Please log in to your VMware Cloud on AWS environment and ensure that everything is operating as expected.`r`n"
	}

	$Body = $Body + @"
	    vCenter Server: $vcenter
        Datacenter: $datacenter
        Datastore: $datastoreName	
"@

    $emailTo = $VC_CONFIG.EMAIL_TO 
    
    # If the JSON file has a custom property email field defined, log into vCenter to find the value
    # This is used to allow admins within vCenter to add an email address for storage alarms independent of the EMAIL_TO value
    if ($VC_CONFIG.DATASTORE_CUSTOM_PROP_EMAIL_TO.length -gt 0)
    {
        Set-PowerCLIConfiguration -InvalidCertificateAction Ignore  -DisplayDeprecationWarnings $false -ParticipateInCeip $false -Confirm:$false | Out-Null
        
        # Connect to vCenter Server
        Write-Host "Connecting to vCenter Server ..."
        Connect-VIServer -Server $($VC_CONFIG.VC) -User $($VC_CONFIG.VC_USERNAME) -Password $($VC_CONFIG.VC_PASSWORD)

        # This objet has all defined custom fields in the vCenter
        $customFieldMgr = Get-View ($global:DefaultVIServer.ExtensionData.Content.CustomFieldsManager)

        $datastoreView = Get-View -ViewType Datastore -Property Name, Value -Filter @{"name"=$datastoreName}

        # Build 2 hash tables for the key-value pairs in the Custom Fields Manager, one to search by Custom Field ID and one to search by Name
        $customKeyLookup = @{}
        $customNameLookup = @{}
        $customFieldMgr.Field | ForEach-Object {
            $customKeyLookup.Add($_.Key, $_.Name)          
            $customNameLookup.Add($_.Name, $_.Key)
        }

        # This is the custom field that we're looking to pull an email address out of
        $emailKey = $customNameLookup[$($VC_CONFIG.DATASTORE_CUSTOM_PROP_EMAIL_TO)]

        #If we find one, this is the email address we will add to the "To" field in the email
        $addEmailAddress = ""
        foreach ($row in $datastoreView.Value) {
            if ($env:function_debug -eq "true") {
                Write-Host "`Datastore:" $datastoreName "has Custom Field:" $customKeyLookup[$row.Key] "with value:" $row.Value "`n"
            }
            if ($row.Key -eq $emailKey) 
            {
                if ($env:function_debug -eq "true") {
                    write-host "Found key" $emailKey "with value" $row.value
                }
                $addEmailAddress = $row.value
            }

        }
        
        if ($addEmailAddress.length -gt 0){
            $emailTo = $emailTo + $addEmailAddress
        }
        else {
            Write-Host "DATASTORE_CUSTOM_PROP_EMAIL_TO value '"$VC_CONFIG.DATASTORE_CUSTOM_PROP_EMAIL_TO "' found in JSON config but not found on datastore"
        }
        Write-Host "Disconnecting from vCenter Server ..."
        Disconnect-VIServer * -Confirm:$false

    }

    # If defined in the config file, send via authenticated SMTP, otherwise use standard SMTP
	if ($VC_CONFIG.SMTP_PASSWORD.length -gt 0 -and $VC_CONFIG.SMTP_USERNAME.length -gt 0)
	{
		$password = ConvertTo-SecureString "$($VC_CONFIG.SMTP_PASSWORD)" -AsPlainText -Force
		$credential = New-Object System.Management.Automation.PSCredential($($VC_CONFIG.SMTP_USERNAME), $password)
		Send-MailMessage -From $($VC_CONFIG.EMAIL_FROM) -to $($emailTo)  -Subject $Subject -Body $Body -SmtpServer $($VC_CONFIG.SMTP_SERVER) -port $($VC_CONFIG.SMTP_PORT) -UseSsl -Credential $credential -Encoding UTF32
	}
	else
	{
		Send-MailMessage -From $($VC_CONFIG.EMAIL_FROM) -to $($emailTo) -Subject $Subject -Body $Body -SmtpServer $($VC_CONFIG.SMTP_SERVER) -port $($VC_CONFIG.SMTP_PORT) -Encoding UTF32
	}
}