# 更新日志

本项目的所有重要变更都将记录在此文件中。

## [1.5.0] - 2026-03-09

### 新增
- **完整卸载功能** - 为 Windows、Linux 和 macOS 添加交互式卸载流程
  - 删除脚本文件和安装目录
  - 可选删除配置文件和 Claude settings
  - 清理 PATH 环境变量
  - 清理 Shell 配置文件（PowerShell profile / .bashrc / .zshrc）
  - 支持 `--keep-config` 和 `--keep-settings` 参数保留特定文件
- **统一 CLI 参数** - 所有平台使用一致的命令参数
  - `-y, --bypass` - 跳过权限
  - `-e, --env` - 仅设置环境变量
  - `-l, --list` - 列出模型
  - `-c, --current` - 显示当前模型
  - `--edit` - 编辑配置
  - `-a, --add` - 添加模型
  - `-d, --delete` - 删除模型
  - `-s, --show` - 显示 API 密钥
  - `--validate` - 验证配置
  - `-U, --upgrade` - 升级
  - `-V, --version` - 版本
  - `--uninstall` - 卸载
  - `-h, --help` - 帮助
- **validate_config 命令** - 验证并自动修复配置文件（Linux/macOS）

### 变更
- **安装脚本** - 创建空配置文件 `[]` 而非默认模板
- **PowerShell 参数** - 不区分大小写，统一使用小写短选项 + 长选项

### 修复
- **中文乱码** - 移除所有中文提示，避免 Windows 终端乱码问题

## [1.4.0] - 2026-03-09

### 新增
- **阿里云百炼 Coding Plan 支持** - 添加阿里云百炼 Coding Plan 模型配置，兼容 Anthropic API 协议
  - 添加 `add_alibaba_coding_plan` / `Add-AlibabaCodingPlan` 函数
  - 支持自动从百炼 API 获取模型列表
  - 支持预定义模型列表作为回退（当 API 不可用时）
  - 支持的模型包括：qwen3.5-plus、qwen3-max、qwen3-coder 系列、glm-5、glm-4.7、kimi-k2.5、minimax-m2.5
  - API 端点：`https://coding.dashscope.aliyuncs.com/apps/anthropic`

### 变更
- bin/ccc - 添加 provider 选项 2) Alibaba Coding Plan
- bin/ccc.ps1 - 添加 Add-AlibabaCodingPlan 函数
- 更新交互提示，支持 3 个 provider 选项

## [1.3.4] - 2026-03-08

### 新增
- **Windows PowerShell Team Subagent 支持** - 为 Windows 脚本添加与 Bash 版本相同的 Team Subagent 模型同步功能
  - 添加 `Update-ClaudeSettings`、`Create-DefaultSettings`、`Get-ModelEnvValue` 函数
  - 修改 `Run-WithModel` 函数，在启动 Claude 前更新 `~/.claude/settings.json`
  - 导出 `CLAUDE_CODE_MODEL`、`CLAUDE_CODE_SMALL_MODEL`、`CLAUDE_CODE_SUBAGENT_MODEL` 环境变量
  - 自动禁用 Explore subagent（使用硬编码 Haiku 模型）

## [1.3.3] - 2026-03-08

### 新增
- **Team Subagent 模型同步** - 启动时自动更新 `~/.claude/settings.json`，确保 team subagent 使用相同的模型
  - 添加 `update_claude_settings` 函数，使用 `jq` 更新 Claude Code 的全局设置
  - 添加 `get_model_env_value` 函数，从配置文件中提取模型环境变量
  - 在 `run_with_model` 函数中自动调用设置更新

### 修复
- **Subagent 模型配置问题** - 解决 team 功能中 subagent 不使用 ccc 脚本配置的模型变量的问题
  - 之前 subagent 使用硬编码的模型（如 haiku, claude-opus-4-6）
  - 现在通过写入 `~/.claude/settings.json` 的 `env` 字段，让 subagent 继承模型配置
  - 添加 `CLAUDE_CODE_MODEL` 和 `CLAUDE_CODE_SMALL_MODEL` 环境变量，让 Claude Code 的 subagent 使用自定义模型
  - 修复 subagent 尝试使用不支持的 Anthropic 官方模型导致的 API 错误
  - **禁用 Explore subagent**：在 permissions.deny 中添加 `Agent(Explore)`，因为 Explore 硬编码使用 Haiku 模型，自定义 API 提供商不支持
  - **添加 CLAUDE_CODE_SUBAGENT_MODEL 环境变量**：这是官方文档中指定的 subagent 模型覆盖方式

### 改进
- **配置持久化** - 将模型配置写入 Claude Code 的全局设置文件
- **jq 集成** - 使用 jq 进行可靠的 JSON 更新操作（Bash 版本）
- **回退机制** - 当 jq 不可用时，使用纯 Bash 方式创建 settings 文件
- **PowerShell 实现** - 使用 PowerShell 原生 JSON 处理，无外部依赖

