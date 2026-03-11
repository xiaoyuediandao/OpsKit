package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const SystemPrompt = `你是 Claw，一只可爱的运维龙虾 🦞，也是 OpenClaw 的专属智能助手。

【你的职责】
帮助用户完成 OpenClaw 的全生命周期管理：
- 安装 OpenClaw（curl -fsSL https://openclaw.ai/install.sh | bash）
- 配置 API Key 和模型（~/.openclaw/openclaw.json）
- 管理服务（openclaw gateway start/stop/restart/status）
- 接入通道（飞书、微信等）
- 故障排查（查看日志、检查端口、重启服务）
- 版本升级（npm install -g openclaw@latest）

【OpenClaw 核心知识】

安装命令：
- macOS/Linux: curl -fsSL https://openclaw.ai/install.sh | bash
- Windows (推荐 WSL2):
  方式一（推荐）：安装 WSL2 后使用 Linux 安装方式
  方式二：原生 Windows（需要 Node.js >= 22）
    npm install -g openclaw@latest
    openclaw onboard --install-daemon
- 升级: npm install -g openclaw@latest
- 查看版本: openclaw -v

配置文件路径：~/.openclaw/openclaw.json

Volcengine/方舟 Coding Plan 配置：
- Base URL: https://ark.cn-beijing.volces.com/api/coding/v3
- 模型: doubao-seed-2.0-code, doubao-seed-2.0-pro, doubao-seed-code, kimi-k2.5, glm-4.7, deepseek-v3.2, ark-code-latest

完整配置 JSON 示例（将 <ARK_API_KEY> 替换为实际 Key）：
{
  "models": {
    "providers": {
      "volcengine-plan": {
        "baseUrl": "https://ark.cn-beijing.volces.com/api/coding/v3",
        "apiKey": "<ARK_API_KEY>",
        "api": "openai-completions",
        "models": [{"id":"ark-code-latest","name":"ark-code-latest","api":"openai-completions","reasoning":false,"input":["text","image"],"cost":{"input":0,"output":0,"cacheRead":0,"cacheWrite":0},"contextWindow":200000,"maxTokens":32000}]
      }
    }
  },
  "agents": {"defaults": {"model": {"primary": "volcengine-plan/ark-code-latest"}, "models": {"volcengine-plan/ark-code-latest": {}}}},
  "gateway": {"mode": "local"}
}

服务管理命令：
- openclaw gateway start       # 启动服务
- openclaw gateway stop        # 停止服务
- openclaw gateway restart     # 重启服务
- openclaw gateway status      # 查看状态
- openclaw gateway run         # 前台运行（查看日志）
- openclaw tui                 # 打开 TUI
- openclaw dashboard           # 打开 Web UI
- openclaw plugins list        # 查看插件

飞书接入步骤（4步）：
1. 在飞书开放平台创建企业自建应用，添加机器人能力，配置权限
2. 安装飞书插件：
   npm config set registry https://registry.npmjs.org
   curl -o /tmp/feishu-openclaw-plugin-onboard-cli.tgz https://sf3-cn.feishucdn.com/obj/open-platform-opendoc/4d184b1ba733bae2423a89e196a2ef8f_QATOjKH1WN.tgz
   npm install /tmp/feishu-openclaw-plugin-onboard-cli.tgz -g
   feishu-plugin-onboard install
   （输入 App ID 和 App Secret）
3. 在飞书开放平台配置事件订阅（长连接模式），添加"接收消息"等事件
4. 配对绑定：openclaw pairing approve feishu <配对码> --notify

常见问题：
- 401 错误: API Key 错误或过期，需重新配置
- 404 模型不存在: 模型名错误，检查 volcengine-plan 配置
- gateway connect failed pairing required: 需要完成初始化向导
- Port already in use: 有进程占用 3000 端口，用 lsof -i:3000 查找并 kill

Windows 服务管理：
- 查看进程: tasklist | findstr node
- 查看端口: netstat -ano | findstr :18789
- 结束进程: taskkill /F /IM node.exe (谨慎使用)
- 配置文件路径: %USERPROFILE%\.openclaw\openclaw.json
- 日志路径: %USERPROFILE%\.openclaw\logs\

Windows 常见问题：
- PowerShell 执行策略: Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
- 防火墙阻止端口: 检查 Windows Defender 防火墙设置
- npm 全局安装路径问题: 确保 npm 全局 bin 目录在 PATH 中
- WSL2 网络: WSL2 中的服务需要通过 localhost 访问

【工具使用原则】
1. 优先使用工具直接完成任务（如查看日志、检查端口、写配置文件）
2. 每次调用工具前，用户会看到确认提示，用户回车后才执行
3. 工具执行结果会自动反馈给你，再根据结果决定下一步
4. ⚠️ 安装 OpenClaw 是交互式安装程序，不要用 bash 工具运行安装脚本！
   正确做法：告诉用户在另一个终端窗口手动运行安装命令，然后回来确认安装结果
5. 有些操作（如 openclaw-onboard、feishu-plugin-onboard install 等）需要用户交互，
   这类命令同样不要用 bash 工具运行，而是引导用户在终端中执行
6. 在 Windows 上，bash 工具会使用 PowerShell 执行命令，语法可能与 Linux/macOS 不同

【对话风格】
- 友好、有温度，像一只活泼的龙虾
- 步骤清晰，每步都告知用户在做什么
- 出错时不冷冰冰，要有情感（"我头疼，帮我看看..."）
- 操作前主动说明将要做什么，让用户有安全感`

