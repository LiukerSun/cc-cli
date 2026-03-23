# CC-CLI PowerShell Installation Script
# https://github.com/LiukerSun/cc-cli

param(
    [string]$Action = "install",
    [string]$Branch = "main"
)

$scriptPath = $MyInvocation.MyCommand.Path
$REPO_URL = "https://github.com/LiukerSun/cc-cli"

if ($scriptPath) {
    $SCRIPT_DIR = Split-Path -Parent $scriptPath
    $VERSION_FILE = Join-Path $SCRIPT_DIR "VERSION"
    if (Test-Path $VERSION_FILE) {
        $VERSION = (Get-Content $VERSION_FILE -Raw).Trim()
    } else {
        $VERSION = "unknown"
    }
} else {
    $SCRIPT_DIR = $null
    try {
        $versionUrl = "$REPO_URL/raw/$Branch/VERSION"
        $VERSION = (New-Object System.Net.WebClient).DownloadString($versionUrl).Trim()
    } catch {
        $VERSION = "unknown"
    }
}

# Installation paths
$INSTALL_DIR = "$env:USERPROFILE\.cc-cli"
$CONFIG_FILE = "$env:USERPROFILE\.cc-config.json"
$SCRIPT_FILE = "$env:USERPROFILE\bin\ccc.ps1"
$CLI_PACKAGES = @{
    claude = "@anthropic-ai/claude-code"
    codex = "@openai/codex"
}
$CLI_MIN_NODE_VERSIONS = @{
    claude = "18.0.0"
    codex = "16.0.0"
}

# Colors
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Success { Write-ColorOutput Green $args }
function Write-Info { Write-ColorOutput Cyan $args }
function Write-Warning { Write-ColorOutput Yellow $args }
function Write-Error { Write-ColorOutput Red $args }

# Helper function to save UTF-8 without BOM
function Save-FileNoBOM {
    param(
        [string]$Path,
        [string]$Content
    )
    
    $utf8NoBom = New-Object System.Text.UTF8Encoding $false
    [System.IO.File]::WriteAllText($Path, $Content, $utf8NoBom)
}

function Get-CliPackageName {
    param(
        [string]$CommandName
    )

    return $CLI_PACKAGES[$CommandName]
}

function Get-CliMinimumNodeVersion {
    param(
        [string]$CommandName
    )

    return $CLI_MIN_NODE_VERSIONS[$CommandName]
}

function ConvertTo-SemanticVersionString {
    param(
        [string]$VersionString
    )

    $clean = ([string]$VersionString).Trim()
    $clean = $clean.TrimStart('v')
    $clean = $clean -replace '^>=\s*', ''
    $clean = $clean -replace '[-+].*$', ''

    $parts = @($clean.Split('.', [System.StringSplitOptions]::RemoveEmptyEntries))
    while ($parts.Count -lt 3) {
        $parts += "0"
    }

    return "$($parts[0]).$($parts[1]).$($parts[2])"
}

function Test-VersionLessThan {
    param(
        [string]$Left,
        [string]$Right
    )

    return ([Version](ConvertTo-SemanticVersionString -VersionString $Left)) -lt ([Version](ConvertTo-SemanticVersionString -VersionString $Right))
}

function Add-NpmGlobalBinToPath {
    if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
        return
    }

    $npmPrefix = (& npm config get prefix 2>$null | Out-String).Trim()
    if (-not $npmPrefix -or $npmPrefix -eq "undefined") {
        return
    }

    $candidate = $npmPrefix
    if ($env:OS -ne "Windows_NT") {
        $unixBin = Join-Path $npmPrefix "bin"
        if (Test-Path $unixBin) {
            $candidate = $unixBin
        }
    }

    $pathEntries = @($env:PATH -split [System.IO.Path]::PathSeparator)
    if ($pathEntries -notcontains $candidate) {
        $env:PATH = "$candidate$([System.IO.Path]::PathSeparator)$env:PATH"
    }
}

