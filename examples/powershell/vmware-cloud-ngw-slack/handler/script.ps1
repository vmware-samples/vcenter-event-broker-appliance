
# Process function Secrets passed in
$SECRETS_FILE = "/var/openfaas/secrets/ngw-slack-config"
$SECRETS_CONFIG = (Get-Content -Raw -Path $SECRETS_FILE | ConvertFrom-Json)

# Process payload sent from NGW

$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

if( ($json.event_id -eq "SDDC-PROVISION") -or ($json.event_id -eq "SDDC-DELETE") ) {

    if($json.event_id -eq "SDDC-PROVISION") {
        $sddcUrl = "https://vmc.vmware.com/console/sddcs/$($json.resource_id)"

        $payload = @{
            attachments = @(
                @{
                    pretext = ":party: New VMC SDDC Provisioned :party:";
                    fields = @(
                        @{
                            title = "SDDC";
                            value = $json.resource_name;
                            short = "false";
                        }
                        @{
                            title = "Org";
                            value = ($json.org_name -replace "`n",' ');
                            short = "false";
                        }
                        @{
                            title = "Date";
                            value = $json.send_date_time;
                            short = "false";
                        }
                        @{
                            title = "User";
                            value = $json.user_name;
                            short = "false";
                        }
                        @{
                            title = "URL";
                            value ="<$sddcUrl|*Click to open SDDC*>";
                            short = "false";
                        }
                    )
                    footer_icon = "https://raw.githubusercontent.com/vmware-samples/vcenter-event-broker-appliance/development/logo/veba_otto_the_orca_320x320.png";
                    footer = "Powered by VEBA";
                }
            )
        }
    }

    if($json.event_id -eq "SDDC-DELETE") {
        $payload = @{
            attachments = @(
                @{
                    pretext = ":rotating_light: SDDC Deleted :rotating_light:";
                    fields = @(
                        @{
                            title = "SDDC";
                            value = $json.resource_name;
                            short = "false";
                        }
                        @{
                            title = "Org";
                            value = ($json.org_name -replace "`n",' ');
                            short = "false";
                        }
                        @{
                            title = "DateTime";
                            value = $json.send_date_time;
                            short = "false";
                        }
                        @{
                            title = "User";
                            value = $json.user_name;
                            short = "false";
                        }
                    )
                    footer_icon = "https://raw.githubusercontent.com/vmware-samples/vcenter-event-broker-appliance/development/logo/veba_otto_the_orca_320x320.png";
                    footer = "Powered by VEBA";
                }
            )
        }
    }

    $body = $payload | ConvertTo-Json -Depth 5
    if($env:function_debug -eq "true") {
        Write-Host "DEBUG: body=$body"
    }

    Write-Host "Invoking Webhook URL ..."
    Invoke-WebRequest -Uri $($SECRETS_CONFIG.SLACK_WEBHOOK_URL) -Method POST -ContentType "application/json" -Body $body
} else {
    Write-Host "Function executed but there is no notification for EventID: $($json.event_id)"
}


