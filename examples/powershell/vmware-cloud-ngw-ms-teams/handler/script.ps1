
# Process function Secrets passed in
$SECRETS_FILE = "/var/openfaas/secrets/ngw-teams-config"
$SECRETS_CONFIG = (Get-Content -Raw -Path $SECRETS_FILE | ConvertFrom-Json)

# Process payload sent from NGW

$json = $args | ConvertFrom-Json
if($env:function_debug -eq "true") {
    Write-Host "DEBUG: json=`"$($json | Format-List | Out-String)`""
}

if( ($json.event_id -eq "SDDC-PROVISION") -or ($json.event_id -eq "SDDC-DELETE") ) {

    if($json.event_id -eq "SDDC-PROVISION") {
        $sddcUrl = "http://vmc.vmware.com/console/sddcs/$($json.resource_id)"

        $teamsMessage = [PSCustomObject][Ordered]@{
            "@type"      = "MessageCard"
            "@context"   = "http://schema.org/extensions"
            "themeColor" = '0078D7'
            "summary"      = "New VMC SDDC Provisioned"
            "sections"   = @(
                @{
                    "activityTitle" = "&#x1F973; **New VMC SDDC Provisioned** &#x1F973;";
                    "activitySubtitle" = "In $($json.org_name) Organization";
                    "activityImage" = "https://blogs.vmware.com/vsphere/files/2019/07/Icon-2019-VMWonAWS-Primary-354-x-256.png";
                    "facts" = @(
                        @{
                            "name" = "SDDC:";
                            "value" = $json.resource_name;
                        },
                        @{
                            "name" = "Date:";
                            "value" = $json.send_date_time;
                        },
                        @{
                            "name" = "User:";
                            "value" = $json.user_name;
                        },
                        @{
                            "name" = "URL:";
                            "value" = "[Click to open SDDC]($sddcUrl)";
                        }
                    );
                    "markdown" = $true
                }
            )
        }
    }

    if($json.event_id -eq "SDDC-DELETE") {
        $teamsMessage = [PSCustomObject][Ordered]@{
            "@type"      = "MessageCard"
            "@context"   = "http://schema.org/extensions"
            "themeColor" = '0078D7'
            "summary"      = "VMC SDDC Deleted"
            "sections"   = @(
                @{
                    "activityTitle" = "&#x1F6A8; **VMC SDDC Deleted** &#x1F6A8;";
                    "activitySubtitle" = "In $($json.org_name) Organization";
                    "activityImage" = "https://blogs.vmware.com/vsphere/files/2019/07/Icon-2019-VMWonAWS-Primary-354-x-256.png"
                    "facts" = @(
                        @{
                            "name" = "SDDC:";
                            "value" = $json.resource_name;
                        },
                        @{
                            "name" = "Date:";
                            "value" = $json.send_date_time;
                        },
                        @{
                            "name" = "User:";
                            "value" = $json.user_name;
                        }
                    );
                    "markdown" = $true
                }
            )
        }
    }

    $body = $teamsMessage | ConvertTo-Json -Depth 5
    if($env:function_debug -eq "true") {
        Write-Host "DEBUG: body=$body"
    }

    Write-Host "Invoking Webhook URL ..."
    Invoke-WebRequest -Uri $($SECRETS_CONFIG.TEAMS_WEBHOOK_URL) -Method POST -ContentType "application/json" -Body $body | Out-Null
} else {
    Write-Host "Function executed but there is no notification for EventID: $($json.event_id)"
}


