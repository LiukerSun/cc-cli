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
$SCRIPT_FILE = "$env:USERPROFILE\bin\cc.ps1"

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
    
    # Check for claude command
    if (-not (Get-Command claude -ErrorAction SilentlyContinue)) {
        Write-Warning "[!] Claude CLI not found (optional)"
        Write-Host "  Install from: https://claude.ai"
    } else {
        Write-Success "[OK] Claude CLI found"
    }
    
    Write-Host ""
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
    Write-Warning "Installing cc command (branch: $Branch)..."
    
    $scriptUrl = "$REPO_URL/raw/$Branch/bin/cc.ps1"
    $installScriptDest = "$INSTALL_DIR\install.ps1"
    
    if ($SCRIPT_DIR) {
        $localScript = Join-Path $SCRIPT_DIR "bin\cc.ps1"
        
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
            Write-Success "[OK] Installed cc.ps1 from local source"
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
        
        Write-Success "[OK] Downloaded cc.ps1 to $SCRIPT_FILE"
        
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
        Write-Warning "  Run 'cc -a' to add your first model"
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
function cc {
    & "$env:USERPROFILE\bin\cc.ps1" @args
}
'@
    
    $profileDir = Split-Path $PROFILE -Parent
    if (-not (Test-Path $profileDir)) {
        New-Item -ItemType Directory -Force -Path $profileDir | Out-Null
    }
    
    # Remove old wrapper if exists
    if (Test-Path $PROFILE) {
        $profileContent = Get-Content $PROFILE -Raw
        $pattern = '(?s)# CC-CLI PowerShell Wrapper.*?function cc \{.*?\}'
        $profileContent = $profileContent -replace $pattern, ''
        $profileContent = $profileContent.TrimEnd()
        
        # Save cleaned profile
        $utf8NoBom = New-Object System.Text.UTF8Encoding $false
        [System.IO.File]::WriteAllText($PROFILE, $profileContent, $utf8NoBom)
    }
    
    # Add new wrapper
    Add-Content -Path $PROFILE -Value "`n`n$wrapperContent"
    Write-Success "[OK] Added cc function to PowerShell profile"
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
    Write-Info "cc -E"
    Write-Host ""
    Write-Host "  2. " -NoNewline
    Write-Warning "Restart PowerShell or run:"
    Write-Host "     " -NoNewline
    Write-Info ". `$PROFILE"
    Write-Host ""
    Write-Host "  3. " -NoNewline
    Write-Warning "Start using cc:"
    Write-Host "     " -NoNewline
    Write-Info "cc              # Interactive selection"
    Write-Host "     " -NoNewline
    Write-Info "cc --list       # List all models"
    Write-Host "     " -NoNewline
    Write-Info "cc --help       # Show help"
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
        $pattern = '(?s)# CC-CLI PowerShell Wrapper.*?function cc \{.*?\}'
        $newContent = $profileContent -replace $pattern, ''
        $newContent = $newContent.TrimEnd()
        
        if ($newContent -ne $profileContent) {
            $utf8NoBom = New-Object System.Text.UTF8Encoding $false
            [System.IO.File]::WriteAllText($PROFILE, $newContent, $utf8NoBom)
            Write-Success "[OK] Removed cc function from $PROFILE"
            Write-Warning "  Please restart PowerShell or run: . `$PROFILE"
        } else {
            Write-Warning "[!] No cc function found in profile"
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
