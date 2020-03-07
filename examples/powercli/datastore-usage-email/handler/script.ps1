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
    
    if ( $threshold -ne "normal" ) {
        $Body = $Body + "Please log in to your VMware Cloud on AWS environment and ensure that everything is operating as expected.`r`n"
    }

    $Body = $Body + @"
        vCenter Server: $vcenter
        Datacenter: $datacenter
        Datastore: $datastoreName	
"@
    if ($VC_CONFIG.SMTP_PASSWORD.length -gt 0 -and $VC_CONFIG.SMTP_USERNAME.length -gt 0) {
        $password = ConvertTo-SecureString "$($VC_CONFIG.SMTP_PASSWORD)" -AsPlainText -Force
        $credential = New-Object System.Management.Automation.PSCredential($($VC_CONFIG.SMTP_USERNAME), $password)
        Send-MailMessage -From $($VC_CONFIG.EMAIL_FROM) -to $($VC_CONFIG.EMAIL_TO) -Subject $Subject -Body $Body -SmtpServer $($VC_CONFIG.SMTP_SERVER) -port $($VC_CONFIG.SMTP_PORT) -UseSsl -Credential $credential -Encoding UTF32
    } else {
        Send-MailMessage -From $($VC_CONFIG.EMAIL_FROM) -to $($VC_CONFIG.EMAIL_TO) -Subject $Subject -Body $Body -SmtpServer $($VC_CONFIG.SMTP_SERVER) -port $($VC_CONFIG.SMTP_PORT) -Encoding UTF32
    }
}

