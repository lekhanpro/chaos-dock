param(
  [ValidateSet("init", "validate", "list", "run-once", "run-scheduled", "panic", "tui")]
  [string]$Action = "tui",
  [string]$Config = "chaos.yaml",
  [string]$Targets = "",
  [switch]$Force
)

$go = "C:\Users\lekhan hr\tools\go\bin\go.exe"
if (-not (Test-Path $go)) {
  Write-Error "Go executable not found at $go"
  exit 1
}

$cmd = @("run", "./cmd/chaos-dock")

switch ($Action) {
  "init" {
    $cmd += @("-init-config", "-config", $Config)
    if ($Force) {
      $cmd += "-force"
    }
  }
  "validate" {
    $cmd += @("-validate-config", "-config", $Config)
  }
  "list" {
    $cmd += "-list"
  }
  "run-once" {
    $cmd += @("-run-once", "-config", $Config)
  }
  "run-scheduled" {
    $cmd += @("-run-scheduled", "-config", $Config)
  }
  "panic" {
    $cmd += "-panic"
    if ($Targets -ne "") {
      $cmd += @("-targets", $Targets)
    }
  }
  "tui" {
    # no extra args
  }
}

Write-Host "Executing: go $($cmd -join ' ')"
& $go @cmd
exit $LASTEXITCODE

