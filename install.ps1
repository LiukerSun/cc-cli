# CC-CLI thin installer for the Go rewrite

param(
    [ValidateSet("install", "uninstall")]
    [string]$Action = "install",
    [string]$Version = "",
    [string]$InstallBinDir = ""
)

$ErrorActionPreference = "Stop"

$RepoOwner = "LiukerSun"
$RepoName = "cc-cli"
$ProjectName = "ccc"
$RepoUrl = "https://github.com/$RepoOwner/$RepoName"

$ScriptPath = $MyInvocation.MyCommand.Path
if ([string]::IsNullOrWhiteSpace($ScriptPath) -and -not [string]::IsNullOrWhiteSpace($PSCommandPath)) {
    $ScriptPath = $PSCommandPath
}

$ScriptDir = ""
if (-not [string]::IsNullOrWhiteSpace($ScriptPath)) {
    $ScriptDir = Split-Path -Parent $ScriptPath
}

$LocalVersion = "dev"
if (-not [string]::IsNullOrWhiteSpace($ScriptDir)) {
    $VersionFile = Join-Path $ScriptDir "VERSION"
    if (Test-Path $VersionFile) {
        $LocalVersion = (Get-Content $VersionFile -Raw).Trim()
    }
}

if (-not $Version) {
    if ($env:CCC_VERSION) {
        $Version = $env:CCC_VERSION
    } else {
        $Version = "latest"
    }
}

if (-not $InstallBinDir) {
    if ($env:CCC_INSTALL_BIN_DIR) {
        $InstallBinDir = $env:CCC_INSTALL_BIN_DIR
    } else {
        $InstallBinDir = Join-Path $env:LOCALAPPDATA "Programs\ccc\bin"
    }
}

$InstallPath = Join-Path $InstallBinDir "ccc.exe"

function Write-Info {
    param([string]$Message)
    Write-Host $Message
}

function Fail {
    param([string]$Message)
    throw $Message
}

function Get-TagName {
    param([string]$RawVersion)

    if ($RawVersion -eq "latest") {
        return "latest"
    }
    if ($RawVersion.StartsWith("v")) {
        return $RawVersion
    }
    return "v$RawVersion"
}

function Get-OsArch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64" { $goArch = "amd64" }
        "Arm64" { $goArch = "arm64" }
        default { Fail "Unsupported architecture: $arch" }
    }

    return @{
        Os = "windows"
        Arch = $goArch
    }
}

function Expand-ZipFile {
    param(
        [string]$ZipPath,
        [string]$Destination
    )

    if (Test-Path $Destination) {
        Remove-Item -Recurse -Force $Destination
    }
    Expand-Archive -Path $ZipPath -DestinationPath $Destination -Force
}

function Verify-ChecksumIfAvailable {
    param(
        [string]$AssetPath,
        [string]$ChecksumsPath
    )

    if (-not (Test-Path $ChecksumsPath)) {
        return
    }

    $assetName = Split-Path -Leaf $AssetPath
    $line = Get-Content $ChecksumsPath | Where-Object { $_ -match " $([regex]::Escape($assetName))$" } | Select-Object -First 1
    if (-not $line) {
        return
    }

    $expected = ($line -split '\s+')[0]
    $actual = (Get-FileHash -Algorithm SHA256 -Path $AssetPath).Hash.ToLowerInvariant()
    if ($expected.ToLowerInvariant() -ne $actual) {
        Fail "Checksum verification failed for $assetName"
    }
}

function Install-FromLocalCheckout {
    if ([string]::IsNullOrWhiteSpace($ScriptDir)) {
        Fail "local checkout build requires a script path; rerun install.ps1 from a checked-out repository"
    }
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Fail "go is required to build from a local checkout"
    }

    $tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("ccc-install-" + [guid]::NewGuid().ToString("N"))
    New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null
    try {
        Write-Info "Building $ProjectName from local checkout..."
        Push-Location $ScriptDir
        try {
            & go build -trimpath -ldflags "-X github.com/LiukerSun/cc-cli/internal/buildinfo.Version=$LocalVersion" -o (Join-Path $tmpDir "ccc.exe") ./cmd/ccc
        } finally {
            Pop-Location
        }

        New-Item -ItemType Directory -Force -Path $InstallBinDir | Out-Null
        Copy-Item -Force (Join-Path $tmpDir "ccc.exe") $InstallPath
    } finally {
        if (Test-Path $tmpDir) {
            Remove-Item -Recurse -Force $tmpDir
        }
    }
}

