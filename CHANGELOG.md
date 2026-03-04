# 更新日志

本项目的所有重要变更都将记录在此文件中。

## [1.3.1] - 2026-03-04

### 修复
- **配置文件损坏问题 (Issue #3)** - 修复运行 `cc -a` 时配置文件被破坏的问题
  - Bash 版本 (bin/cc): 添加 `jq` 支持，改进纯 Bash 实现的 JSON 处理
  - PowerShell 版本 (bin/cc.ps1): 添加 `ConvertTo-Hashtable` 函数，修复 PSCustomObject 序列化问题
  - 之前运行 `cc -a` 会导致现有模型配置消失，出现无效的 "value" 和 "Count" 字段
  - 现在正确保留所有现有配置，只追加新模型

### 改进
- **JSON 处理增强 (Bash)** - 添加 `jq` 支持，提供更可靠的 JSON 操作（当系统安装 jq 时）
- **回退机制优化 (Bash)** - 改进纯 Bash 实现的 JSON 追加逻辑，使用 `head` 替代 `sed` 操作
- **PowerShell 对象转换** - 添加 `ConvertTo-Hashtable` 函数，正确处理 PSCustomObject 到哈希表的转换
- **错误处理** - 添加配置备份和恢复机制，防止配置丢失
- **兼容性** - 同时支持有 jq 和无 jq 的环境（Bash）；跨平台支持（Windows/macOS/Linux）

### 变更
- bin/cc - 重写 `save_model_config` 函数，修复配置损坏问题
- bin/cc.ps1 - 添加 `ConvertTo-Hashtable` 函数，修复 `Save-ModelConfig` 函数

## [1.3.0] - 2025-03-04

### 新增
- **--upgrade 命令** - 添加 `-U, --upgrade` 选项自动升级到最新版本
- **版本检查** - 自动从 GitHub 检查最新版本
- **智能版本比较** - 支持语义化版本号的比较算法

### 功能
- 自动下载并安装最新版本
- 保留用户配置和 API keys
- 跨平台支持（macOS/Linux/Windows）
- 详细的升级进度提示

### 变更
- bin/cc - 添加 upgrade、check_latest_version、compare_versions 函数
- bin/cc.ps1 - 添加 Upgrade-CC、Get-LatestVersion、Compare-Versions 函数
- README.md - 添加升级功能文档和使用示例

## [1.2.0] - 2025-03-03

### 新增
- **--version 命令** - 添加 `-V, --version` 选项显示版本号
- **--delete 命令** - 添加 `-d, --delete N` 选项删除指定模型配置
- **VERSION 文件** - 统一版本号管理，所有脚本读取同一个 VERSION 文件

### 修复
- **Bash 3.2 兼容性** - 将 `[[ =~ ]]` 正则匹配替换为 `case` 语句，提高 macOS 默认 Bash 版本兼容性

### 变更
- install.sh - 从 VERSION 文件读取版本号
- install.ps1 - 从 VERSION 文件读取版本号
- bin/cc - 添加版本号读取和 delete_model 函数
- bin/cc.ps1 - 添加版本号读取和 Remove-Model 函数

## [1.1.0] - 2025-03-03

### 修复
- **Windows PowerShell BOM 处理** - 修复 fix-config.ps1 中的 UTF-8 BOM 检测和移除逻辑，使用字节数组代替字符串
- **UTF-8 编码问题** - 所有脚本现在保存配置文件时不带 BOM，以防止 JSON 解析错误
- **PowerShell 数组解包** - 通过使用 @() 包装 Get-Models 修复单元素数组解包问题
- **空参数处理** - 添加检查以跳过空值或空白参数，防止"未知选项"错误
- **PowerShell 包装函数** - 简化包装器和自动更新机制以提高兼容性
- **Unicode 字符显示** - 将 ✓ 和其他 Unicode 字符替换为 ASCII 等效字符（[OK]）以提高终端兼容性
- **输出格式化** - 修复 Select-Interactive 函数中的标题显示问题

### 变更
- install.ps1 - 添加 Save-FileNoBOM 辅助函数并改进包装器替换逻辑
- fix-config.ps1 - 使用英文消息重写并正确处理 BOM 字节
- bin/cc.ps1 - 添加 Save-JsonNoBOM 辅助函数并为所有 Get-Models 调用添加 @() 包装器

### Windows 兼容性
- 改进与 Windows PowerShell 5.1 的兼容性
- 更好地处理不同 PowerShell 版本中的 UTF-8 编码
- 修复 Windows 终端中的中文字符显示问题

## [1.0.0] - 2024-03-03

### 新增功能
- 使用上下键的交互式模型选择
- 按数字直接选择模型
- API 密钥管理（查看、添加、编辑）
- Bypass 权限支持
- 配置文件管理
- 彩色终端输出
- 安装和卸载脚本
- 完整的文档

### 特性
- 零外部依赖（纯 Bash 实现）
- 兼容 Bash 3.2+（macOS 默认版本）
- 自动 shell 集成
- 持久化配置
- 模型历史记录

### 文档
- README.md
- LICENSE
- 安装指南
- 配置示例
