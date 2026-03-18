# 使用指南

## 查看帮助

```bash
ccc --help
```

## 查看所有模型

```bash
ccc --list
```

## 交互式选择

```bash
ccc
```

使用上下键选择，按 Enter 确认。

## 查看当前模型

```bash
ccc --current
```

## 查看 API Keys

```bash
ccc --show-keys
```

## 编辑配置

```bash
ccc --edit
```

## 添加新模型

```bash
ccc --add
```

## 使用特定模型启动

```bash
ccc 2
```

## Bypass 模式启动

```bash
ccc -y 2
```

## 传递参数给 Claude

```bash
ccc 1 -- --help
ccc 2 -- "写一个 hello world 程序"
```

## 仅设置环境变量

```bash
ccc -e 2
source /tmp/cc-model-env.sh
claude
```
