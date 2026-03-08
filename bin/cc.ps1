# CC-CLI - PowerShell Version
# https://github.com/LiukerSun/cc-cli

$CC_VERSION = "unknown"
$VERSION_FILE = "$env:USERPROFILE\.cc-cli\VERSION"
if (Test-Path $VERSION_FILE) {
    $CC_VERSION = (Get-Content $VERSION_FILE -Raw).Trim()
}

$CONFIG_FILE = "$env:USERPROFILE\.cc-config.json"
$ENV_FILE = "$env:TEMP\cc-model-env.ps1"
$CLAUDE_SETTINGS_FILE = "$env:USERPROFILE\.claude\settings.json"

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
    Write-Host "  -y, --bypass      Enable bypass permissions"
    Write-Host "  -e, --env         Only set environment variables"
    Write-Host "  -l, --list        List all available models"
    Write-Host "  -c, --current     Show current model"
    Write-Host "  --edit            Edit configuration file"
    Write-Host "  -a, --add         Add a new model configuration"
    Write-Host "  -d, --delete N    Delete model #N"
    Write-Host "  -s, --show        Show API keys (partially hidden)"
    Write-Host "  --validate        Validate and repair config file"
    Write-Host "  -U, --upgrade     Upgrade to latest version"
    Write-Host "  -V, --version     Show version"
    Write-Host "  --uninstall       Uninstall cc-cli"
    Write-Host "  -h, --help        Show this help message"
    Write-Host ""
    Write-Host "Note: This command also updates ~/.claude/settings.json with the selected"
    Write-Host "      model to ensure team subagents use the same model configuration."
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  cc              Interactive model selection"
    Write-Host "  cc 2            Start Claude with model #2"
    Write-Host "  cc -y 3         Start Claude with model #3 and bypass"
    Write-Host "  cc -E           Edit configuration file"
    Write-Host "  cc -a           Add a new model"
    Write-Host "  cc -d 2         Delete model #2"
    Write-Host "  cc -v           Validate and repair config"
    Write-Host "  cc -U           Upgrade to latest version"
}

function Show-Version {
    Write-Host "cc version $CC_VERSION"
}

function Uninstall-CC {
    $installScript = "$env:USERPROFILE\.cc-cli\install.ps1"
    
    if (Test-Path $installScript) {
        & $installScript -Action uninstall @args
    } else {
        Write-Error "Install script not found: $installScript"
        Write-Host "Please run uninstall manually:"
        Write-Host "  irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex -ArgumentList 'uninstall'"
        exit 1
    }
}

function Get-LatestVersion {
    try {
        $url = "https://raw.githubusercontent.com/LiukerSun/cc-cli/main/VERSION"
        $client = New-Object System.Net.WebClient
        $version = $client.DownloadString($url).Trim()
        return $version
    } catch {
        Write-Error "Failed to check latest version: $_"
        return $null
    }
}

function Compare-Versions {
    param(
        [string]$v1,
        [string]$v2
    )
    
    if ($v1 -eq $v2) {
        return "equal"
    }
    
    $ver1 = $v1 -split '\.'
    $ver2 = $v2 -split '\.'
    
    $maxLen = [Math]::Max($ver1.Count, $ver2.Count)
    
    for ($i = 0; $i -lt $maxLen; $i++) {
        $num1 = if ($i -lt $ver1.Count) { [int]$ver1[$i] } else { 0 }
        $num2 = if ($i -lt $ver2.Count) { [int]$ver2[$i] } else { 0 }
        
        if ($num1 -gt $num2) {
            return "greater"
        }
        if ($num1 -lt $num2) {
            return "less"
        }
    }
    
    return "equal"
}

