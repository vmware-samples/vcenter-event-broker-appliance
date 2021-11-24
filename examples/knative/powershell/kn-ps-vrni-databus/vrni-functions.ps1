Import-Module CloudEvents.Sdk


function New-vRNICloudEventsFromDatabus {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]$Data,
        [Parameter(Mandatory = $true)]$cloudEventSource
    )

    # empty container for the cloud events. vRNI can send multiple events in a single request, so we need to split it up
    $cloudEvents = @()
    $jsonData = $Data | ConvertFrom-Json -AsHashtable

    # see whether the payload body has problems or application updates and process them differently
    # first up: problems
    if ($jsonData.ContainsKey("EntityMessageList")) {
        # loop through all problems and create CloudEvents from them to give back
        # to the main function (which fires them into the event broker)
        foreach ($problemEvent in $jsonData["EntityMessageList"]) {

            if (!$problemEvent.ContainsKey("entity")) {
                Write-host "$(Get-Date) - Wrong format for problem event, skipping:`n"
                Write-host $problemEvent
                continue
            }

            # the payload has a "status" field with "UPDATE", "OPEN", or "CLOSE" to figure out what the problem update is about
            # determine the cloud event type based on the status field
            $cloudEventType = "com.vmware.event.router/vrni.problem.v0"
            if ($problemEvent.ContainsKey("status")) {
                if ($problemEvent["status"] -eq "OPEN") {
                    $cloudEventType = "com.vmware.event.router/vrni.problem.opened.v0"
                }
                elseif ($problemEvent["status"] -eq "UPDATE") {
                    $cloudEventType = "com.vmware.event.router/vrni.problem.updated.v0"
                }
                elseif ($problemEvent["status"] -eq "CLOSE") {
                    $cloudEventType = "com.vmware.event.router/vrni.problem.closed.v0"
                }
            }

            try {
                $cloudEvent = New-CloudEvent -Type $cloudEventType -Source $cloudEventSource -Id (New-Guid).Guid -Time (Get-Date)
                $cloudEvent = $cloudEvent | Set-CloudEventData -DataContentType "application/json" -Data $problemEvent
            }
            catch {
                Write-host "$(Get-Date) - Failed to construct CloudEvent for problem event, skipping:`n"
                Write-host $problemEvent
                continue
            }

            $cloudEvents += $cloudEvent
        }
    }

    # see if there are any application updates in the payload, and split them up into induvidial cloud events
    if ($jsonData.ContainsKey("ApplicationMessageList")) {
        # loop through all application updates and create CloudEvents from them to give back
        # to the main function (which fires them into the event broker)
        foreach ($applicationEvent in $jsonData["ApplicationMessageList"]) {

            if (!$applicationEvent.ContainsKey("info")) {
                Write-host "$(Get-Date) - Wrong format for application update event, skipping:`n"
                Write-host $applicationEvent
                continue
            }

            # the payload has a "status" field with "UPDATE", "CREATE", or "DELETE" to figure out what the update is about
            $cloudEventType = "com.vmware.event.router/vrni.application.v0"

            try {
                $cloudEvent = New-CloudEvent -Type $cloudEventType -Source $cloudEventSource -Id (New-Guid).Guid -Time (Get-Date)
                $cloudEvent = $cloudEvent | Set-CloudEventData -DataContentType "application/json" -Data $applicationEvent
            }
            catch {
                Write-host "$(Get-Date) - Failed to construct CloudEvent for application update event, skipping:`n"
                Write-host $applicationEvent
                continue
            }

            $cloudEvents += $cloudEvent
        }
    }


    return $cloudEvents
}
