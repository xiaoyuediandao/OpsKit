package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"opskit/internal/ai"
	"opskit/internal/tools"
)

// Mode represents the app mode.
type Mode int

const (
	ModeSetup Mode = iota
	ModeChat
)

// ChatMessage represents a message in the chat history.
type ChatMessage struct {
	Role      string // "user", "assistant", "system", "tool_preview", "tool_result"
	Content   string
	ToolName  string
	Timestamp time.Time
}

// Message types for Bubble Tea

type aiResponseMsg struct {
	content string
	err     error
}

type streamChunkMsg struct {
	chunk string
}

type streamDoneMsg struct {
	err error
}

type cmdOutputMsg struct {
	output string
	err    error
}

type execProcessDoneMsg struct{}

type agentThinkDoneMsg struct {
	resp *ai.ToolChatResponse
	err  error
}

type agentToolDoneMsg struct {
	toolCallID string
	toolName   string
	result     tools.ToolResult
}

type tickMsg time.Time

type setupDoneMsg struct {
	err error
}

// Model is the main Bubble Tea model.
type Model struct {
	mode    Mode
	ready   bool
	width   int
	height  int
	errMsg  string

	// Chat
	messages []ChatMessage
	history  []ai.Message // kept for simple streaming mode reference (not used in agent mode)
	viewport viewport.Model

	// Input
	input textinput.Model

	// Spinner
	spinner   spinner.Model
	loading   bool
	executing bool
	tickCount int

	// Task panel
	showTaskPanel bool

	// Lobster
	lobster Lobster

	// AI client
	aiClient *ai.Client

	// Streaming
	streamBuf strings.Builder
	chunkCh   chan string

	// Agent loop
	agentHistory    []ai.AgentMessage
	pendingToolCall *ai.ToolCall
	agentLoopDepth  int
	agentThinking   bool

	// Setup mode fields
	setupStep    int
	setupInput   textinput.Model
	setupAPIKey  string
	setupBaseURL string
	setupModel   string
	setupErr     string
}

// InitialModel creates the initial model in either Setup or Chat mode.
func InitialModel(aiClient *ai.Client, lobster Lobster, startInSetup bool) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = styleMuted

	inp := textinput.New()
	inp.Placeholder = "Ask Claw anything..."
	inp.Focus()
	inp.CharLimit = 2000
	inp.Width = 60

	setupInp := textinput.New()
	setupInp.CharLimit = 500

	lobster.Status = "idle"

	m := Model{
		spinner:      sp,
		input:        inp,
		setupInput:   setupInp,
		lobster:      lobster,
		aiClient:     aiClient,
		showTaskPanel: false,
	}

	if startInSetup {
		m.mode = ModeSetup
		m.setupStep = 0
		m.setupBaseURL = "https://ark.cn-beijing.volces.com/api/coding/v3"
		m.setupModel = "doubao-seed-2.0-code"
	} else {
		m.mode = ModeChat
		// Add a welcome message
		m.messages = append(m.messages, ChatMessage{
			Role:      "assistant",
			Content:   "你好！我是 Claw 🦞，你的 OpenClaw 运维助手！有什么我可以帮你的吗？\n\n你可以问我：\n- 如何安装 OpenClaw？\n- 如何配置 API Key？\n- 如何接入飞书？\n- 服务状态检查...",
			Timestamp: time.Now(),
		})
	}

	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		textinput.Blink,
		tea.WindowSize(), // 立即触发 WindowSizeMsg，避免卡在"正在初始化"
		tea.Every(100*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}
