$ErrorActionPreference = "Stop"

$RepoDir = Split-Path -Parent $PSScriptRoot
$TempRoot = Join-Path $env:TEMP ("ccc-test-" + [guid]::NewGuid().ToString("N"))
$HomeDir = Join-Path $TempRoot "home"
$LocalAppData = Join-Path $TempRoot "local"
$AppData = Join-Path $TempRoot "roaming"
$InstallBinDir = Join-Path $LocalAppData "Programs\ccc\bin"
$InstallPath = Join-Path $InstallBinDir "ccc.exe"
$LegacyShimPath = Join-Path $HomeDir "bin\ccc.ps1"
$LegacyProfilePath = Join-Path $HomeDir "Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"

function Assert-Contains {
    param(
        [string]$Text,
        [string]$Needle,
        [string]$Message
    )

    if (-not $Text.Contains($Needle)) {
        throw "$Message`nOutput:`n$Text"
    }
}

New-Item -ItemType Directory -Force -Path $HomeDir | Out-Null
New-Item -ItemType Directory -Force -Path $LocalAppData | Out-Null
New-Item -ItemType Directory -Force -Path $AppData | Out-Null
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $LegacyProfilePath) | Out-Null
@'
function ccc {
    & "$env:USERPROFILE\bin\ccc.ps1" @args
}
'@ | Set-Content -Path $LegacyProfilePath -Encoding UTF8

$previousUserProfile = $env:USERPROFILE
$previousLocalAppData = $env:LOCALAPPDATA
$previousAppData = $env:APPDATA

try {
    $env:USERPROFILE = $HomeDir
    $env:LOCALAPPDATA = $LocalAppData
    $env:APPDATA = $AppData

    $installScript = Get-Content (Join-Path $RepoDir "install.ps1") -Raw
    [scriptblock]::Create($installScript).Invoke()

    if (-not (Test-Path $InstallPath)) {
        throw "expected piped install.ps1 invocation to install ccc.exe at $InstallPath"
    }

    Remove-Item -Force $InstallPath

    $installOutput = & powershell -ExecutionPolicy Bypass -File (Join-Path $RepoDir "install.ps1") *>&1
    $installText = $installOutput | Out-String

    if (-not (Test-Path $InstallPath)) {
        throw "expected ccc.exe to be installed at $InstallPath"
    }
    if (-not (Test-Path $LegacyShimPath)) {
        throw "expected compatibility shim to be installed at $LegacyShimPath"
    }
    Assert-Contains -Text $installText -Needle "Installed compatibility shim" -Message "expected install output to mention compatibility shim"

    $versionOutput = & $InstallPath version
    if ($versionOutput -notmatch "ccc version") {
        throw "expected installed ccc.exe to print version information"
    }

    $shimOutput = & powershell -ExecutionPolicy Bypass -File $LegacyShimPath version 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "expected compatibility shim to delegate to ccc.exe`n$($shimOutput | Out-String)"
    }

    & powershell -ExecutionPolicy Bypass -File (Join-Path $RepoDir "install.ps1") -Action uninstall

    if (Test-Path $InstallPath) {
        throw "expected uninstall to remove $InstallPath"
    }
    if (Test-Path $LegacyShimPath) {
        throw "expected uninstall to remove $LegacyShimPath"
    }
} finally {
    $env:USERPROFILE = $previousUserProfile
    $env:LOCALAPPDATA = $previousLocalAppData
    $env:APPDATA = $previousAppData

    if (Test-Path $TempRoot) {
        Remove-Item -Recurse -Force $TempRoot
    }
}

Write-Host "install_windows.ps1: ok"