function Require-NodeForInstall {
    if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
        Write-Error "[X] Node.js is required to install ccc"
        Write-Host "  Current version: not installed"
        Write-Host "  Please install Node.js first, then rerun the installer."
        return $false
    }

    Write-Success "[OK] Node.js $((& node --version 2>$null | Out-String).Trim()) found"
    return $true
}

function Ensure-NodeAndNpmForCommand {
    param(
        [string]$CommandName
    )

    $requiredVersion = Get-CliMinimumNodeVersion -CommandName $CommandName
    if (-not $requiredVersion) {
        Write-Warning "[!] Unsupported CLI command: $CommandName"
        return $false
    }

    if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
        Write-Warning "[!] Skipping automatic install for $CommandName: Node.js is not installed"
        Write-Host "  Current version: not installed"
        Write-Host "  Minimum required version: Node.js >= $requiredVersion"
        Write-Host "  ccc is installed anyway. Install or upgrade Node.js before using '$CommandName'."
        return $false
    }

    if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
        $currentNodeVersion = (& node --version 2>$null | Out-String).Trim()
        Write-Warning "[!] Skipping automatic install for $CommandName: npm is not installed"
        Write-Host "  Current Node.js version: $currentNodeVersion"
        Write-Host "  Minimum required version: Node.js >= $requiredVersion"
        Write-Host "  ccc is installed anyway. Install npm before using '$CommandName'."
        return $false
    }

    $currentNodeVersion = (& node --version 2>$null | Out-String).Trim()
    if (Test-VersionLessThan -Left $currentNodeVersion -Right $requiredVersion) {
        Write-Warning "[!] Skipping automatic install for $CommandName: Node.js version is too old"
        Write-Host "  Current version: $currentNodeVersion"
        Write-Host "  Minimum required version: Node.js >= $requiredVersion"
        Write-Host "  ccc is installed anyway. Upgrade Node.js before using '$CommandName'."
        return $false
    }

    Write-Success "[OK] Node.js $currentNodeVersion found"
    Write-Success "[OK] npm $((& npm --version 2>$null | Out-String).Trim()) found"
    Add-NpmGlobalBinToPath
    return $true
}

function Install-MissingCliCommand {
    param(
        [string]$CommandName
    )

    $packageName = Get-CliPackageName -CommandName $CommandName
    if (-not $packageName) {
        Write-Error "[X] Unsupported CLI command: $CommandName"
        return $false
    }

    Write-Warning "Installing missing $CommandName CLI via npm ($packageName)..."
    & npm install -g $packageName
    if ($LASTEXITCODE -ne 0) {
        Write-Error "[X] Failed to install $CommandName CLI automatically"
        Write-Host "  Try manually: npm install -g $packageName"
        return $false
    }

    Add-NpmGlobalBinToPath
    if (-not (Get-Command $CommandName -ErrorAction SilentlyContinue)) {
        Write-Error "[X] $CommandName CLI is still not available after installation"
        Write-Host "  Try manually: npm install -g $packageName"
        return $false
    }

    Write-Success "[OK] $CommandName CLI installed"
    return $true
}

function Check-AndPrepareCliCommands {
    Add-NpmGlobalBinToPath

    foreach ($commandName in @("claude", "codex")) {
        if (Get-Command $commandName -ErrorAction SilentlyContinue) {
            Write-Success "[OK] $commandName CLI found"
        } else {
            Write-Warning "[!] $commandName CLI not found"
            if (Ensure-NodeAndNpmForCommand -CommandName $commandName) {
                if (-not (Install-MissingCliCommand -CommandName $commandName)) {
                    Write-Warning "[!] Continuing installation without $commandName CLI"
                }
            }
        }
    }

    Write-Host ""
}

# Banner
Write-Info "==================================="
Write-Info "  CC-CLI Installer v$VERSION (PowerShell)"
Write-Info "==================================="
Write-Host ""

