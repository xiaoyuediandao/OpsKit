package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"opskit/internal/achievement"
	"opskit/internal/quest"
)

// Task represents a task in the task panel.
type Task struct {
	ID          int
	Title       string
	Description string
	Status      string // "pending", "running", "done", "failed"
}

// questsToTasks converts quest states to Task structs for the task panel.
func questsToTasks(questStates map[string]string) []Task {
	displays := quest.QuestsToDisplay(questStates)
	tasks := make([]Task, len(displays))
	for i, d := range displays {
		tasks[i] = Task{
			ID:          i + 1,
			Title:       fmt.Sprintf("[%s] %s", d.ID, d.Title),
			Description: d.Description,
			Status:      d.Status,
		}
	}
	return tasks
}

// viewTaskPanel renders the task panel.
func viewTaskPanel(tasks []Task, lobster Lobster, achievements []string, panelWidth int) string {
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
		// Truncate if too wide (use display width for CJK/emoji support)
		if lipgloss.Width(line) > panelWidth-2 {
			// Trim rune by rune until it fits
			runes := []rune(task.Title)
			for lipgloss.Width(line) > panelWidth-5 && len(runes) > 0 {
				runes = runes[:len(runes)-1]
				line = fmt.Sprintf(" %s %s...", iconStyle.Render(icon), titleStyle.Render(string(runes)))
			}
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render(divider) + "\n")

	// Achievements section
	sb.WriteString(lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true).
		Padding(0, 1).
		Render("成就") + "\n")

	allAch := achievement.AllAchievements()
	for _, a := range allAch {
		if achievement.IsUnlocked(a.ID, achievements) {
			line := fmt.Sprintf(" %s %s", a.Icon, a.Name)
			sb.WriteString(lipgloss.NewStyle().Foreground(colorSuccess).Render(line) + "\n")
		}
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
