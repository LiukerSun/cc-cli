# CC-CLI Windows 诊断脚本
# 运行此脚本诊断和修复配置文件问题

Write-Host "=== CC-CLI Windows 诊断 ===" -ForegroundColor Cyan
Write-Host ""

# 1. 检查配置文件
Write-Host "1. 检查配置文件..." -ForegroundColor Yellow
$configPath = "$env:USERPROFILE\.cc-config.json"

if (-not (Test-Path $configPath)) {
    Write-Host "✗ 配置文件不存在: $configPath" -ForegroundColor Red
    Write-Host ""
    Write-Host "正在创建配置文件..." -ForegroundColor Yellow
    
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
    
    $defaultConfig | ConvertTo-Json -Depth 10 | Out-File -FilePath $configPath -Encoding UTF8
    Write-Host "✓ 已创建默认配置文件" -ForegroundColor Green
} else {
    Write-Host "✓ 配置文件存在: $configPath" -ForegroundColor Green
    
    # 检查内容
    $content = Get-Content $configPath -Raw
    
    # 检查 BOM
    if ($content.Length -ge 3 -and $content[0] -eq [char]0xEF -and $content[1] -eq [char]0xBB -and $content[2] -eq [char]0xBF) {
        Write-Host "✗ 检测到 BOM 标记" -ForegroundColor Red
        Write-Host "  正在移除 BOM..." -ForegroundColor Yellow
        
        # 移除 BOM 并重新保存
        $content = $content[3..($content.Length - 1)]
        [System.IO.File]::WriteAllBytes($configPath, $content)
        Write-Host "✓ 已移除 BOM" -ForegroundColor Green
    }
    
    # 尝试解析 JSON
    Write-Host ""
    Write-Host "2. 测试 JSON 解析..." -ForegroundColor Yellow
    try {
        $config = Get-Content $configPath | ConvertFrom-Json
        Write-Host "✓ JSON 解析成功" -ForegroundColor Green
        Write-Host "  模型数量: $($config.Count)" -ForegroundColor Cyan
        
        Write-Host ""
        Write-Host "3. 模型列表:" -ForegroundColor Yellow
        for ($i = 0; $i -lt $config.Count; $i++) {
            $model = $config[$i]
            Write-Host "  $($i + 1)) $($model.name)" -ForegroundColor Cyan
        }
        
    } catch {
        Write-Host "✗ JSON 解析失败" -ForegroundColor Red
        Write-Host "  错误: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host ""
        Write-Host "正在修复配置文件..." -ForegroundColor Yellow
        
        # 创建正确的配置
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
        
        $fixedConfig | ConvertTo-Json -Depth 10 | Out-File -FilePath $configPath -Encoding UTF8
        Write-Host "✓ 配置文件已修复" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "=== 诊断完成 ===" -ForegroundColor Green
Write-Host ""
Write-Host "请运行 'cc --list' 测试" -ForegroundColor Yellow
