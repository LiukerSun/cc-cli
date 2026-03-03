# CC-CLI - PowerShell Version
# https://github.com/LiukerSun/cc-cli

$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$VERSION_FILE = Join-Path $SCRIPT_DIR "..\VERSION"
if (Test-Path $VERSION_FILE) {
    $CC_VERSION = (Get-Content $VERSION_FILE -Raw).Trim()
} else {
    $CC_VERSION = "unknown"
}

$CONFIG_FILE = "$env:USERPROFILE\.cc-config.json"
$ENV_FILE = "$env:TEMP\cc-model-env.ps1"

function Save-JsonNoBOM {
    param(
        [string]$Path,
        $Object
    )
    
    $json = $Object | ConvertTo-Json -Depth 10
    $utf8NoBom = New-Object System.Text.UTF8Encoding $false
    [System.IO.File]::WriteAllText($Path, $json, $utf8NoBom)
}

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
    Write-Host "  -d, -delete N   Delete model #N"
    Write-Host "  -s, -show       Show API keys (partially hidden)"
    Write-Host "  -V, -version    Show version"
    Write-Host "  -h, -help       Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  cc              Interactive model selection"
    Write-Host "  cc 2            Start Claude with model #2"
    Write-Host "  cc -y 3         Start Claude with model #3 and bypass"
    Write-Host "  cc -E           Edit configuration file"
    Write-Host "  cc -a           Add a new model"
    Write-Host "  cc -d 2         Delete model #2"
}

function Show-Version {
    Write-Host "cc version $CC_VERSION"
}

function Get-Models {
    if (-not (Test-Path $CONFIG_FILE)) {
        Write-Error "Config file not found: $CONFIG_FILE"
        exit 1
    }
    
    $config = Get-Content $CONFIG_FILE | ConvertFrom-Json
    return @($config)
}

