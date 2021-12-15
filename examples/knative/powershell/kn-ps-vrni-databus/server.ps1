if (${env:PORT}) {
   $url = "http://*:${env:PORT}/"
   $localUrl = "http://localhost:${env:PORT}/"
}
else {
   $url = "http://*:8080/"
   $localUrl = "http://localhost:8080/"
}

$serverStopMessage = 'break-signal-e2db683c-b8ff-4c4f-8158-c44f734e2bf1'

. ./handler.ps1

$backgroundServer = Start-ThreadJob {
   param($url, $serverStopMessage)

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
         $Url
      )

      # create a HTTP listener that will listen for connections on *:8080
      $listener = New-Object -Type 'System.Net.HttpListener'
      $listener.AuthenticationSchemes = [System.Net.AuthenticationSchemes]::Anonymous
      $listener.Prefixes.Add($Url)

      try {
         # start the HTTP server and run continuously, until the loop says to stop
         $listener.Start()

         while ($true) {
            $context = $listener.GetContext()

            try {
               # read the input stream (the POST body)
               $bodyData = [System.IO.StreamReader]::new($context.Request.InputStream).ReadToEnd()

               # Read Headers
               $headers = @{}
               for ($i = 0; $i -lt $context.Request.Headers.Count; $i++) {
                  $headers[$context.Request.Headers.GetKey($i)] = $context.Request.Headers.GetValues($i)
               }

               # see if there's a X-Forwarded-For header to use as a HTTP request source
               $httpSource = ""
               try {
                  $httpSource = ((${headers}.'X-Forwarded-For') -split ',')[0]

                  if ($httpSource -eq "") {
                     Write-Host "$(Get-Date) - Unable to get remote client, setting CloudEvent source to function hostname"
                     $httpSource = [System.Net.Dns]::GetHostName()
                  }
               }
               catch {
                  throw "`nFailed to set source for CloudEvent"
               }

               # if the body is the stop message, gently shut down the HTTP server
               if ($bodyData -eq $serverStopMessage) {
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

               if (${env:SERVICE_DEBUG} -eq "true") {
                  # Only used for debugging and printing to console
                  Write-Host "$(Get-Date) - [Headers]:"
                  Write-Host $($Headers | Out-String)
                  Write-Host "$(Get-Date) - [Body]:"
                  Write-Host
                  Write-Host $($bodyData)
                  Write-Host
               }

               try {
                  # run the process-handler that's defined in ./handler.ps1
                  Process-Handler -Source $httpSource -Body $bodyData -ErrorAction 'Stop'
                  $context.Response.StatusCode = [int]([System.Net.HttpStatusCode]::OK)
               }
               catch [System.Net.Http.HttpRequestException] {
                  Write-Error "$(Get-Date) - Handler Processing Error: $($_.Exception.ToString())"

                  if ($_.Exception.StatusCode -eq "BadRequest") {
                     $context.Response.StatusCode = [int]([System.Net.HttpStatusCode]::BadRequest)
                     $context.Response.Close();
                  }
               }
               catch {
                  Write-Error "$(Get-Date) - Handler Processing Error: $($_.Exception.ToString())"
                  $context.Response.StatusCode = [int]([System.Net.HttpStatusCode]::InternalServerError)
                  $context.Response.Close();
               }
            }
            catch {
               Write-Error "`n$(Get-Date) - HTTP Request Processing Error: $($_.Exception.ToString())"
               $context.Response.StatusCode = [int]([System.Net.HttpStatusCode]::InternalServerError)
               $context.Response.Close();
            }
            finally {
               $context.Response.Close();
            }
         }
      }
      catch {
         Write-Error "$(Get-Date) - Listener Processing Error: $($_.Exception.ToString())"
         exit 1
      }
      finally {
         $listener.Stop()
      }
   }


   # Runs Init function (defined in handler.ps1) which can be used to enable warm startup
   try {
      Process-Init -ErrorAction 'Stop'
   }
   catch {
      Write-Error "$(Get-Date) - Init Processing Error: $($_.Exception.ToString())"
      exit 1
   }

   $breakSignal = Start-HttpCloudEventListener -Url $url
   if ($breakSignal) {
      Write-Host "$(Get-Date) - PowerShell HTTP server stop requested"
      break;
   }

} -ArgumentList $url, $serverStopMessage

$killEvent = new-object 'System.Threading.AutoResetEvent' -ArgumentList $false

Start-ThreadJob {
   param($killEvent, $url, $serverStopMessage)
   $killEvent.WaitOne()
   Invoke-WebRequest -Uri $url -Body $serverStopMessage
} -ArgumentList $killEvent, $localUrl, $serverStopMessage

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
   $killEvent.Set()
   Get-Job | Wait-Job | Receive-Job
   Write-Host "$(Get-Date) - PowerShell HTTP server is stopped"
}