function Upgrade-CC {
    Write-Host "==================================="
    Write-Host "  CC-CLI Updater" -ForegroundColor Cyan
    Write-Host "==================================="
    Write-Host ""
    
    Write-Host "Checking latest version..." -NoNewline
    
    $latestVersion = Get-LatestVersion
    
    if (-not $latestVersion) {
        exit 1
    }
    
    Write-Host " Done!"
    Write-Host ""
    Write-Host "Current version: $CC_VERSION"
    Write-Host "Latest version:  $latestVersion"
    Write-Host ""
    
    $comparison = Compare-Versions -v1 $CC_VERSION -v2 $latestVersion
    
    if ($comparison -eq "equal") {
        Write-Host "[OK] Already running the latest version!" -ForegroundColor Green
        exit 0
    } elseif ($comparison -eq "greater") {
        Write-Host "[!] You're running a newer version than the latest release" -ForegroundColor Yellow
        exit 0
    }
    
    Write-Host "Upgrading from $CC_VERSION to $latestVersion..." -ForegroundColor Cyan
    Write-Host ""
    
    $scriptUrl = "https://raw.githubusercontent.com/LiukerSun/cc-cli/main/bin/cc.ps1"
    $versionUrl = "https://raw.githubusercontent.com/LiukerSun/cc-cli/main/VERSION"
    
    $installDir = "$env:USERPROFILE\.cc-cli"
    $scriptFile = "$env:USERPROFILE\bin\cc.ps1"
    
    Write-Host "Downloading latest version..." -NoNewline
    
    try {
        $webClient = New-Object System.Net.WebClient
        $webClient.Encoding = [System.Text.Encoding]::UTF8
        
        $scriptContent = $webClient.DownloadString($scriptUrl)
        $versionContent = $webClient.DownloadString($versionUrl).Trim()
        
        Write-Host " Done!"
        Write-Host ""
        
        New-Item -ItemType Directory -Force -Path $installDir | Out-Null
        New-Item -ItemType Directory -Force -Path (Split-Path $scriptFile -Parent) | Out-Null
        
        $utf8NoBom = New-Object System.Text.UTF8Encoding $false
        [System.IO.File]::WriteAllText($scriptFile, $scriptContent, $utf8NoBom)
        [System.IO.File]::WriteAllText("$installDir\VERSION", $versionContent, $utf8NoBom)
        
        Write-Host "==================================="
        Write-Host "  [OK] Upgrade Complete!" -ForegroundColor Green
        Write-Host "==================================="
        Write-Host "Upgraded from $CC_VERSION to $latestVersion"
        Write-Host ""
        Write-Host "Run 'cc -V' to verify the upgrade"
        
    } catch {
        Write-Host ""
        Write-Error "Failed to upgrade: $_"
        Write-Host "Please try manual upgrade:"
        Write-Host "  irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex"
        exit 1
    }
}

function Test-ConfigValid {
    param(
        [object]$Config
    )
    
    if (-not $Config) { return $false }
    
    $configArray = @($Config)
    
    foreach ($item in $configArray) {
        if (-not $item.name) { return $false }
        if (-not $item.env) { return $false }
        if (-not $item.env.ANTHROPIC_BASE_URL) { return $false }
        if (-not $item.env.ANTHROPIC_AUTH_TOKEN) { return $false }
        if (-not $item.env.ANTHROPIC_MODEL) { return $false }
        if (-not $item.env.ANTHROPIC_SMALL_FAST_MODEL) { return $false }
    }
    
    return $true
}

function Repair-Config {
    param(
        [object]$Config
    )
    
    $validModels = @()
    $configArray = @($Config)
    
    foreach ($item in $configArray) {
        if ($item.name -and $item.env -and 
            $item.env.ANTHROPIC_BASE_URL -and 
            $item.env.ANTHROPIC_AUTH_TOKEN -and 
            $item.env.ANTHROPIC_MODEL) {
            
            if (-not $item.env.ANTHROPIC_SMALL_FAST_MODEL) {
                $item.env.ANTHROPIC_SMALL_FAST_MODEL = $item.env.ANTHROPIC_MODEL
            }
            
            $validModels += $item
        }
    }
    
    return $validModels
}