# Check requirements
function Check-Requirements {
    Write-Warning "Checking system requirements..."
    
    # Check PowerShell version
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Error "[X] PowerShell 5.0 or later is required"
        exit 1
    }
    Write-Success "[OK] PowerShell $($PSVersionTable.PSVersion) found"

    if (-not (Require-NodeForInstall)) {
        exit 1
    }
    
    Check-AndPrepareCliCommands
}

# Create directories
function Create-Directories {
    Write-Warning "Creating directories..."
    
    New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path $SCRIPT_FILE -Parent) | Out-Null
    
    Write-Success "[OK] Created $INSTALL_DIR"
    Write-Success "[OK] Created $(Split-Path $SCRIPT_FILE -Parent)"
    Write-Host ""
}

# Download main script
function Install-Script {
    Write-Warning "Installing ccc command (branch: $Branch)..."
    
    $scriptUrl = "$REPO_URL/raw/$Branch/bin/ccc.ps1"
    $installScriptDest = "$INSTALL_DIR\install.ps1"
    
    if ($SCRIPT_DIR) {
        $localScript = Join-Path $SCRIPT_DIR "bin\ccc.ps1"
        
        if (Test-Path $localScript) {
            Copy-Item -Path $localScript -Destination $SCRIPT_FILE -Force
            $localVersion = Join-Path $SCRIPT_DIR "VERSION"
            if (Test-Path $localVersion) {
                Copy-Item -Path $localVersion -Destination "$INSTALL_DIR\VERSION" -Force
            }
            # Copy install.ps1 for uninstall support
            $localInstallScript = Join-Path $SCRIPT_DIR "install.ps1"
            if (Test-Path $localInstallScript) {
                Copy-Item -Path $localInstallScript -Destination $installScriptDest -Force
                Write-Success "[OK] Installed uninstall script to $installScriptDest"
            }
            Write-Success "[OK] Installed ccc.ps1 from local source"

            # Migrate from old 'cc' command name
            $oldScript = Join-Path (Split-Path $SCRIPT_FILE -Parent) "cc.ps1"
            if (Test-Path $oldScript) {
                Remove-Item $oldScript -Force
                Write-Success "[OK] Removed old 'cc.ps1' (now use 'ccc')"
            }

            Write-Host ""
            return
        }
    }
    
    Write-Host "Downloading from: $scriptUrl"
    
    try {
        $webClient = New-Object System.Net.WebClient
        $webClient.Encoding = [System.Text.Encoding]::UTF8
        $content = $webClient.DownloadString($scriptUrl)
        
        $utf8NoBom = New-Object System.Text.UTF8Encoding $false
        [System.IO.File]::WriteAllText($SCRIPT_FILE, $content, $utf8NoBom)
        
        Write-Success "[OK] Downloaded ccc.ps1 to $SCRIPT_FILE"
        
        # Download install.ps1 for uninstall support
        $installScriptUrl = "$REPO_URL/raw/$Branch/install.ps1"
        try {
            $installScriptContent = $webClient.DownloadString($installScriptUrl)
            [System.IO.File]::WriteAllText($installScriptDest, $installScriptContent, $utf8NoBom)
            Write-Success "[OK] Downloaded uninstall script to $installScriptDest"
        } catch {
            Write-Warning "[!] Could not download install.ps1 file"
        }
        
        $versionDest = "$INSTALL_DIR\VERSION"
        $versionUrl = "$REPO_URL/raw/$Branch/VERSION"
        try {
            $versionContent = $webClient.DownloadString($versionUrl).Trim()
            [System.IO.File]::WriteAllText($versionDest, $versionContent, $utf8NoBom)
            Write-Success "[OK] Downloaded VERSION to $versionDest"
        } catch {
            Write-Warning "[!] Could not download VERSION file"
        }

        # Migrate from old 'cc' command name
        $oldScript = Join-Path (Split-Path $SCRIPT_FILE -Parent) "cc.ps1"
        if (Test-Path $oldScript) {
            Remove-Item $oldScript -Force
            Write-Success "[OK] Removed old 'cc.ps1' (now use 'ccc')"
        }
    } catch {
        Write-Error "[X] Failed to download script: $_"
        Write-Host "Please download manually from: $scriptUrl"
        exit 1
    }
    
    Write-Host ""
}

