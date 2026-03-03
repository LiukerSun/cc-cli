# CC-CLI 开发计划

## 版本规划

| 版本 | 主题 | 状态 |
|------|------|------|
| v1.2.0 | 基础功能完善 | 待开发 |
| v1.3.0 | 用户体验增强 | 待开发 |
| v1.4.0 | 安全性与稳定性 | 待开发 |
| v2.0.0 | 高级功能 | 待开发 |

---

## Phase 1: 基础功能完善 (v1.2.0) ✓ 完成

### 1.1 添加 --version 命令 ✓
**优先级**: 高  
**工作量**: 0.5h

**任务**:
- [x] 在 `bin/cc` 添加 `-V, --version` 选项
- [x] 在 `bin/cc.ps1` 添加 `-V, -version` 选项
- [x] 创建 `VERSION` 文件统一管理版本号
- [x] 安装脚本读取 VERSION 文件

**验收标准**:
```bash
cc --version  # 输出: cc version 1.2.0
cc -V         # 同上
```

---

### 1.2 添加删除模型功能 ✓
**优先级**: 高  
**工作量**: 1h

**任务**:
- [x] Bash: 添加 `-d, --delete N` 选项
- [x] PowerShell: 添加 `-d, -delete N` 选项
- [x] 删除前确认提示
- [x] 删除后显示更新后的列表

**验收标准**:
```bash
cc -d 2       # 删除模型 #2，需确认
cc --delete 3 # 同上
```

---

### 1.3 代码重构 - 版本号统一管理 ✓
**优先级**: 中  
**工作量**: 0.5h

**任务**:
- [x] 创建 `VERSION` 文件
- [x] `install.sh` 读取 VERSION 文件
- [x] `install.ps1` 读取 VERSION 文件
- [x] `bin/cc` 读取 VERSION 文件
- [x] 更新 CHANGELOG.md

**验收标准**:
- 修改 VERSION 文件后，所有脚本自动使用新版本号

---

### 1.4 修复 Bash 兼容性问题 ✓
**优先级**: 中  
**工作量**: 0.5h

**任务**:
- [x] 替换 `[[ =~ ]]` 正则为 `case` 语句 (bin/cc:253)
- [x] 测试 Bash 3.2 (macOS 默认) 兼容性
- [x] 测试 Bash 4.x 兼容性

**验收标准**:
- 在 macOS (Bash 3.2) 上所有功能正常

---

## Phase 2: 用户体验增强 (v1.3.0)

### 2.1 PowerShell 交互式选择增强
**优先级**: 高  
**工作量**: 3h

**任务**:
- [ ] 研究 PowerShell 上下键监听方案
- [ ] 实现类似 Bash 版的实时高亮菜单
- [ ] 支持数字键快速选择
- [ ] 支持 ESC 退出

**验收标准**:
- PowerShell 版体验与 Bash 版一致

---

### 2.2 添加配置验证功能
**优先级**: 中  
**工作量**: 2h

**任务**:
- [ ] 添加 `--validate` 选项
- [ ] 验证 JSON 格式
- [ ] 验证必填字段 (name, env)
- [ ] 可选: 测试 API 连通性

**验收标准**:
```bash
cc --validate
# 输出: ✓ Config valid (3 models)
# 或: ✗ Invalid JSON at line 5
```

---

### 2.3 添加 Shell 补全脚本
**优先级**: 中  
**工作量**: 2h

**任务**:
- [ ] 创建 `completions/cc.bash`
- [ ] 创建 `completions/cc.zsh`
- [ ] 创建 `completions/cc.ps1`
- [ ] 更新安装脚本自动安装补全
- [ ] 更新文档

**验收标准**:
```bash
cc --<TAB>    # 显示所有选项
cc -<TAB>     # 同上
cc 1<TAB>     # 显示模型名称
```

---

### 2.4 首次运行配置向导
**优先级**: 低  
**工作量**: 2h

**任务**:
- [ ] 检测是否首次运行 (配置文件不存在)
- [ ] 显示欢迎信息
- [ ] 引导添加第一个模型
- [ ] 显示后续使用提示

**验收标准**:
- 首次运行 `cc` 时自动进入配置向导

---

## Phase 3: 安全性与稳定性 (v1.4.0)

### 3.1 配置文件权限检查
**优先级**: 高  
**工作量**: 1h

**任务**:
- [ ] 启动时检查配置文件权限
- [ ] 权限过于开放时发出警告
- [ ] 提供 `--fix-permissions` 自动修复
- [ ] 更新文档说明安全最佳实践

**验收标准**:
```bash
cc
# 警告: Config file has overly permissive permissions (644)
# 建议: chmod 600 ~/.cc-config.json
# 运行: cc --fix-permissions 自动修复
```

---

### 3.2 改进 JSON 解析
**优先级**: 中  
**工作量**: 2h

**任务**:
- [ ] 检测系统是否安装 `jq`
- [ ] 优先使用 `jq` 解析 JSON
- [ ] 降级到 sed/awk 解析 (当前实现)
- [ ] 添加解析错误提示