function Get-Models {
    if (-not (Test-Path $CONFIG_FILE)) {
        Write-Error "Config file not found: $CONFIG_FILE"
        exit 1
    }
    
    try {
        $rawContent = Get-Content $CONFIG_FILE -Raw
        $config = $rawContent | ConvertFrom-Json
    } catch {
        Write-Error "Config file is not valid JSON: $_"
        Write-Host "Please fix or delete: $CONFIG_FILE" -ForegroundColor Yellow
        exit 1
    }
    
    if (-not (Test-ConfigValid -Config $config)) {
        Write-Host "[!] Config file has invalid entries, repairing..." -ForegroundColor Yellow
        $config = Repair-Config -Config $config
        Save-JsonNoBOM -Path $CONFIG_FILE -Object $config
        Write-Host "[OK] Config file repaired" -ForegroundColor Green
    }
    
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

function Validate-Config {
    Write-Host "==================================="
    Write-Host "  Config File Validator" -ForegroundColor Cyan
    Write-Host "==================================="
    Write-Host ""
    Write-Host "File: $CONFIG_FILE"
    Write-Host ""
    
    if (-not (Test-Path $CONFIG_FILE)) {
        Write-Host "[!] Config file does not exist" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Creating empty config file..."
        Save-JsonNoBOM -Path $CONFIG_FILE -Object @()
        Write-Host "[OK] Empty config file created" -ForegroundColor Green
        return
    }
    
    try {
        $rawContent = Get-Content $CONFIG_FILE -Raw
        $config = $rawContent | ConvertFrom-Json
    } catch {
        Write-Host "[ERROR] Config file is not valid JSON" -ForegroundColor Red
        Write-Host "Error: $_" -ForegroundColor Red
        Write-Host ""
        Write-Host "Please fix manually: cc -E"
        exit 1
    }
    
    Write-Host "[OK] JSON syntax is valid" -ForegroundColor Green
    
    $configArray = @($config)
    $validCount = 0
    $invalidCount = 0
    $errors = @()
    
    for ($i = 0; $i -lt $configArray.Count; $i++) {
        $item = $configArray[$i]
        $num = $i + 1
        
        if (-not $item.name) {
            $errors += "Entry #$num : missing 'name' field"
            $invalidCount++
            continue
        }
        
        if (-not $item.env) {
            $errors += "Entry #$num ($($item.name)): missing 'env' field"
            $invalidCount++
            continue
        }
        
        $missingFields = @()
        if (-not $item.env.ANTHROPIC_BASE_URL) { $missingFields += "ANTHROPIC_BASE_URL" }
        if (-not $item.env.ANTHROPIC_AUTH_TOKEN) { $missingFields += "ANTHROPIC_AUTH_TOKEN" }
        if (-not $item.env.ANTHROPIC_MODEL) { $missingFields += "ANTHROPIC_MODEL" }
        
        if ($missingFields.Count -gt 0) {
            $errors += "Entry #$num ($($item.name)): missing fields: $($missingFields -join ', ')"
            $invalidCount++
            continue
        }
        
        $validCount++
    }
    
    Write-Host ""
    Write-Host "Valid entries:   $validCount" -ForegroundColor Green
    Write-Host "Invalid entries: $invalidCount" -ForegroundColor $(if ($invalidCount -gt 0) { "Red" } else { "Green" })
    
    if ($errors.Count -gt 0) {
        Write-Host ""
        Write-Host "Errors found:" -ForegroundColor Red
        foreach ($err in $errors) {
            Write-Host "  - $err" -ForegroundColor Red
        }
        
        Write-Host ""
        Write-Host "Repairing config file..." -ForegroundColor Yellow
        $repaired = Repair-Config -Config $config
        
        if ($repaired.Count -lt $configArray.Count) {
            Write-Host "[OK] Removed $($configArray.Count - $repaired.Count) invalid entries" -ForegroundColor Yellow
        }
        
        Save-JsonNoBOM -Path $CONFIG_FILE -Object $repaired
        Write-Host "[OK] Config file repaired" -ForegroundColor Green
        
        Write-Host ""
        Write-Host "Remaining models:"
        foreach ($item in $repaired) {
            Write-Host "  - $($item.name)"
        }
    } else {
        Write-Host ""
        Write-Host "[OK] Config file is valid!" -ForegroundColor Green
    }
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
    Write-Host "  2) Alibaba Coding Plan - auto fetch models"
    Write-Host "  3) Manual input"
    Write-Host ""

    $choice = Read-Host "Choice [1-3]"

    switch ($choice) {
        "1" { Add-ZhipuModel }
        "2" { Add-AlibabaCodingPlan }
        "3" { Add-ManualModel }
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

function Add-AlibabaCodingPlan {
    Write-Host ""
    Write-Host "==================================="
    Write-Host "  Alibaba Coding Plan Configuration" -ForegroundColor Cyan
    Write-Host "==================================="
    Write-Host ""

    $apiKey = Read-Host "API Key"
    if (-not $apiKey) {
        Write-Error "API Key is required"
        exit 1
    }

    Write-Host ""
    Write-Host "Fetching models from Alibaba Coding Plan..." -NoNewline

    $allModels = @()
    $apiSuccess = $false

    try {
        $headers = @{
            "Authorization" = "Bearer $apiKey"
        }

        # Try the models API endpoint
        $response = Invoke-RestMethod -Uri "https://dashscope.aliyuncs.com/api/v1/models" -Headers $headers -Method Get -ErrorAction Stop

        Write-Host " Done!"
        Write-Host ""
        $apiSuccess = $true

        # Parse models from response and filter for coding-related models
        $allModels = @($response.data | Where-Object {
            $_.model -match '^(qwen|glm|kimi|minimax)'
        } | Sort-Object -Property model)

        if ($allModels.Count -eq 0) {
            # Try alternative field name
            $allModels = @($response.data | Where-Object {
                $_.id -match '^(qwen|glm|kimi|minimax)'
            } | Sort-Object -Property id)
        }

    } catch {
        Write-Host ""
        Write-Host "Note: API request failed - this is normal if using Coding Plan API key" -ForegroundColor Yellow
        Write-Host "Using predefined model list from Coding Plan documentation..."
    }

    # If API didn't return models, use predefined list from documentation
    if ($allModels.Count -eq 0) {
        $predefinedModels = @(
            "qwen3.5-plus",
            "qwen3-max-2026-01-23",
            "qwen3-coder-next",
            "qwen3-coder-plus",
            "glm-5",
            "glm-4.7",
            "kimi-k2.5",
            "minimax-m2.5"
        )
        $allModels = @($predefinedModels | ForEach-Object { [PSCustomObject]@{ id = $_ } })
    }

    Write-Host "Available Models:"
    Write-Host ""

    for ($i = 0; $i -lt $allModels.Count; $i++) {
        $modelId = if ($allModels[$i].model) { $allModels[$i].model } else { $allModels[$i].id }
        Write-Host "  $($i + 1)) $modelId"
    }
    Write-Host ""

    $mainIdx = Read-Host "Select main model [1-$($allModels.Count)]"
    if (-not ($mainIdx -match "^\d+$") -or [int]$mainIdx -lt 1 -or [int]$mainIdx -gt $allModels.Count) {
        Write-Error "Invalid selection"
        exit 1
    }
    $mainModelObj = $allModels[[int]$mainIdx - 1]
    $mainModel = if ($mainModelObj.model) { $mainModelObj.model } else { $mainModelObj.id }

    $fastIdx = Read-Host "Select fast model [1-$($allModels.Count)] (default: same as main)"
    $fastModel = $mainModel
    if ($fastIdx -and $fastIdx -match "^\d+$" -and [int]$fastIdx -ge 1 -and [int]$fastIdx -le $allModels.Count) {
        $fastModelObj = $allModels[[int]$fastIdx - 1]
        $fastModel = if ($fastModelObj.model) { $fastModelObj.model } else { $fastModelObj.id }
    }

    $modelName = "Alibaba Coding Plan ($mainModel)"

    Save-ModelConfig -Name $modelName -BaseUrl "https://coding.dashscope.aliyuncs.com/apps/anthropic" -ApiKey $apiKey -MainModel $mainModel -FastModel $fastModel
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

function ConvertTo-Hashtable {
    param(
        [Parameter(ValueFromPipeline)]
        $InputObject
    )
    
    if ($null -eq $InputObject) {
        return $null
    }
    
    if ($InputObject -is [System.Collections.IEnumerable] -and $InputObject -isnot [string]) {
        $collection = @()
        foreach ($object in $InputObject) {
            $collection += ConvertTo-Hashtable -InputObject $object
        }
        return $collection
    }
    elseif ($InputObject -is [System.Management.Automation.PSCustomObject]) {
        $hash = @{}
        foreach ($property in $InputObject.PSObject.Properties) {
            $hash[$property.Name] = ConvertTo-Hashtable -InputObject $property.Value
        }
        return $hash
    }
    else {
        return $InputObject
    }
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
        $rawConfig = Get-Content $CONFIG_FILE | ConvertFrom-Json
        $allConfig = @(ConvertTo-Hashtable -InputObject $rawConfig)
        foreach ($item in $allConfig) {
            if ($item.name) {
                $config += $item
            }
        }
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

function Get-ModelEnvValue {
    param(
        [int]$ModelIndex,
        [string]$KeyName
    )

    $models = @(Get-Models)

    if ($ModelIndex -lt 1 -or $ModelIndex -gt $models.Count) {
        return $null
    }

    $model = $models[$ModelIndex - 1]

    if ($model.env -and $model.env.$KeyName) {
        return $model.env.$KeyName
    }

    return $null
}

function Update-ClaudeSettings {
    param(
        [string]$MainModel,
        [string]$FastModel
    )

    # Create .claude directory if it doesn't exist
    $claudeDir = "$env:USERPROFILE\.claude"
    if (-not (Test-Path $claudeDir)) {
        New-Item -ItemType Directory -Force -Path $claudeDir | Out-Null
    }

    # Load existing settings or create new
    if (Test-Path $CLAUDE_SETTINGS_FILE) {
        try {
            $settings = Get-Content $CLAUDE_SETTINGS_FILE -Raw | ConvertFrom-Json
        } catch {
            $settings = [PSCustomObject]@{}
        }
    } else {
        $settings = [PSCustomObject]@{}
    }

    # Initialize env if not exists
    if (-not $settings.env) {
        $settings | Add-Member -NotePropertyName "env" -NotePropertyValue ([PSCustomObject]@{})
    }

    # Set environment variables for team subagent support
    $settings.env | Add-Member -NotePropertyName "ANTHROPIC_MODEL" -NotePropertyValue $MainModel -Force
    $settings.env | Add-Member -NotePropertyName "ANTHROPIC_SMALL_FAST_MODEL" -NotePropertyValue $FastModel -Force
    $settings.env | Add-Member -NotePropertyName "CLAUDE_CODE_MODEL" -NotePropertyValue $MainModel -Force
    $settings.env | Add-Member -NotePropertyName "CLAUDE_CODE_SMALL_MODEL" -NotePropertyValue $FastModel -Force
    $settings.env | Add-Member -NotePropertyName "CLAUDE_CODE_SUBAGENT_MODEL" -NotePropertyValue $MainModel -Force

    # Set model field for main agent
    $settings | Add-Member -NotePropertyName "model" -NotePropertyValue $MainModel -Force

    # Initialize permissions if not exists
    if (-not $settings.permissions) {
        $settings | Add-Member -NotePropertyName "permissions" -NotePropertyValue ([PSCustomObject]@{})
    }

    # Add Agent(Explore) to deny list (Explore uses hard-coded Haiku)
    if (-not $settings.permissions.deny) {
        $settings.permissions | Add-Member -NotePropertyName "deny" -NotePropertyValue @("Agent(Explore)")
    } else {
        $denyList = @($settings.permissions.deny)
        if ("Agent(Explore)" -notin $denyList) {
            $denyList = @($denyList + @("Agent(Explore)"))
            $settings.permissions.deny = $denyList
        }
    }

    # Save settings
    Save-JsonNoBOM -Path $CLAUDE_SETTINGS_FILE -Object $settings
}

function Create-DefaultSettings {
    param(
        [string]$MainModel,
        [string]$FastModel
    )

    # Create .claude directory if it doesn't exist
    $claudeDir = "$env:USERPROFILE\.claude"
    if (-not (Test-Path $claudeDir)) {
        New-Item -ItemType Directory -Force -Path $claudeDir | Out-Null
    }

    $settings = [PSCustomObject]@{
        env = [PSCustomObject]@{
            ANTHROPIC_MODEL = $MainModel
            ANTHROPIC_SMALL_FAST_MODEL = $FastModel
            CLAUDE_CODE_MODEL = $MainModel
            CLAUDE_CODE_SMALL_MODEL = $FastModel
            CLAUDE_CODE_SUBAGENT_MODEL = $MainModel
        }
        model = $MainModel
        permissions = [PSCustomObject]@{
            deny = @("Agent(Explore)")
        }
    }

    Save-JsonNoBOM -Path $CLAUDE_SETTINGS_FILE -Object $settings
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

    # Get model environment values
    $mainModel = Get-ModelEnvValue -ModelIndex $ModelIndex -KeyName "ANTHROPIC_MODEL"
    $fastModel = Get-ModelEnvValue -ModelIndex $ModelIndex -KeyName "ANTHROPIC_SMALL_FAST_MODEL"

    # Default fast model to main model if not set
    if (-not $fastModel) {
        $fastModel = $mainModel
    }

    # Update Claude settings for team subagent support
    Update-ClaudeSettings -MainModel $mainModel -FastModel $fastModel
    Write-Host "[OK] Updated Claude settings (for team subagent)" -ForegroundColor Cyan

    $envContent = "# Generated by cc command`n"
    $envContent += "# Model: $($model.name)`n"
    $envContent += "`$env:ANTHROPIC_BASE_URL = `"$($model.env.ANTHROPIC_BASE_URL)`"`n"
    $envContent += "`$env:ANTHROPIC_AUTH_TOKEN = `"$($model.env.ANTHROPIC_AUTH_TOKEN)`"`n"
    $envContent += "`$env:ANTHROPIC_MODEL = `"$($model.env.ANTHROPIC_MODEL)`"`n"
    $envContent += "`$env:ANTHROPIC_SMALL_FAST_MODEL = `"$($model.env.ANTHROPIC_SMALL_FAST_MODEL)`"`n"

    if ($SkipPerm) {
        $envContent += "`$env:CLAUDE_SKIP_PERMISSIONS = `"1`"`n"
    }

    # Add CLAUDE_CODE_MODEL environment variables for subagent support
    $envContent += "`$env:CLAUDE_CODE_MODEL = `"$mainModel`"`n"
    $envContent += "`$env:CLAUDE_CODE_SMALL_MODEL = `"$fastModel`"`n"
    $envContent += "`$env:CLAUDE_CODE_SUBAGENT_MODEL = `"$mainModel`"`n"

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

    # Add CLAUDE_CODE_MODEL environment variables for subagent support
    $env:CLAUDE_CODE_MODEL = $mainModel
    $env:CLAUDE_CODE_SMALL_MODEL = $fastModel
    $env:CLAUDE_CODE_SUBAGENT_MODEL = $mainModel

    if (Get-Command claude -ErrorAction SilentlyContinue) {
        if ($SkipPerm) {
            & claude --dangerously-skip-permissions @ClaudeArgs
        } else {
            & claude @ClaudeArgs
        }
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
    $argLower = $arg.ToLower()
    
    if ($foundSeparator) {
        $claudeArgs += $arg
    } else {
        switch ($argLower) {
            { $_ -in "-y", "--bypass" } { $skipPerm = $true }
            { $_ -in "-e", "--env" } { $onlyEnv = $true }
            { $_ -in "-l", "--list" } { Show-List; exit 0 }
            { $_ -in "-c", "--current" } { Show-Current; exit 0 }
            "--edit" { Edit-Config; exit 0 }
            { $_ -in "-a", "--add" } { Add-Model; exit 0 }
            { $_ -in "-d", "--delete" } {
                $i++
                if ($i -ge $args.Count) {
                    Write-Error "--delete requires a model index"
                    exit 1
                }
                Remove-Model -Index ([int]$args[$i])
                exit 0
            }
            { $_ -in "-s", "--show" } { Show-Keys; exit 0 }
            "--validate" { Validate-Config; exit 0 }
            { $_ -in "-u", "--upgrade" } { Upgrade-CC; exit 0 }
            { $_ -in "-v", "--version" } { Show-Version; exit 0 }
            "--uninstall" {
                $remainingArgs = @()
                for ($j = $i + 1; $j -lt $args.Count; $j++) {
                    $remainingArgs += $args[$j]
                }
                Uninstall-CC @remainingArgs
                exit 0
            }
            { $_ -in "-h", "--help" } { Show-Help; exit 0 }
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