function Install-FromRelease {
    $osArch = Get-OsArch
    $tag = Get-TagName -RawVersion $Version
    $assetName = "ccc_{0}_{1}.zip" -f $osArch.Os, $osArch.Arch

    if ($tag -eq "latest") {
        $baseUrl = "$RepoUrl/releases/latest/download"
    } else {
        $baseUrl = "$RepoUrl/releases/download/$tag"
    }

    $tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("ccc-install-" + [guid]::NewGuid().ToString("N"))
    New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null
    try {
        $archivePath = Join-Path $tmpDir $assetName
        $checksumsPath = Join-Path $tmpDir "checksums.txt"

        Write-Info "Downloading $assetName..."
        Invoke-WebRequest -Uri "$baseUrl/$assetName" -OutFile $archivePath
        try {
            Invoke-WebRequest -Uri "$baseUrl/checksums.txt" -OutFile $checksumsPath
        } catch {
        }
        Verify-ChecksumIfAvailable -AssetPath $archivePath -ChecksumsPath $checksumsPath

        $extractDir = Join-Path $tmpDir "extract"
        Expand-ZipFile -ZipPath $archivePath -Destination $extractDir
        $binaryPath = Join-Path $extractDir "ccc.exe"
        if (-not (Test-Path $binaryPath)) {
            Fail "Release archive did not contain ccc.exe"
        }

        New-Item -ItemType Directory -Force -Path $InstallBinDir | Out-Null
        Copy-Item -Force $binaryPath $InstallPath
    } finally {
        if (Test-Path $tmpDir) {
            Remove-Item -Recurse -Force $tmpDir
        }
    }
}

function Uninstall-Ccc {
    $paths = @(
        $InstallPath,
        (Join-Path $env:USERPROFILE "bin\ccc.exe"),
        (Join-Path $env:USERPROFILE ".ccc\bin\ccc.exe"),
        (Join-Path $env:USERPROFILE "bin\ccc.ps1")
    )

    $removed = $false
    foreach ($path in $paths) {
        if (Test-Path $path) {
            Remove-Item -Force $path
            Write-Info "Removed $path"
            $removed = $true
        }
    }

    if (-not $removed) {
        Write-Info "$ProjectName is not installed in the known locations."
    }

    Write-Info "Config and data were left untouched."
}

function Get-ProfileCandidates {
    return @(
        (Join-Path $env:USERPROFILE "Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"),
        (Join-Path $env:USERPROFILE "Documents\PowerShell\Microsoft.PowerShell_profile.ps1")
    ) | Select-Object -Unique
}

function Find-LegacyProfileReferences {
    $legacyShimPath = Join-Path $env:USERPROFILE "bin\ccc.ps1"
    $matches = @()

    foreach ($profilePath in (Get-ProfileCandidates)) {
        if (-not (Test-Path $profilePath)) {
            continue
        }

        $content = Get-Content -Path $profilePath -Raw -ErrorAction SilentlyContinue
        if ([string]::IsNullOrWhiteSpace($content)) {
            continue
        }

        if ($content.Contains($legacyShimPath) -or $content.Contains("bin\ccc.ps1")) {
            $matches += $profilePath
        }
    }

    return $matches
}

function Install-LegacyProfileShim {
    $legacyShimPath = Join-Path $env:USERPROFILE "bin\ccc.ps1"
    $legacyShimDir = Split-Path -Parent $legacyShimPath

    New-Item -ItemType Directory -Force -Path $legacyShimDir | Out-Null

    @"
# Auto-generated compatibility shim for legacy PowerShell profiles.
Write-Warning "Using legacy ccc PowerShell shim. Update your profile to call ccc.exe directly or add $InstallBinDir to PATH."
& "$InstallPath" @args
exit `$LASTEXITCODE
"@ | Set-Content -Path $legacyShimPath -Encoding UTF8

    return $legacyShimPath
}

if ($Action -eq "uninstall") {
    Uninstall-Ccc
    exit 0
}

$shouldUseLocalBuild = (-not [string]::IsNullOrWhiteSpace($ScriptDir)) -and (Test-Path (Join-Path $ScriptDir "go.mod")) -and (Test-Path (Join-Path $ScriptDir "cmd\ccc")) -and (($Version -eq "latest") -or ($Version -eq $LocalVersion) -or ($Version -eq "v$LocalVersion"))
if ($shouldUseLocalBuild) {
    Install-FromLocalCheckout
} else {
    Install-FromRelease
}

$legacyProfileRefs = Find-LegacyProfileReferences
$legacyShimPath = $null
if ($legacyProfileRefs.Count -gt 0) {
    $legacyShimPath = Install-LegacyProfileShim
}

Write-Info ""
Write-Info "$ProjectName installed to $InstallPath"
if (-not (($env:PATH -split ';') -contains $InstallBinDir)) {
    Write-Info "Warning: $InstallBinDir is not in PATH."
}
if ($legacyShimPath) {
    Write-Info "Warning: detected a legacy ccc PowerShell profile entry."
    foreach ($profilePath in $legacyProfileRefs) {
        Write-Info "  $profilePath"
    }
    Write-Info "Installed compatibility shim to $legacyShimPath"
    Write-Info "Update your profile to stop calling $legacyShimPath and prefer $InstallPath or PATH lookup."
}
Write-Info "Run 'ccc version' to verify the installation."
