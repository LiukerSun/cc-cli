# CC-CLI - PowerShell Version
# https://github.com/LiukerSun/cc-cli

param(
    [Parameter(ValueFromRemainingArguments)]
    [string[]]$Arguments
)

# Configuration
$CONFIG_FILE = "$env:USERPROFILE\.cc-config.json"
$ENV_FILE = "$env:TEMP\cc-model-env.ps1"

# Functions
function Show-Help {
    Write-Host "Usage: cc [OPTIONS] [MODEL_INDEX] [-- CLAUDE_ARGS...]"
    Write-Host ""
    Write-Host "Start Claude with selected model configuration."
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -y, -bypass     Enable bypass permissions"
    Write-Host "  -e, -env        Only set environment variables"
    Write-Host "  -l, -list       List all available models"
    Write-Host "  -c, -current    Show current model"
    Write-Host "  -E, -edit       Edit configuration file"
    Write-Host "  -a, -add        Add a new model configuration"
    Write-Host "  -s, -show       Show API keys (partially hidden)"
    Write-Host "  -h, -help       Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  cc              Interactive model selection"
    Write-Host "  cc 2            Start Claude with model #2"
    Write-Host "  cc -y 3         Start Claude with model #3 and bypass"
    Write-Host "  cc -E           Edit configuration file"
    Write-Host "  cc -a           Add a new model"
}

function Get-Models {
    if (-not (Test-Path $CONFIG_FILE)) {
        Write-Error "Config file not found: $CONFIG_FILE"
        exit 1
    }
    
    $config = Get-Content $CONFIG_FILE | ConvertFrom-Json
    return $config
}

function Show-List {
    $models = Get-Models
    
    Write-Host "═══════════════════════════════════════"
    Write-Host "  Available AI Models" -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════"
    Write-Host ""
    
    $currentModel = ""
    if (Test-Path $ENV_FILE) {
        $currentModel = Select-String -Path $ENV_FILE -Pattern "# Model: (.*)" | ForEach-Object { $_.Matches.Groups[1].Value }
    }
    
    for ($i = 0; $i -lt $models.Count; $i++) {
        $model = $models[$i]
        $num = $i + 1
        
        if ($model.name -eq $currentModel) {
            Write-Host "  ➜ $num) $($model.name)" -ForegroundColor Green -NoNewline
            Write-Host " (current)" -ForegroundColor Gray
        } else {
            Write-Host "    $num) $($model.name)"
        }
    }
    
    Write-Host ""
    Write-Host "───────────────────────────────────────"
}

function Show-Current {
    if (Test-Path $ENV_FILE) {
        $model = Select-String -Path $ENV_FILE -Pattern "# Model: (.*)" | ForEach-Object { $_.Matches.Groups[1].Value }
        Write-Host "Current model: $model"
        
        if (Select-String -Path $ENV_FILE -Pattern "CLAUDE_SKIP_PERMISSIONS" -Quiet) {
            Write-Host "Bypass permissions: enabled"
        }
    } else {
        Write-Host "No model currently selected"
    }
}

function Show-Keys {
    $models = Get-Models
    
    Write-Host "═══════════════════════════════════════"
    Write-Host "  API Keys Overview" -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════"
    Write-Host ""
    
    for ($i = 0; $i -lt $models.Count; $i++) {
        $model = $models[$i]
        $num = $i + 1
        
        # Mask the key
        $key = $model.env.ANTHROPIC_AUTH_TOKEN
        $maskedKey = if ($key.Length -gt 12) { 
            $key.Substring(0,8) + "..." + $key.Substring($key.Length - 4) 
        } else { 
            $key.Substring(0,4) + "***" 
        }
        
        Write-Host "$num) $($model.name)" -ForegroundColor Cyan
        Write-Host "   URL: $($model.env.ANTHROPIC_BASE_URL)"
        Write-Host "   Key: $maskedKey" -ForegroundColor Gray
        Write-Host ""
    }
    
    Write-Host "───────────────────────────────────────"
    Write-Host ""
    Write-Host "To edit keys: cc -E"
    Write-Host "To add model: cc -a"
}

