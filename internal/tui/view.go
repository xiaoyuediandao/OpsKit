package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const taskPanelWidth = 32

// Color theme
var (
	colorPrimary   = lipgloss.Color("#FF6B35")
	colorSecondary = lipgloss.Color("#5BC8E8")
	colorSuccess   = lipgloss.Color("#2ED573")
	colorError     = lipgloss.Color("#FF4757")
	colorMuted     = lipgloss.Color("#6C757D")
	colorUser      = lipgloss.Color("#F8F9FA")
)

// Styles
var (
	styleHeader = lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 2)

	styleUserMsg = lipgloss.NewStyle().
			Foreground(colorUser).
			Background(lipgloss.Color("#2D3748")).
			Padding(0, 1).
			MarginLeft(2)

	styleAssistantMsg = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E2E8F0")).
				PaddingLeft(2)

	styleSystemMsg = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	styleToolPreview = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorSecondary).
				Padding(0, 1).
				MarginLeft(2)

	styleToolResult = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(colorSuccess).
			Padding(0, 1).
			MarginLeft(2)

	styleToolResultErr = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorError).
				Padding(0, 1).
				MarginLeft(2)

	styleMuted = lipgloss.NewStyle().Foreground(colorMuted)

	styleInput = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	stylePendingConfirm = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Bold(true)
)

// View implements tea.Model.
func (m Model) View() string {
	if m.mode == ModeSetup {
		return m.viewSetup()
	}

	if !m.ready {
		return "\n  正在初始化 OpsKit..."
	}

	return m.viewChat()
}

func (m Model) viewChat() string {
	var sb strings.Builder

	// Header
	sb.WriteString(m.renderHeader())
	sb.WriteString("\n")

	// Main content area
	if m.showTaskPanel {
		// Split layout: chat + task panel
		chatContent := m.viewport.View()
		taskContent := viewTaskPanel(m.tasks, m.lobster, m.achievements, taskPanelWidth)

		// Render side by side
		combined := lipgloss.JoinHorizontal(
			lipgloss.Top,
			chatContent,
			lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorMuted).
				Render(taskContent),
		)
		sb.WriteString(combined)
	} else {
		sb.WriteString(m.viewport.View())
	}

	sb.WriteString("\n")

	// Footer / Input area
	sb.WriteString(m.renderFooter())

	return sb.String()
}

func (m Model) renderHeader() string {
	lobsterInfo := fmt.Sprintf("%s Claw  %s %s  Lv.%d  HP:%d/%d",
		m.lobster.StageEmoji(),
		m.lobster.StatusIcon(),
		m.lobster.StageName(),
		m.lobster.Level,
		m.lobster.HP,
		m.lobster.MaxHP,
	)

	width := m.width
	if width < 40 {
		width = 40
	}

	title := "OpsKit"
	rightInfo := lobsterInfo

	// Calculate spacing
	titleWidth := width / 2
	rightWidth := width - titleWidth - 4

	titlePart := styleHeader.Width(titleWidth).Render(title)
	infoPart := lipgloss.NewStyle().
		Background(lipgloss.Color("#1A202C")).
		Foreground(colorPrimary).
		Bold(true).
		Padding(0, 1).
		Width(rightWidth).
		Align(lipgloss.Right).
		Render(rightInfo)

	return titlePart + infoPart
}

func (m Model) renderFooter() string {
	width := m.width
	if m.showTaskPanel {
		width = m.width - taskPanelWidth - 1
	}
	if width < 20 {
		width = 20
	}

	// Ensure input width is set properly for this render (safe: value receiver = local copy)
	inputWidth := width - 8
	if inputWidth < 10 {
		inputWidth = 10
	}
	m.input.Width = inputWidth

	var statusLine string
	if m.pendingToolCall != nil {
		statusLine = stylePendingConfirm.Render("  Enter=confirm / Esc=cancel tool call")
	} else if m.agentThinking || m.loading {
		dots := strings.Repeat(".", (m.tickCount/3)%4)
		statusLine = styleMuted.Render(fmt.Sprintf("%s Thinking%s", m.spinner.View(), dots))
	} else if m.executing {
		statusLine = styleMuted.Render(fmt.Sprintf("%s Running command...", m.spinner.View()))
	} else {
		statusLine = styleMuted.Render("Enter=send  T=tasks  Ctrl+C=quit")
	}

	inputLine := styleInput.Width(width - 4).Render(m.input.View())

	return statusLine + "\n" + inputLine
}

func (m Model) renderMessages() string {
	var sb strings.Builder

	viewportW := m.viewport.Width
	if viewportW < 20 {
		viewportW = 80
	}

	for _, msg := range m.messages {
		sb.WriteString(renderMessage(msg, viewportW))
		sb.WriteString("\n")
	}

	// Show thinking indicator at the bottom
	if (m.agentThinking || m.loading) && m.pendingToolCall == nil {
		dots := strings.Repeat(".", (m.tickCount/3)%4)
		thinking := styleMuted.Render(fmt.Sprintf("\n🦞 Claw 正在思考%s", dots))
		sb.WriteString(thinking)
	}

	return sb.String()
}

func renderMessage(msg ChatMessage, width int) string {
	timeStr := msg.Timestamp.Format("15:04")
	contentWidth := width - 6
	if contentWidth < 20 {
		contentWidth = 20
	}

	switch msg.Role {
	case "user":
		label := lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			Render("  💬 你") + "  " + styleMuted.Render(timeStr)
		content := styleUserMsg.Width(contentWidth).Render(msg.Content)
		return "\n" + label + "\n" + content

	case "assistant", "assistant_streaming":
		label := lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Render("  🦞 Claw") + "  " + styleMuted.Render(timeStr)
		rendered := renderMarkdown(msg.Content)
		rendered = strings.TrimRight(rendered, "\n")
		content := styleAssistantMsg.Width(contentWidth).Render(rendered)
		return "\n" + label + "\n" + content

	case "system":
		return styleSystemMsg.PaddingLeft(2).Width(contentWidth).Render(fmt.Sprintf("⚙ %s", msg.Content))

	case "tool_preview":
		label := lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			PaddingLeft(2).
			Render(fmt.Sprintf("🔧 工具预览 [%s]", msg.ToolName))
		rendered := renderMarkdown(msg.Content)
		rendered = strings.TrimRight(rendered, "\n")
		content := styleToolPreview.Width(contentWidth).Render(rendered)
		return "\n" + label + "\n" + content

	case "tool_result":
		isErr := msg.IsError
		label := lipgloss.NewStyle().Bold(true).PaddingLeft(2)
		if isErr {
			label = label.Foreground(colorError)
		} else {
			label = label.Foreground(colorSuccess)
		}

		labelStr := label.Render(fmt.Sprintf("📋 执行结果 [%s]", msg.ToolName)) + "  " + styleMuted.Render(timeStr)

		content := msg.Content
		if len(content) > 3000 {
			content = content[:3000] + "\n...(输出已截断)"
		}

		var resultStyle lipgloss.Style
		if isErr {
			resultStyle = styleToolResultErr
		} else {
			resultStyle = styleToolResult
		}
		rendered := resultStyle.Width(contentWidth).Render(content)
		return "\n" + labelStr + "\n" + rendered

	default:
		return styleMuted.Render(msg.Content)
	}
}