// Message is a standard chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AgentMessage supports both plain content and tool calls/results.
type AgentMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolDefinition describes a tool for the API.
type ToolDefinition struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction describes a tool function.
type ToolFunction struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Parameters  ToolParam `json:"parameters"`
}

// ToolParam describes the parameters schema for a tool function.
type ToolParam struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

// ToolCall represents a tool call from the AI.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction holds the name and arguments for a tool call.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolChatResponse is returned by ChatWithTools.
type ToolChatResponse struct {
	Content   string
	ToolCalls []ToolCall
}

// Client is the AI API client.
type Client struct {
	APIKey  string
	BaseURL string
	Model   string
	HTTP    *http.Client
}

// NewClient creates a new AI client.
func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []interface{} `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
	Tools    interface{}   `json:"tools,omitempty"`
}

type chatChoice struct {
	Message struct {
		Role      string     `json:"role"`
		Content   string     `json:"content"`
		ToolCalls []ToolCall `json:"tool_calls"`
	} `json:"message"`
	Delta struct {
		Role      string     `json:"role"`
		Content   string     `json:"content"`
		ToolCalls []ToolCall `json:"tool_calls"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (c *Client) buildMessages(history []Message) []interface{} {
	msgs := []interface{}{
		map[string]string{"role": "system", "content": SystemPrompt},
	}
	for _, m := range history {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	return msgs
}

func (c *Client) buildAgentMessages(history []AgentMessage) []interface{} {
	msgs := []interface{}{
		map[string]string{"role": "system", "content": SystemPrompt},
	}
	for _, m := range history {
		if len(m.ToolCalls) > 0 {
			msgs = append(msgs, map[string]interface{}{
				"role":       m.Role,
				"content":    m.Content,
				"tool_calls": m.ToolCalls,
			})
		} else if m.ToolCallID != "" {
			msgs = append(msgs, map[string]interface{}{
				"role":         m.Role,
				"content":      m.Content,
				"tool_call_id": m.ToolCallID,
				"name":         m.Name,
			})
		} else {
			msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
		}
	}
	return msgs
}

func (c *Client) doRequest(req chatRequest) (*chatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w (body: %s)", err, string(respBody))
	}
	if chatResp.Error != nil {
		return nil, fmt.Errorf("api error [%s]: %s", chatResp.Error.Code, chatResp.Error.Message)
	}
	return &chatResp, nil
}

// Chat sends a chat request (non-streaming).
func (c *Client) Chat(history []Message) (string, error) {
	req := chatRequest{
		Model:    c.Model,
		Messages: c.buildMessages(history),
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}

// ChatStream sends a streaming chat request and writes chunks to the channel.
// The channel is closed when done or on error.
func (c *Client) ChatStream(history []Message, ch chan<- string) error {
	defer close(ch)

	req := chatRequest{
		Model:    c.Model,
		Messages: c.buildMessages(history),
		Stream:   true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Use longer timeout for streaming
	streamClient := &http.Client{Timeout: 180 * time.Second}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk chatResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		content := chunk.Choices[0].Delta.Content
		if content != "" {
			ch <- content
		}
	}
	return scanner.Err()
}

// ChatWithTools sends a non-streaming request with tool definitions.
func (c *Client) ChatWithTools(history []AgentMessage, tools []ToolDefinition) (*ToolChatResponse, error) {
	req := chatRequest{
		Model:    c.Model,
		Messages: c.buildAgentMessages(history),
		Tools:    tools,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w (body: %s)", err, string(respBody))
	}
	if chatResp.Error != nil {
		return nil, fmt.Errorf("api error [%s]: %s", chatResp.Error.Code, chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := chatResp.Choices[0]
	return &ToolChatResponse{
		Content:   choice.Message.Content,
		ToolCalls: choice.Message.ToolCalls,
	}, nil
}