function Show-List {
    $models = @(Get-Models)
    
    Write-Host "==================================="
    Write-Host "  Available AI Models" -ForegroundColor Cyan
    Write-Host "==================================="
    Write-Host ""
    
    $currentModel = ""
    if (Test-Path $ENV_FILE) {
        $currentModel = Select-String -Path $ENV_FILE -Pattern "# Model: (.*)" | ForEach-Object { $_.Matches.Groups[1].Value }
    }
    
    $validCount = 0
    for ($i = 0; $i -lt $models.Count; $i++) {
        $model = $models[$i]
        if (-not $model.name) { continue }
        $validCount++
        $num = $i + 1
        
        if ($model.name -eq $currentModel) {
            Write-Host "  > $num) $($model.name)" -ForegroundColor Green -NoNewline
            Write-Host " (current)" -ForegroundColor Gray
        } else {
            Write-Host "    $num) $($model.name)"
        }
    }
    
    if ($validCount -eq 0) {
        Write-Host "  No models configured." -ForegroundColor Yellow
    }
    
    Write-Host ""
    Write-Host "-----------------------------------"
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
    $models = @(Get-Models)
    
    Write-Host "==================================="
    Write-Host "  API Keys Overview" -ForegroundColor Cyan
    Write-Host "==================================="
    Write-Host ""
    
    $hasModels = $false
    for ($i = 0; $i -lt $models.Count; $i++) {
        $model = $models[$i]
        if (-not $model.name) { continue }
        $hasModels = $true
        $num = $i + 1
        
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
    
    if (-not $hasModels) {
        Write-Host "  No models configured." -ForegroundColor Yellow
        Write-Host ""
    }
    
    Write-Host "-----------------------------------"
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
    
    Write-Host "==================================="
    Write-Host "  Add New Model Configuration" -ForegroundColor Cyan
    Write-Host "==================================="
    Write-Host ""
    Write-Host "Select provider:"
    Write-Host "  1) ZHIPU AI - auto fetch models"
    Write-Host "  2) Manual input"
    Write-Host ""
    
    $choice = Read-Host "Choice [1-2]"
    
    switch ($choice) {
        "1" { Add-ZhipuModel }
        "2" { Add-ManualModel }
        default {
            Write-Error "Invalid choice"
            exit 1
        }
    }
}

function Add-ZhipuModel {
    Write-Host ""
    Write-Host "==================================="
    Write-Host "  ZHIPU AI Configuration" -ForegroundColor Cyan
    Write-Host "==================================="
    Write-Host ""
    
    $apiKey = Read-Host "API Key"
    if (-not $apiKey) {
        Write-Error "API Key is required"
        exit 1
    }
    
    Write-Host ""
    Write-Host "Fetching models from ZHIPU AI..." -NoNewline
    
    try {
        $headers = @{
            "Authorization" = "Bearer $apiKey"
        }
        $response = Invoke-RestMethod -Uri "https://open.bigmodel.cn/api/paas/v4/models" -Headers $headers -Method Get -ErrorAction Stop
        
        Write-Host " Done!"
        Write-Host ""
        
        $models = @($response.data | Sort-Object -Property id)
        
        if ($models.Count -eq 0) {
            Write-Error "No models found"
            exit 1
        }
        
        Write-Host "Available Models:"
        Write-Host ""
        
        for ($i = 0; $i -lt $models.Count; $i++) {
            Write-Host "  $($i + 1)) $($models[$i].id)"
        }
        Write-Host ""
        
        $mainIdx = Read-Host "Select main model [1-$($models.Count)]"
        if (-not ($mainIdx -match "^\d+$") -or [int]$mainIdx -lt 1 -or [int]$mainIdx -gt $models.Count) {
            Write-Error "Invalid selection"
            exit 1
        }
        $mainModel = $models[[int]$mainIdx - 1].id
        
        $fastIdx = Read-Host "Select fast model [1-$($models.Count)] (default: same as main)"
        $fastModel = $mainModel
        if ($fastIdx -and $fastIdx -match "^\d+$" -and [int]$fastIdx -ge 1 -and [int]$fastIdx -le $models.Count) {
            $fastModel = $models[[int]$fastIdx - 1].id
        }
        
        $modelName = "ZHIPU ($mainModel)"
        
        Save-ModelConfig -Name $modelName -BaseUrl "https://open.bigmodel.cn/api/anthropic" -ApiKey $apiKey -MainModel $mainModel -FastModel $fastModel
        
    } catch {
        Write-Host ""
        Write-Error "Failed to fetch models: $_"
        exit 1
    }
}

function Add-ManualModel {
    Write-Host ""
    Write-Host "==================================="
    Write-Host "  Manual Configuration" -ForegroundColor Cyan
    Write-Host "==================================="
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
    
    Save-ModelConfig -Name $name -BaseUrl $baseUrl -ApiKey $apiKey -MainModel $mainModel -FastModel $fastModel
}

function Save-ModelConfig {
    param(
        [string]$Name,
        [string]$BaseUrl,
        [string]$ApiKey,
        [string]$MainModel,
        [string]$FastModel
    )
    
    $config = @()
    if (Test-Path $CONFIG_FILE) {
        $config = @(Get-Content $CONFIG_FILE | ConvertFrom-Json)
    }
    
    $newModel = @{
        name = $Name
        env = @{
            ANTHROPIC_BASE_URL = $BaseUrl
            ANTHROPIC_AUTH_TOKEN = $ApiKey
            ANTHROPIC_MODEL = $MainModel
            ANTHROPIC_SMALL_FAST_MODEL = $FastModel
        }
    }
    
    $config = @($config) + @($newModel)
    
    Save-JsonNoBOM -Path $CONFIG_FILE -Object $config
    
    Write-Host ""
    Write-Host "[OK] Model '$Name' added successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Configuration saved to: $CONFIG_FILE"
}

function Remove-Model {
    param(
        [int]$Index
    )
    
    if (-not (Test-Path $CONFIG_FILE)) {
        Write-Error "Config file not found: $CONFIG_FILE"
        exit 1
    }
    
    $models = @(Get-Models)
    
    if ($Index -lt 1 -or $Index -gt $models.Count) {
        Write-Error "Invalid model index. Must be between 1 and $($models.Count)"
        exit 1
    }
    
    $modelName = $models[$Index - 1].name
    
    Write-Host "==================================="
    Write-Host "  Delete Model Configuration" -ForegroundColor Cyan
    Write-Host "==================================="
    Write-Host ""
    Write-Host "  Model to delete: $Index) $modelName"
    Write-Host ""
    
    $confirm = Read-Host "Are you sure? (y/N)"
    
    if ($confirm -notin @("y", "Y", "yes", "YES")) {
        Write-Host "Cancelled."
        exit 0
    }
    
    $newModels = @()
    for ($i = 0; $i -lt $models.Count; $i++) {
        if ($i -ne ($Index - 1)) {
            $newModels += $models[$i]
        }
    }
    
    if ($newModels.Count -eq 0) {
        "[]" | Out-File -FilePath $CONFIG_FILE -Encoding UTF8
    } else {
        Save-JsonNoBOM -Path $CONFIG_FILE -Object $newModels
    }
    
    Write-Host ""
    Write-Host "[OK] Model '$modelName' deleted successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Remaining models:"
    Show-List
}

function Select-Interactive {
    $models = @(Get-Models)
    
    if ($models.Count -eq 0) {
        Write-Host "==================================="
        Write-Host "  No Models Configured" -ForegroundColor Yellow
        Write-Host "==================================="
        Write-Host ""
        Write-Host "Please add a model first:" -ForegroundColor Cyan
        Write-Host "  cc -a        Add a new model"
        Write-Host ""
        exit 0
    }
    
    $currentModel = ""
    if (Test-Path $ENV_FILE) {
        $currentModel = Select-String -Path $ENV_FILE -Pattern "# Model: (.*)" | ForEach-Object { $_.Matches.Groups[1].Value }
    }
    
    $selected = 1
    for ($i = 0; $i -lt $models.Count; $i++) {
        if ($models[$i].name -eq $currentModel) {
            $selected = $i + 1
            break
        }
    }
    
    function Draw-Menu {
        param([int]$Selected)
        
        [Console]::Clear()
        Write-Host "==================================="
        Write-Host "  Available AI Models" -ForegroundColor Cyan
        Write-Host "==================================="
        Write-Host ""
        
        for ($i = 0; $i -lt $models.Count; $i++) {
            $num = $i + 1
            $name = $models[$i].name
            $isCurrent = ($name -eq $currentModel)
            
            if ($i + 1 -eq $Selected) {
                Write-Host "  > " -NoNewline
                Write-Host "$num) $name" -BackgroundColor DarkBlue -ForegroundColor White -NoNewline
                if ($isCurrent) {
                    Write-Host " (current)" -ForegroundColor Gray
                } else {
                    Write-Host ""
                }
            } else {
                if ($isCurrent) {
                    Write-Host "  > " -NoNewline
                    Write-Host "$num) $name" -ForegroundColor Green -NoNewline
                    Write-Host " (current)" -ForegroundColor Gray
                } else {
                    Write-Host "    $num) $name"
                }
            }
        }
        
        Write-Host ""
        Write-Host "-----------------------------------"
        Write-Host ""
        Write-Host "  Up/Down: Navigate  |  Enter: Select  |  q: Exit" -ForegroundColor DarkGray
    }
    
    Draw-Menu -Selected $selected
    [Console]::CursorVisible = $false
    
    while ($true) {
        $key = [Console]::ReadKey($true)
        
        switch ($key.Key) {
            "UpArrow" {
                $selected--
                if ($selected -lt 1) { $selected = $models.Count }
                Draw-Menu -Selected $selected
            }
            "DownArrow" {
                $selected++
                if ($selected -gt $models.Count) { $selected = 1 }
                Draw-Menu -Selected $selected
            }
            "Enter" {
                [Console]::CursorVisible = $true
                Write-Host ""
                Write-Host ""
                Write-Host "[OK] Using model: $($models[$selected - 1].name)" -ForegroundColor Cyan
                return $selected
            }
            "Q" {
                [Console]::CursorVisible = $true
                exit 1
            }
            "Escape" {
                [Console]::CursorVisible = $true
                exit 1
            }
            { $_ -match "D([1-9])" } {
                $num = [int]$key.KeyChar
                if ($num -le $models.Count) {
                    $selected = $num
                    Draw-Menu -Selected $selected
                }
            }
        }
    }
}

function Run-WithModel {
    param(
        [int]$ModelIndex,
        [bool]$SkipPerm,
        [bool]$OnlyEnv,
        [string[]]$ClaudeArgs
    )
    
    $models = @(Get-Models)
    
    if ($ModelIndex -lt 1 -or $ModelIndex -gt $models.Count) {
        Write-Error "Invalid model index: $ModelIndex"
        exit 1
    }
    
    $model = $models[$ModelIndex - 1]
    
    Write-Host "[OK] Using model: $($model.name)" -ForegroundColor Cyan
    
    if ($SkipPerm) {
        Write-Host "[OK] Bypass permissions enabled" -ForegroundColor Yellow
    }
    
    Write-Host "-----------------------------------"
    
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
    
    $env:ANTHROPIC_BASE_URL = $model.env.ANTHROPIC_BASE_URL
    $env:ANTHROPIC_AUTH_TOKEN = $model.env.ANTHROPIC_AUTH_TOKEN
    $env:ANTHROPIC_MODEL = $model.env.ANTHROPIC_MODEL
    $env:ANTHROPIC_SMALL_FAST_MODEL = $model.env.ANTHROPIC_SMALL_FAST_MODEL
    
    if ($SkipPerm) {
        $env:CLAUDE_SKIP_PERMISSIONS = "1"
    }
    
    if (Get-Command claude -ErrorAction SilentlyContinue) {
        & claude @ClaudeArgs
    } else {
        Write-Error "Claude CLI not found"
        Write-Host "Install from: https://claude.ai"
        exit 1
    }
}

$skipPerm = $false
$onlyEnv = $false
$modelIndex = 0
$claudeArgs = @()
$foundSeparator = $false

for ($i = 0; $i -lt $args.Count; $i++) {
    $arg = $args[$i]
    
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
            { $_ -in "-d", "-delete", "--delete" } {
                $i++
                if ($i -ge $args.Count) {
                    Write-Error "--delete requires a model index"
                    exit 1
                }
                Remove-Model -Index ([int]$args[$i])
                exit 0
            }
            { $_ -in "-s", "-show", "--show-keys" } { Show-Keys; exit 0 }
            { $_ -in "-V", "-version", "--version" } { Show-Version; exit 0 }
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

if ($modelIndex -eq 0) {
    $modelIndex = Select-Interactive
}

Run-WithModel -ModelIndex $modelIndex -SkipPerm $skipPerm -OnlyEnv $onlyEnv -ClaudeArgs $claudeArgs
