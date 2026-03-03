# CC-CLI Windows Diagnostic Script
# Run this script to diagnose and fix configuration file issues

# Helper function to save UTF-8 without BOM
function Save-JsonNoBOM {
    param(
        [string]$Path,
        $Object
    )
    
    $json = $Object | ConvertTo-Json -Depth 10
    $utf8NoBom = New-Object System.Text.UTF8Encoding $false
    [System.IO.File]::WriteAllText($Path, $json, $utf8NoBom)
}

Write-Host "=== CC-CLI Windows Diagnostic ===" -ForegroundColor Cyan
Write-Host ""

# 1. Check config file
Write-Host "1. Checking config file..." -ForegroundColor Yellow
$configPath = "$env:USERPROFILE\.cc-config.json"

if (-not (Test-Path $configPath)) {
    Write-Host "X Config file not found: $configPath" -ForegroundColor Red
    Write-Host ""
    Write-Host "Creating config file..." -ForegroundColor Yellow
    
    $defaultConfig = @(
        @{
            name = "Claude (Official)"
            env = @{
                ANTHROPIC_BASE_URL = "https://api.anthropic.com"
                ANTHROPIC_AUTH_TOKEN = "your-api-key-here"
                ANTHROPIC_MODEL = "claude-sonnet-4-20250514"
                ANTHROPIC_SMALL_FAST_MODEL = "claude-haiku-4-5-20251001"
            }
        }
    )
    
    Save-JsonNoBOM -Path $configPath -Object $defaultConfig
    Write-Host "+ Created default config file" -ForegroundColor Green
} else {
    Write-Host "+ Config file exists: $configPath" -ForegroundColor Green
    
    # Check content and handle BOM
    $bytes = [System.IO.File]::ReadAllBytes($configPath)
    
    # Check BOM (UTF-8 BOM: EF BB BF)
    if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF) {
        Write-Host "X BOM marker detected" -ForegroundColor Red
        Write-Host "  Removing BOM..." -ForegroundColor Yellow
        
        # Remove BOM and save
        $newBytes = $bytes[3..($bytes.Length - 1)]
        [System.IO.File]::WriteAllBytes($configPath, $newBytes)
        Write-Host "+ BOM removed" -ForegroundColor Green
    }
    
    # Re-read content
    $content = [System.IO.File]::ReadAllText($configPath)
    
    # Try to parse JSON
    Write-Host ""
    Write-Host "2. Testing JSON parsing..." -ForegroundColor Yellow
    try {
        $config = Get-Content $configPath | ConvertFrom-Json
        Write-Host "+ JSON parsing successful" -ForegroundColor Green
        Write-Host "  Model count: $($config.Count)" -ForegroundColor Cyan
        
        Write-Host ""
        Write-Host "3. Model list:" -ForegroundColor Yellow
        for ($i = 0; $i -lt $config.Count; $i++) {
            $model = $config[$i]
            $num = $i + 1
            $line = "  $num) " + $model.name
            Write-Host $line -ForegroundColor Cyan
        }
        
    } catch {
        Write-Host "X JSON parsing failed" -ForegroundColor Red
        Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host ""
        Write-Host "Fixing config file..." -ForegroundColor Yellow
        
        # Create correct config
        $fixedConfig = @(
            @{
                name = "ZHIPU AI"
                env = @{
                    ANTHROPIC_AUTH_TOKEN = "b9dc0380ae024c489995902fdc15c08b.jXiWqnWOD4pqPOaS"
                    ANTHROPIC_MODEL = "glm-5"
                    ANTHROPIC_SMALL_FAST_MODEL = "glm-4.7"
                    ANTHROPIC_BASE_URL = "https://open.bigmodel.cn/api/anthropic"
                }
            }
        )
        
        Save-JsonNoBOM -Path $configPath -Object $fixedConfig
        Write-Host "+ Config file fixed" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "=== Diagnostic Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Please run 'cc --list' to test" -ForegroundColor Yellow
