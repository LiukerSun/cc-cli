$ErrorActionPreference = "Stop"

$RepoDir = Split-Path -Parent $PSScriptRoot
$TempRoot = Join-Path $env:TEMP ("ccc-test-" + [guid]::NewGuid().ToString("N"))
$HomeDir = Join-Path $TempRoot "home"
$LocalAppData = Join-Path $TempRoot "local"
$AppData = Join-Path $TempRoot "roaming"
$InstallBinDir = Join-Path $LocalAppData "Programs\ccc\bin"
$InstallPath = Join-Path $InstallBinDir "ccc.exe"

New-Item -ItemType Directory -Force -Path $HomeDir | Out-Null
New-Item -ItemType Directory -Force -Path $LocalAppData | Out-Null
New-Item -ItemType Directory -Force -Path $AppData | Out-Null

$previousUserProfile = $env:USERPROFILE
$previousLocalAppData = $env:LOCALAPPDATA
$previousAppData = $env:APPDATA

try {
    $env:USERPROFILE = $HomeDir
    $env:LOCALAPPDATA = $LocalAppData
    $env:APPDATA = $AppData

    & powershell -ExecutionPolicy Bypass -File (Join-Path $RepoDir "install.ps1")

    if (-not (Test-Path $InstallPath)) {
        throw "expected ccc.exe to be installed at $InstallPath"
    }

    $versionOutput = & $InstallPath version
    if ($versionOutput -notmatch "ccc version") {
        throw "expected installed ccc.exe to print version information"
    }

    & powershell -ExecutionPolicy Bypass -File (Join-Path $RepoDir "install.ps1") -Action uninstall

    if (Test-Path $InstallPath) {
        throw "expected uninstall to remove $InstallPath"
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
