package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"opskit/internal/ai"
	"opskit/internal/config"
	"opskit/internal/state"
)

// updateSetup handles key events in setup mode.
func (m Model) updateSetup(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleSetupEnter()

		default:
			var cmd tea.Cmd
			m.setupInput, cmd = m.setupInput.Update(msg)
			m.setupErr = ""
			return m, cmd
		}

	case setupDoneMsg:
		if msg.err != nil {
			m.setupErr = fmt.Sprintf("保存失败: %v", msg.err)
			m.setupStep = 3
			return m, nil
		}
		// Transition to chat mode
		m.mode = ModeChat
		m.aiClient = ai.NewClient(m.setupAPIKey, m.setupBaseURL, m.setupModel)
		m.lobster.Stage = 1
		m.lobster.Status = "active"
		m.lobster.AddXP(50)
		m.messages = append(m.messages, ChatMessage{
			Role:      "assistant",
			Content:   fmt.Sprintf("配置成功！🎉\n\n我已经准备好了，Claw 在线！\n\n你好，我是 Claw 🦞，你的 OpenClaw 运维龙虾助手！现在可以开始和我对话了。\n\n有什么我可以帮你的吗？"),
			Timestamp: time.Now(),
		})
		_ = state.Save(m.lobster.ToState())
		return m, m.focusInput()
	}

	return m, nil
}

func (m Model) handleSetupEnter() (Model, tea.Cmd) {
	switch m.setupStep {
	case 0:
		// Welcome -> Step 1: API Key
		m.setupStep = 1
		m.setupInput.Reset()
		m.setupInput.Placeholder = "输入你的方舟 API Key..."
		m.setupInput.EchoMode = textinputEchoPassword
		m.setupInput.Focus()
		return m, textinputBlink()

	case 1:
		// Validate API Key
		val := strings.TrimSpace(m.setupInput.Value())
		if val == "" {
			m.setupErr = "API Key 不能为空"
			return m, nil
		}
		m.setupAPIKey = val
		m.setupStep = 2
		m.setupInput.Reset()
		m.setupInput.EchoMode = 0 // normal
		m.setupInput.Placeholder = fmt.Sprintf("回车使用默认: %s", "https://ark.cn-beijing.volces.com/api/coding/v3")
		m.setupInput.SetValue("")
		m.setupInput.Focus()
		return m, textinputBlink()

	case 2:
		// Base URL
		val := strings.TrimSpace(m.setupInput.Value())
		if val == "" {
			val = "https://ark.cn-beijing.volces.com/api/coding/v3"
		}
		m.setupBaseURL = val
		m.setupStep = 3
		m.setupInput.Reset()
		m.setupInput.Placeholder = fmt.Sprintf("回车使用默认: %s", "doubao-seed-2.0-code")
		m.setupInput.SetValue("")
		m.setupInput.Focus()
		return m, textinputBlink()

	case 3:
		// Model
		val := strings.TrimSpace(m.setupInput.Value())
		if val == "" {
			val = "doubao-seed-2.0-code"
		}
		m.setupModel = val
		m.setupStep = 4
		m.setupInput.Blur()
		return m, m.saveSetupConfig()
	}

	return m, nil
}

func (m Model) saveSetupConfig() tea.Cmd {
	apiKey := m.setupAPIKey
	baseURL := m.setupBaseURL
	modelName := m.setupModel
	return func() tea.Msg {
		cfg := &config.Config{
			APIKey:  apiKey,
			BaseURL: baseURL,
			Model:   modelName,
		}
		err := config.Save(cfg)
		return setupDoneMsg{err: err}
	}
}

func (m Model) focusInput() tea.Cmd {
	m.input.Focus()
	return textinputBlink()
}

// viewSetup renders the setup wizard.
func (m Model) viewSetup() string {
	w := m.width
	if w < 40 {
		w = 80
	}

	var sb strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Padding(1, 2).
		Render("🦞 OpsKit 初始化向导")
	sb.WriteString(header + "\n")

	// Progress indicator
	steps := []string{"欢迎", "API Key", "Base URL", "模型", "完成"}
	var progressParts []string
	for i, step := range steps {
		style := lipgloss.NewStyle().Foreground(colorMuted)
		if i == m.setupStep {
			style = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
		} else if i < m.setupStep {
			style = lipgloss.NewStyle().Foreground(colorSuccess)
		}
		progressParts = append(progressParts, style.Render(fmt.Sprintf("%d.%s", i+1, step)))
	}
	progress := strings.Join(progressParts, lipgloss.NewStyle().Foreground(colorMuted).Render(" → "))
	sb.WriteString("  " + progress + "\n\n")

	// Content based on step
	switch m.setupStep {
	case 0:
		content := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			Width(w - 8).
			Render(
				lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render("欢迎使用 OpsKit！") + "\n\n" +
					"我是 Claw 🦞，你的 OpenClaw 智能运维助手。\n\n" +
					"在开始之前，我需要你提供方舟 API Key，以便我能够与 AI 模型通信。\n\n" +
					"你可以在以下地址获取 API Key：\n" +
					lipgloss.NewStyle().Foreground(colorSecondary).Render("  https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement\n\n") +
					lipgloss.NewStyle().Foreground(colorMuted).Render("按 Enter 开始配置..."),
			)
		sb.WriteString(content + "\n")

	case 1:
		sb.WriteString(
			lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render("  步骤 1：输入方舟 API Key") + "\n\n",
		)
		sb.WriteString(
			lipgloss.NewStyle().Foreground(colorMuted).Render("  API Key 格式通常为: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx") + "\n\n",
		)
		sb.WriteString("  > " + m.setupInput.View() + "\n")
		if m.setupErr != "" {
			sb.WriteString("\n  " + lipgloss.NewStyle().Foreground(colorError).Render("✗ "+m.setupErr) + "\n")
		}

	case 2:
		sb.WriteString(
			lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render("  步骤 2：Base URL") + "\n\n",
		)
		sb.WriteString(
			lipgloss.NewStyle().Foreground(colorMuted).Render("  默认使用方舟 Coding Plan 端点，直接回车即可") + "\n\n",
		)
		sb.WriteString("  > " + m.setupInput.View() + "\n")

	case 3:
		sb.WriteString(
			lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render("  步骤 3：选择模型") + "\n\n",
		)
		sb.WriteString(
			lipgloss.NewStyle().Foreground(colorMuted).Render("  可选模型：doubao-seed-2.0-code, doubao-seed-2.0-pro, ark-code-latest") + "\n\n",
		)
		sb.WriteString("  > " + m.setupInput.View() + "\n")
		if m.setupErr != "" {
			sb.WriteString("\n  " + lipgloss.NewStyle().Foreground(colorError).Render("✗ "+m.setupErr) + "\n")
		}

	case 4:
		sb.WriteString(
			lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Render("  正在保存配置...") + "\n\n",
		)
		sb.WriteString(
			lipgloss.NewStyle().Foreground(colorMuted).Render("  🦞 Claw 正在整理小钳子...") + "\n",
		)
	}

	// Help text
	if m.setupStep < 4 {
		sb.WriteString("\n" + lipgloss.NewStyle().Foreground(colorMuted).Render("  Enter: 确认  Ctrl+C: 退出"))
	}

	return sb.String()
}

// textinputEchoPassword is the password echo mode constant.
const textinputEchoPassword = textinput.EchoPassword

func textinputBlink() tea.Cmd {
	return textinput.Blink
}