function Edit-Config {
    if (-not (Test-Path $CONFIG_FILE)) {
        Write-Error "Config file not found: $CONFIG_FILE"
        exit 1
    }
    
    $editor = $env:EDITOR
    if (-not $editor) {
        $editor = "notepad"
    }
    
    Write-Host "Opening config file with $editor..."
    Write-Host "File: $CONFIG_FILE"
    Write-Host ""
    
    & $editor $CONFIG_FILE
}

function Add-Model {
    if (-not (Test-Path $CONFIG_FILE)) {
        "[]" | Out-File -FilePath $CONFIG_FILE
    }
    
    Write-Host "═══════════════════════════════════════"
    Write-Host "  Add New Model Configuration" -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════"
    Write-Host ""
    
    $name = Read-Host "Model name (e.g., 'GPT-4')"
    if (-not $name) {
        Write-Error "Name is required"
        exit 1
    }
    
    Write-Host ""
    Write-Host "API Configuration:"
    $baseUrl = Read-Host "API Base URL"
    if (-not $baseUrl) {
        Write-Error "Base URL is required"
        exit 1
    }
    
    $apiKey = Read-Host "API Key"
    if (-not $apiKey) {
        Write-Error "API Key is required"
        exit 1
    }
    
    $mainModel = Read-Host "Main Model (e.g., 'gpt-4')"
    if (-not $mainModel) {
        Write-Error "Model name is required"
        exit 1
    }
    
    $fastModel = Read-Host "Fast Model (optional, press Enter to use main model)"
    if (-not $fastModel) {
        $fastModel = $mainModel
    }
    
    # Load existing config
    $config = Get-Content $CONFIG_FILE | ConvertFrom-Json
    
    # Add new model
    $newModel = @{
        name = $name
        env = @{
            ANTHROPIC_BASE_URL = $baseUrl
            ANTHROPIC_AUTH_TOKEN = $apiKey
            ANTHROPIC_MODEL = $mainModel
            ANTHROPIC_SMALL_FAST_MODEL = $fastModel
        }
    }
    
    $config += $newModel
    
    # Save config
    $config | ConvertTo-Json -Depth 10 | Out-File -FilePath $CONFIG_FILE -Encoding UTF8
    
    Write-Host ""
    Write-Host "✓ Model '$name' added successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Configuration saved to: $CONFIG_FILE"
}

function Select-Interactive {
    $models = Get-Models
    
    if ($models.Count -eq 0) {
        Write-Error "No models found"
        exit 1
    }
    
    $currentModel = ""
    if (Test-Path $ENV_FILE) {
        $currentModel = Select-String -Path $ENV_FILE -Pattern "# Model: (.*)" | ForEach-Object { $_.Matches.Groups[1].Value }
    }
    
    Write-Host "═══════════════════════════════════════" -NoNewline
    Write-Host "`r" -NoNewline
    Write-Host "  Available AI Models" -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════"
    Write-Host ""
    
    for ($i = 0; $i -lt $models.Count; $i++) {
        $model = $models[$i]
        $num = $i + 1
        
        if ($model.name -eq $currentModel) {
            Write-Host "  ➜ $num) $($model.name)" -ForegroundColor Green -NoNewline
            Write-Host " (current)" -ForegroundColor Gray
        } else {
            Write-Host "    $num) $($model.name)"
        }
    }
    
    Write-Host ""
    Write-Host "───────────────────────────────────────"
    Write-Host ""
    
    $selection = Read-Host "Select model [1-$($models.Count)]"
    
    if ($selection -match "^\d+$" -and [int]$selection -ge 1 -and [int]$selection -le $models.Count) {
        return [int]$selection
    } else {
        Write-Error "Invalid selection"
        exit 1
    }
}

