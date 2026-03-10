package quest

import (
	"strings"
)

// Quest defines a single quest in the chain.
type Quest struct {
	ID          string
	Code        string
	Title       string
	Description string
	XPReward    int
	StageReward int // lobster stage to evolve to, -1 means no evolution
}

// QuestDisplay holds quest info for the TUI task panel.
type QuestDisplay struct {
	ID          string
	Title       string
	Description string
	Status      string // "pending", "running", "done", "locked"
	XPReward    int
}

// allQuests is the ordered list of all 9 quests.
var allQuests = []Quest{
	{ID: "Q1", Code: "hatch", Title: "孵化", Description: "环境检查与依赖安装", XPReward: 50, StageReward: -1},
	{ID: "Q2", Code: "break_shell", Title: "破壳", Description: "安装 OpenClaw", XPReward: 50, StageReward: 1},
	{ID: "Q3", Code: "feed", Title: "喂食", Description: "配置 Coding-Plan 凭证", XPReward: 80, StageReward: -1},
	{ID: "Q4", Code: "choose_food", Title: "选粮", Description: "选择并配置 AI 模型", XPReward: 50, StageReward: -1},
	{ID: "Q5", Code: "speak", Title: "开嗓", Description: "启动 OpenClaw 网关", XPReward: 50, StageReward: 2},
	{ID: "Q6", Code: "first_cry", Title: "初啼", Description: "通过 TUI 发送第一条消息", XPReward: 100, StageReward: -1},
	{ID: "Q7", Code: "nest", Title: "筑巢", Description: "配置开机自启", XPReward: 80, StageReward: -1},
	{ID: "Q8", Code: "telepathy", Title: "通灵", Description: "接入飞书机器人", XPReward: 120, StageReward: 3},
	{ID: "Q9", Code: "open_eyes", Title: "开眼", Description: "启用 Web Dashboard", XPReward: 50, StageReward: -1},
}

// questByID maps quest ID to its index in allQuests.
var questByID map[string]int

func init() {
	questByID = make(map[string]int, len(allQuests))
	for i, q := range allQuests {
		questByID[q.ID] = i
	}
}

// AllQuests returns the ordered list of all quests.
func AllQuests() []Quest {
	result := make([]Quest, len(allQuests))
	copy(result, allQuests)
	return result
}

// CheckCompletion checks if a tool execution result completes any available quest.
// Returns the quest ID if completed, empty string otherwise.
func CheckCompletion(toolName string, args map[string]interface{}, output string, questStates map[string]string) string {
	lowerOutput := strings.ToLower(output)

	for _, q := range allQuests {
		if questStates[q.ID] != "available" {
			continue
		}
		if matchQuest(q.ID, toolName, args, lowerOutput) {
			return q.ID
		}
	}
	return ""
}

func getArgString(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return strings.ToLower(s)
		}
	}
	return ""
}

func matchQuest(questID, toolName string, args map[string]interface{}, lowerOutput string) bool {
	cmd := getArgString(args, "command")
	path := getArgString(args, "path")
	content := getArgString(args, "content")

	switch questID {
	case "Q1":
		return toolName == "bash" &&
			(strings.Contains(lowerOutput, "node") ||
				strings.Contains(lowerOutput, "npm"))
	case "Q2":
		return toolName == "bash" &&
			(strings.Contains(cmd, "install.sh") ||
				(strings.Contains(lowerOutput, "openclaw") && strings.Contains(lowerOutput, "install")))
	case "Q3":
		return toolName == "write_file" &&
			strings.Contains(path, "openclaw.json") &&
			strings.Contains(lowerOutput, "successfully wrote")
	case "Q4":
		return toolName == "write_file" &&
			strings.Contains(path, "openclaw.json") &&
			containsAny(content+lowerOutput, []string{"doubao", "ark-code", "deepseek", "kimi", "glm"})
	case "Q5":
		return toolName == "bash" &&
			(strings.Contains(cmd, "gateway start") ||
				strings.Contains(cmd, "gateway restart") ||
				strings.Contains(cmd, "gateway run")) &&
			(strings.Contains(lowerOutput, "running") ||
				strings.Contains(lowerOutput, "active") ||
				strings.Contains(lowerOutput, "started"))
	case "Q6":
		return toolName == "bash" &&
			strings.Contains(cmd, "openclaw tui")
	case "Q7":
		return toolName == "bash" &&
			(strings.Contains(cmd, "launchctl") ||
				strings.Contains(cmd, "systemctl enable") ||
				strings.Contains(cmd, "gateway install"))
	case "Q8":
		return toolName == "bash" &&
			(strings.Contains(cmd, "feishu") ||
				strings.Contains(lowerOutput, "feishu") ||
				strings.Contains(lowerOutput, "飞书")) &&
			(strings.Contains(lowerOutput, "success") ||
				strings.Contains(lowerOutput, "ok") ||
				strings.Contains(lowerOutput, "done") ||
				strings.Contains(lowerOutput, "成功"))
	case "Q9":
		return toolName == "bash" &&
			(strings.Contains(cmd, "dashboard") ||
				strings.Contains(lowerOutput, "dashboard"))
	}
	return false
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// CompleteQuest marks a quest as done and unlocks the next one.
// Returns the completed Quest definition, or nil if the quest ID is invalid.
func CompleteQuest(questID string, questStates map[string]string) *Quest {
	idx, ok := questByID[questID]
	if !ok {
		return nil
	}
	questStates[questID] = "done"

	// Unlock the next quest in the chain
	if idx+1 < len(allQuests) {
		nextID := allQuests[idx+1].ID
		if questStates[nextID] == "locked" {
			questStates[nextID] = "available"
		}
	}

	q := allQuests[idx]
	return &q
}

// QuestsToDisplay converts quest states to display structs for the TUI task panel.
func QuestsToDisplay(questStates map[string]string) []QuestDisplay {
	displays := make([]QuestDisplay, 0, len(allQuests))
	for _, q := range allQuests {
		status := questStates[q.ID]
		displayStatus := mapStatus(status)
		displays = append(displays, QuestDisplay{
			ID:          q.ID,
			Title:       q.Title,
			Description: q.Description,
			Status:      displayStatus,
			XPReward:    q.XPReward,
		})
	}
	return displays
}

// mapStatus converts quest state to TUI task status.
func mapStatus(questStatus string) string {
	switch questStatus {
	case "done":
		return "done"
	case "available":
		return "pending"
	default: // "locked"
		return "pending"
	}
}
