if (${env:PORT}) {
    $url = "http://*:${env:PORT}/"
    $localUrl = "http://localhost:${env:PORT}/"
}
else {
    $url = "http://*:8080/"
    $localUrl = "http://localhost:8080/"
}

if (${env:UNIT_TEST_SERVER}) {
    $url = ${env:UNIT_TEST_SERVER}
    $localUrl = ${env:UNIT_TEST_SERVER}
}

$serverStopMessage = 'break-signal-e2db683c-b8ff-4c4f-8158-c44f734e2bf1'

. ./handler.ps1

$serverSharedState = [hashtable]::Synchronized(@{})
$serverSharedState.ExitCode = 0
$serverSharedState.ListenerOpened = $false

$backgroundServer = Start-ThreadJob {
    param($url, $serverStopMessage, $serverSharedState)

    Import-Module 'Microsoft.PowerShell.Utility'
    Import-Module CloudEvents.Sdk


    . ./handler.ps1

    function Start-HttpCloudEventListener {
        <#
        .SYNOPSIS
        Starts a HTTP listener on specified Url

        .DESCRIPTION
        Starts a HTTP listener and processes requests sequentially using the listener's context in an infinite loop.
        The listener stops graefully on $serverStopMessage request.

        .PARAMETER Url
        Specifies which Url the HTTP listener should process

        .OUTPUTS
        Boolean $true when setver stop request is received, otherwise $null.
        #>

        [CmdletBinding()]
        param(
            [Parameter(
                Mandatory = $true,
                ValueFromPipeline = $false,
                ValueFromPipelineByPropertyName = $false)]
            [ValidateNotNull()]
            [string]
            $Url,

            [Parameter()]
            $ServerSharedState
        )

        $listener = New-Object -Type 'System.Net.HttpListener'
        $listener.AuthenticationSchemes = [System.Net.AuthenticationSchemes]::Anonymous
        $listener.Prefixes.Add($Url)

        $cloudEvent = $null

        try {
            $listener.Start()
            $serverSharedState.ListenerOpened = $true

            while ($true) {
                $context = $listener.GetContext()

                try {
                    # Read Input Stream
                    $buffer = New-Object 'byte[]' -ArgumentList 1024
                    $ms = New-Object 'IO.MemoryStream'
                    $read = 0
                    while (($read = $context.Request.InputStream.Read($buffer, 0, 1024)) -gt 0) {
                        $ms.Write($buffer, 0, $read);
                    }
                    $bodyData = $ms.ToArray()
                    $ms.Dispose()

                    # Read Headers
                    $headers = @{}
                    for ($i = 0; $i -lt $context.Request.Headers.Count; $i++) {
                        $headers[$context.Request.Headers.GetKey($i)] = $context.Request.Headers.GetValues($i)
                    }

                    $cloudEvent = ConvertFrom-HttpMessage -Headers $headers -Body $bodyData

                    if ([System.Text.Encoding]::UTF8.GetString($bodyData) -eq $serverStopMessage) {
                        # Server Stop request
                        $context.Response.StatusCode = [int]([System.Net.HttpStatusCode]::OK)
                        $context.Response.Close();

                        # Runs Shutdown function (defined in handler.ps1) to clean up connections from warm startup
                        try {
                            Process-Shutdown -ErrorAction 'Stop'
                        }
                        catch {
                            Write-Error "`n$(Get-Date) - Shutdown Processing Error: $($_.Exception.ToString())"
                        }

                        # Set function result
                        $true
                        # Break the infinite loop
                        break
                    }

                    if ( $cloudEvent -ne $null ) {
                        try {
                            Process-Handler -CloudEvent $cloudEvent -ErrorAction 'Stop'
                            $context.Response.StatusCode = [int]([System.Net.HttpStatusCode]::OK)
                        }
                        catch {
                            Write-Error "$(Get-Date) - Handler Processing Error: $($_.Exception.ToString())"
                            # TODO: Consider returning more specific HTTP status based on the Process-Handler error.
                            # The handler interface could be extended to provide expectations fot the event and to return HTTP 4xx.
                            $context.Response.StatusCode = [int]([System.Net.HttpStatusCode]::InternalServerError)
                        }
                    }
                    else {
                        $context.Response.StatusCode = [int]([System.Net.HttpStatusCode]::BadRequest)
                    }

                }
                catch {
                    Write-Error "`n$(Get-Date) - HTTP Request Processing Error: $($_.Exception.ToString())"
                    # TODO: Consider returning more specific HTTP status based on the exception.
                    # If the comes from CloudEvents SDK regarding the cloud event formatting the error might be 4xx

                    $context.Response.StatusCode = [int]([System.Net.HttpStatusCode]::InternalServerError)
                }
                finally {
                    $context.Response.Close();
                }
            }
        }
        catch {
            Write-Error "$(Get-Date) - Listener Processing Error: $($_.Exception.ToString())"
            $ServerSharedState.ExitCode = 1
        }
        finally {
            $listener.Stop()
            $ServerSharedState.ListenerOpened = $false
        }
    }

    # Runs Init function (defined in handler.ps1) which can be used to enable warm startup
    try {
        Process-Init -ErrorAction 'Stop'
    }
    catch {
        Write-Error "$(Get-Date) - Init Processing Error: $($_.Exception.ToString())"
        $serverSharedState.ExitCode = 1
        return
    }

    Write-Host "$(Get-Date) - Starting HTTP CloudEvent listener"
    $breakSignal = Start-HttpCloudEventListener -Url $url -ServerSharedState $serverSharedState
    if ($breakSignal) {
        Write-Host "$(Get-Date) - PowerShell HTTP server stop requested"
        break;
    }

} -ArgumentList $url, $serverStopMessage, $serverSharedState

$killEvent = new-object 'System.Threading.AutoResetEvent' -ArgumentList $false

Start-ThreadJob {
    param($killEvent, $url, $serverStopMessage, $serverSharedState)
    $killEvent.WaitOne()
    if ($serverSharedState.ListenerOpened) {
        Invoke-WebRequest -Uri $url -Body $serverStopMessage
    }
} -ArgumentList $killEvent, $localUrl, $serverStopMessage, $serverSharedState | Out-Null

try {
    Write-Host "$(Get-Date) - PowerShell HTTP server start listening on '$url'"
    $running = $true
    while ($running) {
        Start-Sleep  -Milliseconds 100
        $running = ($backgroundServer.State -eq 'Running')
        $backgroundServer = $backgroundServer | Get-Job
        $backgroundServer | Receive-Job
    }
}
finally {
    Write-Host "$(Get-Date) - PowerShell HTTP Server stop requested. Waiting for server to stop"
    $killEvent.Set() | Out-Null
    Get-Job | Wait-Job | Receive-Job
    Write-Host "$(Get-Date) - PowerShell HTTP server is stopped"
    [Environment]::Exit($serverSharedState.ExitCode)
}
