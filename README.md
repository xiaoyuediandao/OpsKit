# OpsKit

OpsKit 是一个基于终端 UI 的 AI 运维助手，专为 [OpenClaw](https://openclaw.ai) 的全生命周期管理而设计。

内置智能龙虾 **Claw 🦞**，帮助你完成 OpenClaw 的安装、配置、服务管理、渠道接入和故障排查。

![Go](https://img.shields.io/badge/Go-1.24-blue)
![TUI](https://img.shields.io/badge/TUI-Bubble%20Tea-pink)
![AI](https://img.shields.io/badge/AI-Volcengine%20Ark-orange)

## 功能特性

- **智能对话** — 与 Claw 对话，用自然语言完成运维操作
- **工具执行** — AI 可直接调用工具（执行命令、读写文件、查看目录），每次执行前需用户确认
- **流式响应** — 实时展示 AI 思考过程
- **Markdown 渲染** — 终端内渲染格式化输出
- **龙虾成长系统** — Claw 拥有 HP、XP、等级，随使用成长
- **初始化向导** — 首次运行自动引导配置 API Key 和模型

### Claw 能帮你做什么

- 安装 OpenClaw（引导用户在终端执行）
- 配置 API Key 和模型（写入 `~/.openclaw/openclaw.json`）
- 管理服务（`openclaw gateway start/stop/restart/status`）
- 接入飞书、微信等渠道
- 故障排查（查日志、检查端口、重启服务）
- 版本升级

## 安装

### 前置条件

- Go 1.24+
- 方舟 API Key（[获取地址](https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement)）

### 从源码构建

```bash
git clone https://github.com/xiaoyuediandao/opskit.git
cd OpsKit
go build -o opskit .
```

### 运行

```bash
./opskit
```

首次运行会启动配置向导，输入你的方舟 API Key 即可开始使用。

## 配置

配置文件存储在 `~/.opskit/config.json`，包含以下字段：

```json
{
  "api_key": "your-api-key",
  "base_url": "https://ark.cn-beijing.volces.com/api/coding/v3",
  "model": "doubao-seed-2.0-code"
}
```

**API Key 永远不会提交到代码仓库**，仅存储在本地配置文件中。

### 支持的模型

| 模型 | 说明 |
|------|------|
| `doubao-seed-2.0-code` | 默认，代码能力强 |
| `doubao-seed-2.0-pro` | 综合能力强 |
| `ark-code-latest` | 最新代码模型 |
| `kimi-k2.5` | Kimi 系列 |
| `deepseek-v3.2` | DeepSeek 系列 |

## 使用

| 按键 | 功能 |
|------|------|
| `Enter` | 发送消息 / 确认工具执行 |
| `Esc` | 取消待执行的工具调用 |
| `T` | 显示/隐藏任务面板 |
| `Ctrl+C` | 退出 |

## 项目结构

```
OpsKit/
├── main.go                  # 入口，初始化 TUI
├── internal/
│   ├── ai/
│   │   └── client.go        # AI 客户端（流式/工具调用）
│   ├── config/
│   │   └── config.go        # 配置读写（~/.opskit/config.json）
│   ├── exec/
│   │   ├── runner.go        # Shell 命令执行器
│   │   └── config.go        # OpenClaw 配置写入
│   ├── state/
│   │   └── state.go         # 龙虾状态持久化
│   ├── tools/
│   │   ├── tools.go         # 工具定义与分发
│   │   ├── bash.go          # bash 工具
│   │   ├── read.go          # 文件读取工具
│   │   ├── write.go         # 文件写入工具
│   │   └── list.go          # 目录列举工具
│   └── tui/
│       ├── model.go         # Bubble Tea 主模型
│       ├── update.go        # 事件更新逻辑
│       ├── view.go          # UI 渲染
│       ├── setup.go         # 初始化向导
│       ├── tasks.go         # 任务面板
│       ├── lobster.go       # 龙虾成长系统
│       └── markdown.go      # Markdown 渲染
└── go.mod
```

## 技术栈

- **语言**: Go 1.24
- **TUI 框架**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **AI 后端**: Volcengine Ark（兼容 OpenAI API 格式）
- **Markdown**: [Glamour](https://github.com/charmbracelet/glamour)

## License

MIT
