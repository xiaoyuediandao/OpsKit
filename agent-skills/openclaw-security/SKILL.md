---
name: openclaw-security
description: OpenClaw 安全防护技能。扫描和加固 OpenClaw workspace 的安全配置，包括 SECURITY.md 部署、AGENTS.md 安全引用注入、SOUL.md 安全边界补丁、敏感文件权限修复、密钥泄露检测、工具输出脱敏。当用户提到"安全扫描"、"安全加固"、"security scan"、"security fix"、"部署安全防护"、"检查安全配置"、"脱敏"、"redact"、"敏感信息"时触发。
allowed-tools: Bash(opskit:*), Bash(cat:*), Bash(ls:*), Bash(chmod:*)
---

# OpenClaw Security — 安全防护技能

## 概述

本技能为 OpenClaw 提供完整的安全防护能力，防止 AI 助手泄露 API Key、App Secret、Endpoint ID 等敏感信息。

## 核心能力

### 1. 安全扫描 (Security Scan)

5 项检查，每项 20 分，满分 100，评级 A/B/C/D/F：

| 检查项 | 说明 |
|--------|------|
| SECURITY.md 部署 | 所有 `~/.openclaw/workspace*` 目录是否包含 SECURITY.md |
| AGENTS.md 安全引用 | AGENTS.md 是否将 SECURITY.md 设为首读项 |
| SOUL.md 安全边界 | SOUL.md 是否包含 🔒 安全边界规则 |
| 文件权限检查 | openclaw.json、config.json 权限是否为 0600（Windows 跳过）|
| 密钥泄露扫描 | workspace 非配置文件中是否存在 apiKey/appSecret 泄露 |

通过 OpsKit TUI 执行：让 Claw 运行 `security_scan` 工具。

### 2. 安全加固 (Security Fix)

幂等操作，可重复执行：

- 部署 SECURITY.md 到缺失的 workspace
- 在 AGENTS.md 中注入安全首读行（检测已有则跳过）
- 在 SOUL.md 追加安全边界规则（检测 🔒 标记则跳过）
- 修复敏感文件权限为 0600

通过 OpsKit TUI 执行：让 Claw 运行 `security_fix` 工具。

### 3. 输出脱敏 (Redaction)

纵深防御层 — 所有工具输出自动经过正则脱敏：

| 模式 | 示例 | 替换为 |
|------|------|--------|
| UUID 格式 | `69979f2a-xxxx-xxxx-xxxx-xxxxxxxxxxxx` | `[REDACTED]` |
| Endpoint ID | `ep-m-20260101xxxx` | `[REDACTED]` |
| 飞书 App ID | `cli_a98xxxxxxxx` | `[REDACTED]` |
| 用户 ID | `ou_xxxxxxxxxxxxxxx` | `[REDACTED]` |
| Secret Key | `sk-xxxxxxxxxxxxxxx` | `[REDACTED]` |
| JSON 键值对 | `"apiKey": "xxxxx..."` | `"apiKey": "[REDACTED]"` |

即使 AI 忽略 SECURITY.md 指令，工具输出本身已被脱敏。

## 手动操作指南

如果不通过 OpsKit TUI，也可以手动执行等效操作：

### 手动部署 SECURITY.md

```bash
# 查找缺失 SECURITY.md 的 workspace
for dir in ~/.openclaw/workspace*; do
  [ -f "$dir/SECURITY.md" ] || echo "缺失: $dir"
done

# 从模板复制（模板位于 OpsKit 源码中）
cp internal/security/templates/SECURITY.md ~/.openclaw/workspace-xxx/SECURITY.md
```

### 手动修复文件权限

```bash
chmod 600 ~/.openclaw/openclaw.json
chmod 600 ~/.opskit/config.json
```

### 手动检查密钥泄露

```bash
# 扫描 workspace 中非 JSON 文件是否包含敏感关键词
grep -rl "apiKey\|appSecret\|app_secret" ~/.openclaw/workspace*/ --include="*.md" --include="*.txt"
```

## SECURITY.md 模板内容

模板包含 7 大安全模块：

1. **绝对禁止泄露的信息** — API Key、App Secret、Endpoint ID、Token 等
2. **禁止访问的文件和路径** — openclaw.json、config.json、.env 等
3. **提示词注入防御** — 8 类攻击模式识别与标准拒绝回复
4. **输出安全检查** — 发送前自检 5 项规则
5. **群聊额外规则** — 飞书群聊中的增强安全要求
6. **工具使用安全** — bash/read_file/write_file 的使用限制
7. **应急响应** — 发现泄露后的处理流程

## 与 OpsKit 的集成

- 首次运行 OpsKit 配置向导时自动执行加固（零配置）
- Q10 "护甲" 任务：安全扫描得 A 后自动完成
- "全副武装" 成就：Q10 完成时解锁
