function Start-ServerProcess {
    param(
        [Parameter(Mandatory=$true)]
        [ValidateScript({Test-Path $_})]
        $HandlerPath
    )

    Get-Content $HandlerPath -Raw | Set-Content "./handler.ps1"
    $powershellProcessName = (Get-Process -Id $pid).ProcessName

    $serverScriptPath = (Join-Path ($PSScriptRoot | Split-Path) 'server.ps1')

    $script:serverProcess = Start-Process `
        -FilePath $powershellProcessName `
        -ArgumentList @('-c', "$serverScriptPath") `
        -PassThru

}



function Wait-ServerExit {
    $MAX_PROCESS_RUTIME_SECONDS = 60

    $serverProcessRuntimeMs = 0
    $iterationTimeoutMs = 300
    while (-not $script:serverProcess.HasExited -and `
           ($serverProcessRuntimeMs / 1000) -lt $MAX_PROCESS_RUTIME_SECONDS) {
        Start-Sleep -Milliseconds 300
        $serverProcessRuntimeMs += 300
    }

    if (($serverProcessRuntimeMs / 1000) -gt $MAX_PROCESS_RUTIME_SECONDS ) {
        throw "Server process time out"
    }

    # return server exit code    
    $script:serverProcess.ExitCode
}