### 变更
- bin/ccc - 添加 `update_claude_settings`、`create_default_settings`、`get_model_env_value` 函数
- bin/ccc - 修改 `run_with_model` 函数，在启动 Claude 前更新 settings.json
- bin/ccc - 添加 `CLAUDE_CODE_MODEL`、`CLAUDE_CODE_SMALL_MODEL`、`CLAUDE_CODE_SUBAGENT_MODEL` 环境变量导出
- bin/ccc - 修改 `update_claude_settings` 函数，自动禁用 Explore subagent 并设置 model 字段
- bin/ccc.ps1 - 添加 `Update-ClaudeSettings`、`Create-DefaultSettings`、`Get-ModelEnvValue` 函数
- bin/ccc.ps1 - 修改 `Run-WithModel` 函数，在启动 Claude 前更新 settings.json
- bin/ccc.ps1 - 添加 `CLAUDE_CODE_MODEL`、`CLAUDE_CODE_SMALL_MODEL`、`CLAUDE_CODE_SUBAGENT_MODEL` 环境变量导出
- docs/configuration.md - 添加 Team Subagent 模型配置说明文档

## [1.3.1] - 2026-03-04

### 修复
- **配置文件损坏问题 (Issue #3)** - 修复运行 `ccc -a` 时配置文件被破坏的问题
  - Bash 版本 (bin/ccc): 添加 `jq` 支持，改进纯 Bash 实现的 JSON 处理
  - PowerShell 版本 (bin/ccc.ps1): 添加 `ConvertTo-Hashtable` 函数，修复 PSCustomObject 序列化问题
  - 之前运行 `ccc -a` 会导致现有模型配置消失，出现无效的 "value" 和 "Count" 字段
  - 现在正确保留所有现有配置，只追加新模型

### 改进
- **JSON 处理增强 (Bash)** - 添加 `jq` 支持，提供更可靠的 JSON 操作（当系统安装 jq 时）
- **回退机制优化 (Bash)** - 改进纯 Bash 实现的 JSON 追加逻辑，使用 `head` 替代 `sed` 操作
- **PowerShell 对象转换** - 添加 `ConvertTo-Hashtable` 函数，正确处理 PSCustomObject 到哈希表的转换
- **错误处理** - 添加配置备份和恢复机制，防止配置丢失
- **兼容性** - 同时支持有 jq 和无 jq 的环境（Bash）；跨平台支持（Windows/macOS/Linux）

### 变更
- bin/ccc - 重写 `save_model_config` 函数，修复配置损坏问题
- bin/ccc.ps1 - 添加 `ConvertTo-Hashtable` 函数，修复 `Save-ModelConfig` 函数

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
- bin/ccc - 添加 upgrade、check_latest_version、compare_versions 函数
- bin/ccc.ps1 - 添加 Upgrade-CC、Get-LatestVersion、Compare-Versions 函数
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
- bin/ccc - 添加版本号读取和 delete_model 函数
- bin/ccc.ps1 - 添加版本号读取和 Remove-Model 函数

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
- bin/ccc.ps1 - 添加 Save-JsonNoBOM 辅助函数并为所有 Get-Models 调用添加 @() 包装器

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

---

## 版本提交统计

| 版本 | 提交日期 | 提交类型 | 提交摘要 |
|------|----------|----------|----------|
| 1.5.0 | 2026-03-09 | feat | 完整卸载功能 + 统一 CLI 参数 |
| 1.4.0 | 2026-03-09 | feat | add Alibaba Coding Plan (百炼) support |
| 1.3.4 | 2026-03-08 | feat | Team Subagent 模型同步 (Windows) |
| 1.3.3 | 2026-03-08 | feat | Team Subagent 模型同步 |
| 1.3.1 | 2026-03-04 | fix | 配置文件损坏问题修复 |
| 1.3.0 | 2025-03-04 | feat | --upgrade 自动升级命令 |
| 1.2.0 | 2025-03-03 | feat | --version 和 --delete 命令 |
| 1.1.0 | 2025-03-03 | fix | Windows BOM 处理和兼容性修复 |
| 1.0.0 | 2024-03-03 | feat | initial release |

---

## 提交类型说明

本项目的提交信息遵循 [约定式提交规范](https://www.conventionalcommits.org/)：

- **feat** - 新功能
- **fix** - Bug 修复
- **docs** - 文档更新
- **refactor** - 代码重构（既不是新功能也不是修复）
- **style** - 代码格式调整（不影响代码运行）
- **perf** - 性能优化
- **test** - 测试相关
- **chore** - 构建过程或辅助工具变动
- **ci** - CI 配置
- **build** - 构建相关

---

*本文件最后更新：2026-03-09*
