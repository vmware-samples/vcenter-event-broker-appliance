Describe "server.ps1 tests" {
    Context "Termination on error" {
        BeforeAll {
            $env:UNIT_TEST_SERVER = "http://localhost:52474/"

            . (Join-Path $PSScriptRoot 'server-test-helper.ps1')

            # set current folder to ensure handler.ps1 is loded by the server.ps1
            Push-Location $PSScriptRoot
        }

        AfterAll {
            if (Test-Path "./handler.ps1") {
                Remove-Item "./handler.ps1" -Confirm:$false
            }
            Pop-Location

            $env:PORT = $null
        }

        It "Should handle error in Process-Init and exit the server process with exit code 1" {
            # Arrange
            Start-ServerProcess -HandlerPath (Join-Path $PSScriptRoot 'process-init-error-handler.ps1')

            # Act
            $exitCode = Wait-ServerExit

            # Assert
            $exitCode | Should -Be 1
        }
    }
}