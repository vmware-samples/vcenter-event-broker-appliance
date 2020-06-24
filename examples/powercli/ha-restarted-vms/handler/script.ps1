# Process function Secrets passed in
$VC_CONFIG_FILE = "/var/openfaas/secrets/vcconfig-ha-restarted-vms"
$VC_CONFIG = (Get-Content -Raw -Path $VC_CONFIG_FILE | ConvertFrom-Json)
if($env:function_debug -eq "true") {
    Write-host "DEBUG: `"$VC_CONFIG`""
}

# Process payload sent from vCenter Server Event
$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: `"$json`""
}

Set-PowerCLIConfiguration -InvalidCertificateAction Ignore  -DisplayDeprecationWarnings $false -ParticipateInCeip $false -Confirm:$false | Out-Null

# Connect to vCenter Server
Write-Host "Connecting to vCenter Server ..."
Connect-VIServer -Server $($VC_CONFIG.VC) -User $($VC_CONFIG.VC_USERNAME) -Password $($VC_CONFIG.VC_PASSWORD)

# Main processing of gathering specific events
$report = @()
$eventnumber = 1000
# get vCenter EventManager
$si = get-view ServiceInstance
$em = get-view $si.Content.EventManager

# Create Event Filter Spec
$EventFilterSpec = New-Object VMware.Vim.EventFilterSpec
# Get specific Events based on EventTypeIDs for EventFilterSpec
$tgtEvtTypeIDs = "com.vmware.vc.ha.VmRestartedByHAEvent", "com.vmware.vc.HA.DasHostFailedEvent"
$EventFilterSpec.EventTypeId = $tgtEvtTypeIDs
# Time range to look for specific IDs (current day specified) for EventFilterSpec
$EventFilterSpecByTime = New-Object VMware.Vim.EventFilterSpecByTime
$EventFilterSpecByTime.BeginTime = [datetime]::Today
$EventFilterSpecByTime.EndTime = ([datetime]::Today).AddDays(+1)
$EventFilterSpec.Time = $EventFilterSpecByTime
# Create an Event Collector and loop through events storing them into an array
$eCollector = Get-View ($em.CreateCollectorForEvents($EventFilterSpec))
$events = $eCollector.ReadNextEvents($eventnumber)

While ($events) {
    $events | %{
        $report += $_
    }
    $events = $eCollector.ReadNextEvents($eventnumber)
}

$eCollector.DestroyCollector()

# Retrieve Host name of ESXi host which crashed - used for email report
if ($report | Where-Object{$_.EventTypeId -eq "com.vmware.vc.HA.DasHostFailedEvent"}) {
    # Displays just hostname - strips FQDN
    $vmHost = ($report | Where-Object {$_.EventTypeId -eq "com.vmware.vc.HA.DasHostFailedEvent"} | Select-Object ObjectName -ExpandProperty ObjectName -First 1).split(".")[0]
    # Displays host FQDN
    #$vmHost = $report | Where-Object {$_.EventTypeId -eq "com.vmware.vc.HA.DasHostFailedEvent"} | Select-Object ObjectName -ExpandProperty ObjectName -First 1
}

# Set up fields for email body and send email - only sending VMname, time VM restarted on another host, and VM description
$output = $report | Where-Object {$_.ObjectType -eq "VirtualMachine" } | Select-Object ObjectName, @{N = "Date"; E = { $_.CreatedTime } }, @{N = "Description"; E = { (" - " + (Get-view -id $_.vm.vm).config.annotation | Out-String) } } | Sort-Object ObjectName
$msgBody = $output | ConvertTo-Html | Out-String
$subject = "** $vmHost Failure - VMs Restarted **"

# If defined in the config file, send via authenticated SMTP, otherwise use standard SMTP (credit @kremerpatrick)
if ($VC_CONFIG.SMTP_PASSWORD.length -gt 0 -and $VC_CONFIG.SMTP_USERNAME.length -gt 0)
{
    $password = ConvertTo-SecureString "$($VC_CONFIG.SMTP_PASSWORD)" -AsPlainText -Force
    $credential = New-Object System.Management.Automation.PSCredential($($VC_CONFIG.SMTP_USERNAME), $password)
    Send-MailMessage -From $($VC_CONFIG.EMAIL_FROM) -to $($VC_CONFIG.EMAIL_TO)  -Subject $subject -Body $msgBody -BodyAsHtml -SmtpServer $($VC_CONFIG.SMTP_SERVER) -port $($VC_CONFIG.SMTP_PORT) -UseSsl -Credential $credential #-Encoding UTF32
}
else
{
    Send-MailMessage -From $($VC_CONFIG.EMAIL_FROM) -to $($VC_CONFIG.EMAIL_TO) -Subject $subject -Body $msgBody -BodyAsHtml -SmtpServer $($VC_CONFIG.SMTP_SERVER) -port $($VC_CONFIG.SMTP_PORT) -Encoding UTF32
}

# Disconnect from vCenter
Write-Host "Disconnecting from vCenter Server ..."
Disconnect-VIServer * -Confirm:$false
