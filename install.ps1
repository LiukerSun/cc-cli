# CC-CLI PowerShell Installation Script
# https://github.com/LiukerSun/cc-cli

param(
    [string]$Action = "install"
)

# Version
$VERSION = "1.0.0"
$REPO_URL = "https://github.com/LiukerSun/cc-cli"

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

# Banner
Write-Info "═══════════════════════════════════════"
Write-Info "  CC-CLI Installer v$VERSION (PowerShell)"
Write-Info "═══════════════════════════════════════"
Write-Host ""

# Check requirements
function Check-Requirements {
    Write-Warning "Checking system requirements..."
    
    # Check PowerShell version
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Error "✗ PowerShell 5.0 or later is required"
        exit 1
    }
    Write-Success "✓ PowerShell $($PSVersionTable.PSVersion) found"
    
    # Check for claude command
    if (-not (Get-Command claude -ErrorAction SilentlyContinue)) {
        Write-Warning "⚠ Claude CLI not found (optional)"
        Write-Host "  Install from: https://claude.ai"
    } else {
        Write-Success "✓ Claude CLI found"
    }
    
    Write-Host ""
}

# Create directories
function Create-Directories {
    Write-Warning "Creating directories..."
    
    New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path $SCRIPT_FILE -Parent) | Out-Null
    
    Write-Success "✓ Created $INSTALL_DIR"
    Write-Success "✓ Created $(Split-Path $SCRIPT_FILE -Parent)"
    Write-Host ""
}

# Download main script
function Install-Script {
    Write-Warning "Installing cc command..."
    
    # Download the PowerShell script
    $scriptUrl = "$REPO_URL/raw/main/bin/cc.ps1"
    try {
        Invoke-WebRequest -Uri $scriptUrl -OutFile $SCRIPT_FILE -ErrorAction Stop
        Write-Success "✓ Downloaded cc.ps1 to $SCRIPT_FILE"
    } catch {
        Write-Error "✗ Failed to download script: $_"
        Write-Host "Please download manually from: $scriptUrl"
        exit 1
    }
    
    Write-Host ""
}

# Create default config
function Create-Config {
    if (-not (Test-Path $CONFIG_FILE)) {
        Write-Warning "Creating default configuration..."
        
        $defaultConfig = @(
            "{",
            '    "name": "Claude (Official)",',
            '    "env": {',
            '        "ANTHROPIC_BASE_URL": "https://api.anthropic.com",',
            '        "ANTHROPIC_AUTH_TOKEN": "your-api-key-here",',
            '        "ANTHROPIC_MODEL": "claude-sonnet-4-20250514",',
            '        "ANTHROPIC_SMALL_FAST_MODEL": "claude-haiku-4-5-20251001"',
            "    }",
            "}"
        ) -join "`n"
        
        Set-Content -Path $CONFIG_FILE -Value "[$defaultConfig]"
        Write-Success "✓ Created config file: $CONFIG_FILE"
        Write-Warning "  Please edit this file to add your API keys"
    } else {
        Write-Success "✓ Config file already exists: $CONFIG_FILE"
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
        Write-Success "✓ Added $binDir to PATH"
        Write-Warning "  Please restart your terminal or run: `$env:PATH += `";$binDir`""
    } else {
        Write-Success "✓ Already in PATH"
    }
    Write-Host ""
}

# Create wrapper function
function Create-Wrapper {
    Write-Warning "Creating PowerShell wrapper..."
    
    $wrapperContent = @'
# CC-CLI PowerShell Wrapper
function cc {
    param(
        [Parameter(ValueFromRemainingArguments)]
        [string[]]$Arguments
    )
    
    & "$env:USERPROFILE\bin\cc.ps1" @Arguments
}
'@
    
    $profileDir = Split-Path $PROFILE -Parent
    if (-not (Test-Path $profileDir)) {
        New-Item -ItemType Directory -Force -Path $profileDir | Out-Null
    }
    
    # Add to profile
    $needsAdd = $true
    if (Test-Path $PROFILE) {
        if (Select-String -Path $PROFILE -Pattern "CC-CLI PowerShell Wrapper" -Quiet) {
            $needsAdd = $false
        }
    }
    
    if ($needsAdd) {
        Add-Content -Path $PROFILE -Value "`n$wrapperContent"
        Write-Success "✓ Added cc function to PowerShell profile"
        Write-Warning "  Please restart PowerShell or run: . `$PROFILE"
    } else {
        Write-Success "✓ Already in PowerShell profile"
    }
    Write-Host ""
}

# Print success message
function Print-Success {
    Write-Success "═══════════════════════════════════════"
    Write-Success "  Installation Complete!"
    Write-Success "═══════════════════════════════════════"
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
    Write-Warning "Uninstalling cc-cli..."
    
    if (Test-Path $SCRIPT_FILE) {
        Remove-Item $SCRIPT_FILE -Force
    }
    if (Test-Path $INSTALL_DIR) {
        Remove-Item $INSTALL_DIR -Recurse -Force
    }
    
    Write-Success "✓ Uninstalled cc-cli"
    Write-Warning "  Config file preserved: $CONFIG_FILE"
    Write-Warning "  Remove manually if needed: Remove-Item `"$CONFIG_FILE`""
    exit 0
}

# Main
switch ($Action.ToLower()) {
    "uninstall" { Uninstall }
    "help" {
        Write-Host "Usage: .\install.ps1 [ACTION]"
        Write-Host ""
        Write-Host "Actions:"
        Write-Host "  install     Install cc-cli (default)"
        Write-Host "  uninstall   Remove cc-cli"
        Write-Host "  help        Show this help message"
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