**验收标准**:
- 有 jq 时使用 jq (更可靠)
- 无 jq 时使用 sed/awk (兼容性)

---

### 3.3 添加配置备份/恢复
**优先级**: 中  
**工作量**: 1.5h

**任务**:
- [ ] 添加 `--backup [FILE]` 导出配置
- [ ] 添加 `--restore FILE` 导入配置
- [ ] 备份时包含版本信息和时间戳
- [ ] 恢复时验证格式

**验收标准**:
```bash
cc --backup                    # 备份到 ~/.cc-config.backup.json
cc --backup my-config.json     # 备份到指定文件
cc --restore my-config.json    # 从文件恢复
```

---

### 3.4 添加单元测试
**优先级**: 中  
**工作量**: 3h

**任务**:
- [ ] 创建 `tests/` 目录
- [ ] Bash 测试框架 (bats 或 shell 单元测试)
- [ ] PowerShell 测试 (Pester)
- [ ] 测试核心功能: 列表、选择、添加、删除
- [ ] 添加 CI 配置 (GitHub Actions)

**验收标准**:
- `npm test` 或 `make test` 运行所有测试
- CI 在 PR 时自动运行测试

---

### 3.5 统一退出码
**优先级**: 低  
**工作量**: 0.5h

**任务**:
- [ ] 定义退出码规范
- [ ] 更新所有脚本使用统一退出码
- [ ] 文档说明退出码含义

**退出码规范**:
| 码 | 含义 |
|----|------|
| 0 | 成功 |
| 1 | 通用错误 |
| 2 | 配置文件错误 |
| 3 | 无效参数 |
| 4 | 用户取消 |
| 5 | 依赖缺失 |

---

## Phase 4: 高级功能 (v2.0.0)

### 4.1 API Key 加密存储
**优先级**: 高  
**工作量**: 4h

**任务**:
- [ ] macOS: 使用 Keychain 存储敏感信息
- [ ] Windows: 使用 Credential Manager
- [ ] Linux: 使用 libsecret/gnome-keyring
- [ ] 添加 `--encrypt` 迁移现有配置
- [ ] 添加 `--decrypt` 导出明文配置
- [ ] 更新文档

**验收标准**:
```bash
cc --encrypt   # 将 API keys 迁移到系统密钥环
cc --decrypt   # 导出为明文配置 (需确认)
```

---

### 4.2 自更新功能
**优先级**: 中  
**工作量**: 2h

**任务**:
- [ ] 添加 `--update` 选项
- [ ] 检查 GitHub Releases 最新版本
- [ ] 下载并替换脚本
- [ ] 支持指定版本: `--update 1.3.0`
- [ ] 支持 `--check-update` 仅检查不更新

**验收标准**:
```bash
cc --update           # 更新到最新版本
cc --update 1.3.0     # 更新到指定版本
cc --check-update     # 检查是否有新版本
```

---

### 4.3 多服务商配置模板
**优先级**: 低  
**工作量**: 1.5h

**任务**:
- [ ] 添加 `--template` 选项
- [ ] 预设模板: OpenAI, Gemini, Deepseek, Kimi, 智谱, etc.
- [ ] 用户选择模板后只需填 API Key
- [ ] 更新文档

**验收标准**:
```bash
cc --template         # 显示可用模板列表
cc --template openai  # 使用 OpenAI 模板添加
```

---

### 4.4 模型别名功能
**优先级**: 低  
**工作量**: 1.5h

**任务**:
- [ ] 支持为模型设置短别名
- [ ] 通过别名快速选择: `cc work`, `cc personal`
- [ ] 配置格式扩展

**配置格式**:
```json
{
    "name": "Claude (Work)",
    "alias": "work",
    "env": { ... }
}
```

**验收标准**:
```bash
cc work       # 使用别名为 "work" 的模型
cc --alias work 2   # 为模型 #2 设置别名
```

---

## 开发规范

### Git 分支策略
```
main           # 稳定版本
  ├── develop  # 开发分支
  │     ├── feature/version-command
  │     ├── feature/delete-model
  │     └── feature/shell-completion
```

### 提交信息格式
```
feat: add --version command
fix: resolve bash 3.2 compatibility issue
docs: update installation guide
refactor: unify version management
test: add unit tests for model selection
```

### 发布流程
1. 更新 `VERSION` 文件
2. 更新 `CHANGELOG.md`
3. 创建 Git tag: `git tag v1.2.0`
4. 推送: `git push origin main --tags`
5. GitHub Actions 自动发布

---

## 时间估算

| Phase | 预计工时 | 建议周期 |
|-------|----------|----------|
| v1.2.0 | 2.5h | 1 周 |
| v1.3.0 | 9h | 2 周 |
| v1.4.0 | 8h | 2 周 |
| v2.0.0 | 9h | 3 周 |

**总计**: 约 28.5 小时，8 周完成全部功能

---

## 快速开始

开始开发 Phase 1:
```bash
# 创建开发分支
git checkout -b feature/v1.2.0

# 按顺序完成以下任务:
# 1. 创建 VERSION 文件
# 2. 添加 --version 命令
# 3. 添加删除模型功能
# 4. 修复 Bash 兼容性
```