function Run-WithModel {
    param(
        [int]$ModelIndex,
        [bool]$SkipPerm,
        [bool]$OnlyEnv,
        [string[]]$ClaudeArgs
    )
    
    $models = Get-Models
    
    if ($ModelIndex -lt 1 -or $ModelIndex -gt $models.Count) {
        Write-Error "Invalid model index: $ModelIndex"
        exit 1
    }
    
    $model = $models[$ModelIndex - 1]
    
    Write-Host "✓ Using model: $($model.name)" -ForegroundColor Cyan
    
    if ($SkipPerm) {
        Write-Host "✓ Bypass permissions enabled" -ForegroundColor Yellow
    }
    
    Write-Host "───────────────────────────────────────"
    
    # Create env file
    $envContent = "# Generated by cc command`n"
    $envContent += "# Model: $($model.name)`n"
    $envContent += "`$env:ANTHROPIC_BASE_URL = `"$($model.env.ANTHROPIC_BASE_URL)`"`n"
    $envContent += "`$env:ANTHROPIC_AUTH_TOKEN = `"$($model.env.ANTHROPIC_AUTH_TOKEN)`"`n"
    $envContent += "`$env:ANTHROPIC_MODEL = `"$($model.env.ANTHROPIC_MODEL)`"`n"
    $envContent += "`$env:ANTHROPIC_SMALL_FAST_MODEL = `"$($model.env.ANTHROPIC_SMALL_FAST_MODEL)`"`n"
    
    if ($SkipPerm) {
        $envContent += "`$env:CLAUDE_SKIP_PERMISSIONS = `"1`"`n"
    }
    
    $envContent | Out-File -FilePath $ENV_FILE -Encoding UTF8
    
    if ($OnlyEnv) {
        Write-Host ""
        Write-Host "Environment variables saved to: $ENV_FILE"
        Write-Host "Run: . $ENV_FILE"
        exit 0
    }
    
    # Set environment variables and run claude
    $env:ANTHROPIC_BASE_URL = $model.env.ANTHROPIC_BASE_URL
    $env:ANTHROPIC_AUTH_TOKEN = $model.env.ANTHROPIC_AUTH_TOKEN
    $env:ANTHROPIC_MODEL = $model.env.ANTHROPIC_MODEL
    $env:ANTHROPIC_SMALL_FAST_MODEL = $model.env.ANTHROPIC_SMALL_FAST_MODEL
    
    if ($SkipPerm) {
        $env:CLAUDE_SKIP_PERMISSIONS = "1"
    }
    
    # Run claude
    if (Get-Command claude -ErrorAction SilentlyContinue) {
        & claude @ClaudeArgs
    } else {
        Write-Error "Claude CLI not found"
        Write-Host "Install from: https://claude.ai"
        exit 1
    }
}

# Parse arguments
$skipPerm = $false
$onlyEnv = $false
$modelIndex = 0
$claudeArgs = @()
$foundSeparator = $false

for ($i = 0; $i -lt $Arguments.Count; $i++) {
    $arg = $Arguments[$i]
    
    if ($foundSeparator) {
        $claudeArgs += $arg
    } else {
        switch ($arg) {
            { $_ -in "-y", "-bypass", "--bypass" } { $skipPerm = $true }
            { $_ -in "-e", "-env", "--env" } { $onlyEnv = $true }
            { $_ -in "-l", "-list", "--list" } { Show-List; exit 0 }
            { $_ -in "-c", "-current", "--current" } { Show-Current; exit 0 }
            { $_ -in "-E", "-edit", "--edit" } { Edit-Config; exit 0 }
            { $_ -in "-a", "-add", "--add" } { Add-Model; exit 0 }
            { $_ -in "-s", "-show", "--show-keys" } { Show-Keys; exit 0 }
            { $_ -in "-h", "-help", "--help" } { Show-Help; exit 0 }
            "--" { $foundSeparator = $true }
            { $_ -match "^\d+$" } { $modelIndex = [int]$_ }
            default {
                Write-Error "Unknown option: $arg"
                Show-Help
                exit 1
            }
        }
    }
}

# Run
if ($modelIndex -eq 0) {
    $modelIndex = Select-Interactive
}

Run-WithModel -ModelIndex $modelIndex -SkipPerm $skipPerm -OnlyEnv $onlyEnv -ClaudeArgs $claudeArgs
