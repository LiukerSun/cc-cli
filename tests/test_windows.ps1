$ErrorActionPreference = "Stop"

$RepoDir = Split-Path -Parent $PSScriptRoot
$TempRoot = Join-Path $env:TEMP ("ccc-test-" + [guid]::NewGuid().ToString("N"))
$HomeDir = Join-Path $TempRoot "home"
$LocalAppData = Join-Path $TempRoot "local"
$AppData = Join-Path $TempRoot "roaming"
$BinDir = Join-Path $LocalAppData "Programs\ccc\bin"
$ConfigFile = Join-Path $AppData "ccc\config.json"
$CccBin = Join-Path $BinDir "ccc.exe"
$GoCache = Join-Path $TempRoot "go-cache"
$GoModCache = Join-Path $TempRoot "go-mod-cache"

function Assert-Contains {
    param(
        [string]$Text,
        [string]$Needle,
        [string]$Message
    )

    if ($Text -notlike "*$Needle*") {
        throw "$Message`nOutput:`n$Text"
    }
}

function Invoke-Ccc {
    param(
        [string[]]$CommandArgs,
        [string]$FailureMessage
    )

    $output = & $script:CccBin @CommandArgs 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "$FailureMessage`n$($output | Out-String)"
    }
    return ($output | Out-String)
}

New-Item -ItemType Directory -Force -Path $HomeDir | Out-Null
New-Item -ItemType Directory -Force -Path $LocalAppData | Out-Null
New-Item -ItemType Directory -Force -Path $AppData | Out-Null
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

$previousHome = $env:HOME
$previousUserProfile = $env:USERPROFILE
$previousLocalAppData = $env:LOCALAPPDATA
$previousAppData = $env:APPDATA
$previousXdgConfigHome = $env:XDG_CONFIG_HOME
$previousXdgDataHome = $env:XDG_DATA_HOME
$previousXdgCacheHome = $env:XDG_CACHE_HOME
$previousXdgStateHome = $env:XDG_STATE_HOME
$previousGoCache = $env:GOCACHE
$previousGoModCache = $env:GOMODCACHE
$previousPath = $env:PATH

try {
    $env:HOME = $HomeDir
    $env:USERPROFILE = $HomeDir
    $env:LOCALAPPDATA = $LocalAppData
    $env:APPDATA = $AppData
    $env:XDG_CONFIG_HOME = Join-Path $HomeDir ".config"
    $env:XDG_DATA_HOME = Join-Path $HomeDir ".local\share"
    $env:XDG_CACHE_HOME = Join-Path $HomeDir ".cache"
    $env:XDG_STATE_HOME = Join-Path $HomeDir ".local\state"
    $env:GOCACHE = $GoCache
    $env:GOMODCACHE = $GoModCache
    $env:PATH = "$BinDir;$previousPath"

    Push-Location $RepoDir
    try {
        & go build -o $CccBin ./cmd/ccc
        if ($LASTEXITCODE -ne 0) {
            throw "failed to build ccc.exe"
        }
    } finally {
        Pop-Location
    }

    $helpOutput = Invoke-Ccc -CommandArgs @("help") -FailureMessage "expected help to succeed"
    Assert-Contains -Text $helpOutput -Needle "ccc add <preset> <api-key> [model]" -Message "expected help output to include add shortcut"

    $profileAddOutput = Invoke-Ccc -CommandArgs @(
        "add", "anthropic", "test-key", "test-model",
        "--name", "Claude Test"
    ) -FailureMessage "expected profile add to succeed"
    Assert-Contains -Text $profileAddOutput -Needle "Added profile" -Message "expected profile add output to confirm success"

    if (-not (Test-Path $ConfigFile)) {
        throw "expected config file to be created at $ConfigFile"
    }

    $profileListOutput = Invoke-Ccc -CommandArgs @("profile", "list") -FailureMessage "expected profile list to succeed"
    Assert-Contains -Text $profileListOutput -Needle "claude-test" -Message "expected profile list to include generated profile id"

    $currentOutput = Invoke-Ccc -CommandArgs @("current") -FailureMessage "expected current to succeed"
    Assert-Contains -Text $currentOutput -Needle "Name: Claude Test" -Message "expected current profile output to include Claude Test"

    $profileUpdateOutput = Invoke-Ccc -CommandArgs @(
        "profile", "update", "claude-test",
        "--name", "Claude Prod",
        "--id", "claude-prod",
        "--env", "FOO=bar",
        "--no-sync"
    ) -FailureMessage "expected profile update to succeed"
    Assert-Contains -Text $profileUpdateOutput -Needle "Updated profile" -Message "expected profile update output to confirm success"

    $updatedCurrentOutput = Invoke-Ccc -CommandArgs @("current") -FailureMessage "expected current after update to succeed"
    Assert-Contains -Text $updatedCurrentOutput -Needle "ID: claude-prod" -Message "expected current profile output to include updated id"

    $dryRunOutput = Invoke-Ccc -CommandArgs @("run", "--dry-run") -FailureMessage "expected run --dry-run to succeed"
    Assert-Contains -Text $dryRunOutput -Needle "Command: claude" -Message "expected dry-run output to target claude"

    $configShowOutput = Invoke-Ccc -CommandArgs @("config", "show") -FailureMessage "expected config show to succeed"
    Assert-Contains -Text $configShowOutput -Needle '"current_profile": "claude-prod"' -Message "expected config show output to include current profile"
    Assert-Contains -Text $configShowOutput -Needle '"FOO": "bar"' -Message "expected config show output to include updated env"

    $upgradeOutput = Invoke-Ccc -CommandArgs @("upgrade", "--version", "2.2.1", "--dry-run") -FailureMessage "expected upgrade --dry-run to succeed"
    Assert-Contains -Text $upgradeOutput -Needle "Target version: 2.2.1" -Message "expected upgrade --dry-run output to include target version"

    $wrapperOutput = & powershell -ExecutionPolicy Bypass -File (Join-Path $RepoDir "bin/ccc.ps1") help 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "expected legacy PowerShell wrapper to delegate to installed Go binary`n$($wrapperOutput | Out-String)"
    }

    $wrapperText = $wrapperOutput | Out-String
    Assert-Contains -Text $wrapperText -Needle "ccc add <preset> <api-key> [model]" -Message "expected wrapper output to come from Go CLI help"
    Assert-Contains -Text $wrapperText -Needle "legacy compatibility wrapper" -Message "expected wrapper to print a deprecation warning"
} finally {
    $env:HOME = $previousHome
    $env:USERPROFILE = $previousUserProfile
    $env:LOCALAPPDATA = $previousLocalAppData
    $env:APPDATA = $previousAppData
    $env:XDG_CONFIG_HOME = $previousXdgConfigHome
    $env:XDG_DATA_HOME = $previousXdgDataHome
    $env:XDG_CACHE_HOME = $previousXdgCacheHome
    $env:XDG_STATE_HOME = $previousXdgStateHome
    $env:GOCACHE = $previousGoCache
    $env:GOMODCACHE = $previousGoModCache
    $env:PATH = $previousPath

    if (Test-Path $TempRoot) {
        Remove-Item -Recurse -Force $TempRoot
    }
}

Write-Host "test_windows.ps1: ok"
