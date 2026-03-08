# CC-CLI 配置文件诊断和修复指南

## 🔍 问题诊断

你的配置文件显示只显示了标题，但没有显示模型列表。可能的原因：
1. JSON 格式不正确（BOM、编码、格式）
2. PowerShell 解析 JSON 失败

---

## 🛠️ 诊断步骤

### 1. 检查配置文件
```powershell
# 查看配置文件内容
Get-Content $env:USERPROFILE\.cc-config.json

# 检查是否包含 BOM
$content = Get-Content $env:USERPROFILE\.cc-config.json -Raw
if ($content[0] -eq [char]0xEF -and $content[1] -eq [char]0xBB -and $content[2] -eq [char]0xBF) {
    Write-Host "✗ 文件包含 BOM 标记" -ForegroundColor Red
}

# 尝试解析 JSON
try {
    $config = Get-Content $env:USERPROFILE\.cc-config.json | ConvertFrom-Json
    Write-Host "✓ JSON 解析成功" -ForegroundColor Green
    Write-Host "模型数量: $($config.Count)"
} catch {
    Write-Host "✗ JSON 解析失败" -ForegroundColor Red
    Write-Host $_.Exception.Message
}
```

---

## 🔧 修复方案

### 方式 1: 使用 PowerShell 重新创建（推荐）

```powershell
# 创建正确的配置
$config = @(
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

# 保存为 UTF-8 无 BOM
[System.IO.File]::WriteAllText(
    "$env:USERPROFILE\.cc-config.json",
    ($config | ConvertTo-Json -Depth 10),
    [System.Text.UTF8Encoding]::new($false)
)

Write-Host "✓ 配置文件已重新创建" -ForegroundColor Green
```

### 方式 2: 手动编辑

```powershell
# 用记事本打开
notepad $env:USERPROFILE\.cc-config.json
```

确保格式正确：
```json
[
    {
        "name": "ZHIPU AI",
        "env": {
            "ANTHROPIC_AUTH_TOKEN": "b9dc0380ae024c489995902fdc15c08b.jXiWqnWOD4pqPOaS",
            "ANTHROPIC_MODEL": "glm-5",
            "ANTHROPIC_SMALL_FAST_MODEL": "glm-4.7",
            "ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic"
        }
    }
]
```

**注意事项：**
- ✅ 使用双引号 `"`
- ✅ 正确的 JSON 语法
- ✅ 保存为 UTF-8 编码
- ❌ 不要有多余的逗号或空格

---

## ✅ 验证

```powershell
# 测试配置
cc --list

# 应该显示：
===================================
  Available AI Models
===================================

  > 1) ZHIPU AI

-----------------------------------
```

---

## 🐛 常见问题

### 问题 1: JSON 格式错误

错误示例（你的配置）：
```json
[
    {
        "name":  "zhipu",
        "env":  {
```

正确示例：
```json
[
    {
        "name": "ZHIPU AI",
        "env": {
```

### 问题 2: BOM 标记
某些编辑器会添加 BOM（字节顺序标记），导致解析失败。

**解决方法:**
```powershell
# 移除 BOM
$content = Get-Content $env:USERPROFILE\.cc-config.json -Raw
if ($content[0..2] -eq @(0xEF, 0xBB, 0xBF)) {
    $content = $content[3..($content.Length - 1)]
    [System.IO.File]::WriteAllBytes($env:USERPROFILE\.cc-config.json, $content)
}
```

---

## 📝 快速修复命令

```powershell
# 一键修复配置文件
$config = @(
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

# 保存
[System.IO.File]::WriteAllText(
    "$env:USERPROFILE\.cc-config.json",
    ($config | ConvertTo-Json -Depth 10),
    [System.Text.UTF8Encoding]::new($false)
)

# 验证
cc --list
```

---

## 🎯 如果仍然不工作

```powershell
# 完全重新安装
Remove-Item $env:USERPROFILE\.cc-config.json -Force
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex

# 编辑配置文件
cc -E
# 粘贴你的配置
# 保存并测试
cc --list
```
