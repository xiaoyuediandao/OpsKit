package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Task represents a task in the task panel.
type Task struct {
	ID          int
	Title       string
	Description string
	Status      string // "pending", "running", "done", "failed"
}

// defaultTasks returns a set of standard OpenClaw tasks.
func defaultTasks() []Task {
	return []Task{
		{ID: 1, Title: "安装 OpenClaw", Status: "pending", Description: "curl -fsSL https://openclaw.ai/install.sh | bash"},
		{ID: 2, Title: "配置 API Key", Status: "pending", Description: "写入 ~/.openclaw/openclaw.json"},
		{ID: 3, Title: "启动 Gateway", Status: "pending", Description: "openclaw gateway start"},
		{ID: 4, Title: "接入飞书", Status: "pending", Description: "安装飞书插件并配对"},
		{ID: 5, Title: "加载行业模板", Status: "pending", Description: "选择适合的行业模板"},
	}
}

// viewTaskPanel renders the task panel.
func viewTaskPanel(tasks []Task, lobster Lobster, panelWidth int) string {
	if panelWidth < 20 {
		panelWidth = 30
	}

	var sb strings.Builder

	// Panel header
	header := lipgloss.NewStyle().
		Background(colorPrimary).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1).
		Width(panelWidth - 2).
		Render(fmt.Sprintf(" %s %s  Lv.%d", lobster.StageEmoji(), lobster.StageName(), lobster.Level))
	sb.WriteString(header + "\n")

	// Lobster stats
	statsStyle := lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(0, 1)

	sb.WriteString(statsStyle.Render(fmt.Sprintf("HP: %s", lobster.HPBar())) + "\n")
	sb.WriteString(statsStyle.Render(fmt.Sprintf("XP: %s", lobster.XPBar())) + "\n")

	divider := strings.Repeat("─", panelWidth-2)
	sb.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(divider) + "\n")

	// Tasks header
	sb.WriteString(lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true).
		Padding(0, 1).
		Render("任务清单") + "\n")

	// Tasks list
	for _, task := range tasks {
		icon := taskStatusIcon(task.Status)
		iconStyle := taskStatusStyle(task.Status)

		titleStyle := lipgloss.NewStyle().Foreground(colorUser)
		if task.Status == "done" {
			titleStyle = lipgloss.NewStyle().Foreground(colorMuted)
		}

		line := fmt.Sprintf(" %s %s", iconStyle.Render(icon), titleStyle.Render(task.Title))
		// Truncate if too long
		if len(line) > panelWidth-2 {
			line = line[:panelWidth-5] + "..."
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(divider) + "\n")

	// Help
	sb.WriteString(lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(0, 1).
		Render("T: 关闭面板") + "\n")

	return sb.String()
}

func taskStatusIcon(status string) string {
	switch status {
	case "done":
		return "✓"
	case "running":
		return "◉"
	case "failed":
		return "✗"
	default:
		return "○"
	}
}

func taskStatusStyle(status string) lipgloss.Style {
	switch status {
	case "done":
		return lipgloss.NewStyle().Foreground(colorSuccess)
	case "running":
		return lipgloss.NewStyle().Foreground(colorSecondary)
	case "failed":
		return lipgloss.NewStyle().Foreground(colorError)
	default:
		return lipgloss.NewStyle().Foreground(colorMuted)
	}
}