# Create default config
function Create-Config {
    if (-not (Test-Path $CONFIG_FILE)) {
        Write-Warning "Creating empty configuration..."
        
        Save-FileNoBOM -Path $CONFIG_FILE -Content "[]"
        Write-Success "[OK] Created empty config file: $CONFIG_FILE"
        Write-Warning "  Run 'ccc -a' to add your first model"
    } else {
        Write-Success "[OK] Config file already exists: $CONFIG_FILE"
    }
    Write-Host ""
}

# Add to PATH
function Add-ToPath {
    Write-Warning "Configuring PATH..."
    
    $binDir = Split-Path $SCRIPT_FILE -Parent
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    if ($currentPath -notlike "*$binDir*") {
        [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$binDir", "User")
        Write-Success "[OK] Added $binDir to PATH"
        Write-Warning "  Please restart your terminal or run: `$env:PATH += `";$binDir`""
    } else {
        Write-Success "[OK] Already in PATH"
    }
    Write-Host ""
}

# Create wrapper function
function Create-Wrapper {
    Write-Warning "Creating PowerShell wrapper..."
    
    $wrapperContent = @'
# CC-CLI PowerShell Wrapper
function ccc {
    & "$env:USERPROFILE\bin\ccc.ps1" @args
}
'@
    
    $profileDir = Split-Path $PROFILE -Parent
    if (-not (Test-Path $profileDir)) {
        New-Item -ItemType Directory -Force -Path $profileDir | Out-Null
    }
    
    # Remove old wrapper if exists
    if (Test-Path $PROFILE) {
        $profileContent = Get-Content $PROFILE -Raw
        $pattern = '(?s)# CC-CLI PowerShell Wrapper.*?function ccc \{.*?\}'
        $profileContent = $profileContent -replace $pattern, ''
        $profileContent = $profileContent.TrimEnd()
        
        # Save cleaned profile
        $utf8NoBom = New-Object System.Text.UTF8Encoding $false
        [System.IO.File]::WriteAllText($PROFILE, $profileContent, $utf8NoBom)
    }
    
    # Add new wrapper
    Add-Content -Path $PROFILE -Value "`n`n$wrapperContent"
    Write-Success "[OK] Added ccc function to PowerShell profile"
    Write-Warning "  Please restart PowerShell or run: . `$PROFILE"
    Write-Host ""
}

# Print success message
function Print-Success {
    Write-Success "==================================="
    Write-Success "  Installation Complete!"
    Write-Success "==================================="
    Write-Host ""
    
    Write-Info "Next steps:"
    Write-Host ""
    Write-Host "  1. " -NoNewline
    Write-Warning "Add your API keys:"
    Write-Host "     " -NoNewline
    Write-Info "ccc -E"
    Write-Host ""
    Write-Host "  2. " -NoNewline
    Write-Warning "Restart PowerShell or run:"
    Write-Host "     " -NoNewline
    Write-Info ". `$PROFILE"
    Write-Host ""
    Write-Host "  3. " -NoNewline
    Write-Warning "Start using ccc:"
    Write-Host "     " -NoNewline
    Write-Info "ccc              # Interactive selection"
    Write-Host "     " -NoNewline
    Write-Info "ccc --list       # List all models"
    Write-Host "     " -NoNewline
    Write-Info "ccc --help       # Show help"
    Write-Host ""
    Write-Info "Documentation:"
    Write-Host "  $REPO_URL"
    Write-Host ""
    Write-Info "Configuration file:"
    Write-Host "  $CONFIG_FILE"
    Write-Host ""
}

# Uninstall function
function Uninstall {
    param(
        [switch]$KeepConfig,
        [switch]$KeepSettings
    )
    
    Write-Info "==================================="
    Write-Info "  CC-CLI Uninstaller"
    Write-Info "==================================="
    Write-Host ""
    
    $confirm = Read-Host "Are you sure you want to uninstall cc-cli? (y/N)"
    if ($confirm -notin @("y", "Y", "yes", "YES")) {
        Write-Host "Uninstall cancelled."
        exit 0
    }
    
    Write-Host ""
    Write-Warning "Removing files..."
    
    # Remove main script
    if (Test-Path $SCRIPT_FILE) {
        Remove-Item $SCRIPT_FILE -Force
        Write-Success "[OK] Removed $SCRIPT_FILE"
    } else {
        Write-Warning "[!] Script file not found: $SCRIPT_FILE"
    }
    
    # Remove installation directory
    if (Test-Path $INSTALL_DIR) {
        Remove-Item $INSTALL_DIR -Recurse -Force
        Write-Success "[OK] Removed $INSTALL_DIR"
    } else {
        Write-Warning "[!] Installation directory not found: $INSTALL_DIR"
    }
    
    # Remove temp env file
    $ENV_FILE = "$env:TEMP\cc-model-env.ps1"
    if (Test-Path $ENV_FILE) {
        Remove-Item $ENV_FILE -Force
        Write-Success "[OK] Removed $ENV_FILE"
    }
    
    # Remove config file (optional)
    if (-not $KeepConfig) {
        $configConfirm = Read-Host "`nDelete config file? (y/N)"
        if ($configConfirm -in @("y", "Y", "yes", "YES")) {
            if (Test-Path $CONFIG_FILE) {
                Remove-Item $CONFIG_FILE -Force
                Write-Success "[OK] Removed $CONFIG_FILE"
            } else {
                Write-Warning "[!] Config file not found: $CONFIG_FILE"
            }
        } else {
            Write-Warning "[OK] Config file preserved: $CONFIG_FILE"
        }
    }
    
    # Remove Claude settings (optional)
    $CLAUDE_SETTINGS_FILE = "$env:USERPROFILE\.claude\settings.json"
    if (-not $KeepSettings) {
        if (Test-Path $CLAUDE_SETTINGS_FILE) {
            $settingsConfirm = Read-Host "`nRemove cc-cli entries from Claude settings? (y/N)"
            if ($settingsConfirm -in @("y", "Y", "yes", "YES")) {
                try {
                    $settings = Get-Content $CLAUDE_SETTINGS_FILE -Raw | ConvertFrom-Json
                    
                    # Remove cc-cli added env variables
                    if ($settings.env) {
                        $settings.env.PSObject.Properties.Remove('ANTHROPIC_MODEL')
                        $settings.env.PSObject.Properties.Remove('ANTHROPIC_SMALL_FAST_MODEL')
                        $settings.env.PSObject.Properties.Remove('CLAUDE_CODE_MODEL')
                        $settings.env.PSObject.Properties.Remove('CLAUDE_CODE_SMALL_MODEL')
                        $settings.env.PSObject.Properties.Remove('CLAUDE_CODE_SUBAGENT_MODEL')
                        
                        # Remove env section if empty
                        if ($settings.env.PSObject.Properties.Count -eq 0) {
                            $settings.PSObject.Properties.Remove('env')
                        }
                    }
                    
                    # Remove model field
                    if ($settings.model) {
                        $settings.PSObject.Properties.Remove('model')
                    }
                    
                    # Remove Agent(Explore) from deny list
                    if ($settings.permissions -and $settings.permissions.deny) {
                        $denyList = @($settings.permissions.deny) | Where-Object { $_ -ne "Agent(Explore)" }
                        if ($denyList.Count -eq 0) {
                            $settings.permissions.PSObject.Properties.Remove('deny')
                        } else {
                            $settings.permissions.deny = $denyList
                        }
                    }
                    
                    # Save cleaned settings
                    $utf8NoBom = New-Object System.Text.UTF8Encoding $false
                    $json = $settings | ConvertTo-Json -Depth 10
                    [System.IO.File]::WriteAllText($CLAUDE_SETTINGS_FILE, $json, $utf8NoBom)
                    
                    Write-Success "[OK] Cleaned $CLAUDE_SETTINGS_FILE"
                } catch {
                    Write-Warning "[!] Failed to clean settings file: $_"
                }
            } else {
                Write-Warning "[OK] Claude settings preserved: $CLAUDE_SETTINGS_FILE"
            }
        }
    }
    
    # Remove from PATH
    Write-Host ""
    Write-Warning "Removing from PATH..."
    $binDir = Split-Path $SCRIPT_FILE -Parent
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    if ($currentPath -like "*$binDir*") {
        $newPath = ($currentPath -split ';' | Where-Object { $_ -ne $binDir }) -join ';'
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Success "[OK] Removed $binDir from PATH"
        Write-Warning "  Please restart your terminal or run: `$env:PATH = `$env:PATH"
    } else {
        Write-Warning "[!] $binDir not found in PATH"
    }
    
    # Remove wrapper from profile
    Write-Host ""
    Write-Warning "Cleaning PowerShell profile..."
    if (Test-Path $PROFILE) {
        $profileContent = Get-Content $PROFILE -Raw
        $pattern = '(?s)# CC-CLI PowerShell Wrapper.*?function ccc \{.*?\}'
        $newContent = $profileContent -replace $pattern, ''
        $newContent = $newContent.TrimEnd()
        
        if ($newContent -ne $profileContent) {
            $utf8NoBom = New-Object System.Text.UTF8Encoding $false
            [System.IO.File]::WriteAllText($PROFILE, $newContent, $utf8NoBom)
            Write-Success "[OK] Removed ccc function from $PROFILE"
            Write-Warning "  Please restart PowerShell or run: . `$PROFILE"
        } else {
            Write-Warning "[!] No ccc function found in profile"
        }
    } else {
        Write-Warning "[!] PowerShell profile not found"
    }
    
    Write-Host ""
    Write-Success "==================================="
    Write-Success "  Uninstall Complete!"
    Write-Success "==================================="
    Write-Host ""
    
    exit 0
}

# Main
switch ($Action.ToLower()) {
    "uninstall" {
        [switch]$KeepConfig = $false
        [switch]$KeepSettings = $false
        
        for ($i = 0; $i -lt $args.Count; $i++) {
            switch ($args[$i]) {
                "--keep-config" { $KeepConfig = $true }
                "--keep-settings" { $KeepSettings = $true }
            }
        }
        
        Uninstall -KeepConfig:$KeepConfig -KeepSettings:$KeepSettings
    }
    "help" {
        Write-Host "Usage: .\install.ps1 [-Branch <name>] [ACTION]"
        Write-Host ""
        Write-Host "Actions:"
        Write-Host "  install     Install cc-cli (default)"
        Write-Host "  uninstall   Remove cc-cli"
        Write-Host "  help        Show this help message"
        Write-Host ""
        Write-Host "Options:"
        Write-Host "  -Branch     Specify branch to install (default: main)"
        Write-Host ""
        Write-Host "Uninstall Options:"
        Write-Host "  --keep-config      Preserve config file"
        Write-Host "  --keep-settings    Preserve Claude settings"
        Write-Host ""
        Write-Host "Examples:"
        Write-Host "  .\install.ps1"
        Write-Host "  .\install.ps1 -Branch feature/auto-fetch-zhipu-models"
        Write-Host "  .\install.ps1 -Action uninstall"
        Write-Host "  .\install.ps1 -Action uninstall --keep-config"
        exit 0
    }
    default {
        Check-Requirements
        Create-Directories
        Install-Script
        Create-Config
        Add-ToPath
        Create-Wrapper
        Print-Success
    }
}
