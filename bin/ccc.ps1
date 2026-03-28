# Legacy compatibility wrapper for the Go-based ccc binary.

$ErrorActionPreference = "Stop"

$ScriptPath = $MyInvocation.MyCommand.Path
$ScriptDir = Split-Path -Parent $ScriptPath
$RepoRoot = Split-Path -Parent $ScriptDir

function Write-LegacyWarning {
    Write-Warning "bin/ccc.ps1 is now a legacy compatibility wrapper. Install and use the Go binary instead."
}

function Get-TargetBinary {
    $candidates = @(
        (Join-Path $env:LOCALAPPDATA "Programs\ccc\bin\ccc.exe"),
        (Join-Path $env:USERPROFILE "bin\ccc.exe"),
        (Join-Path $RepoRoot "ccc.exe")
    )

    foreach ($candidate in $candidates) {
        if ((Test-Path $candidate) -and ($candidate -ne $ScriptPath)) {
            return $candidate
        }
    }

    return $null
}

$target = Get-TargetBinary
if ($target) {
    Write-LegacyWarning
    & $target @args
    exit $LASTEXITCODE
}

Write-Error @"
No installed Go-based ccc binary was found.

Install cc-cli:
  irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex

Or build it locally from this repository:
  go build -o $env:LOCALAPPDATA\Programs\ccc\bin\ccc.exe ./cmd/ccc
"@
exit 1
