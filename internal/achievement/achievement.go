package achievement

import "fmt"

// Achievement defines a single achievement.
type Achievement struct {
	ID   string
	Icon string
	Name string
	Desc string // short description of unlock condition
}

// allAchievements is the ordered list of all 9 achievements.
var allAchievements = []Achievement{
	{ID: "first_install", Icon: "🥇", Name: "破壳而出", Desc: "完成首次安装"},
	{ID: "first_tool", Icon: "🔧", Name: "工具人", Desc: "执行了第一个自动化操作"},
	{ID: "key_config", Icon: "🔑", Name: "钥匙匠", Desc: "成功配置 API Key"},
	{ID: "revive", Icon: "🩺", Name: "急救医生", Desc: "龙虾 HP 降到 0 后成功复活"},
	{ID: "rollback", Icon: "🔄", Name: "时光旅人", Desc: "执行第一次版本回滚"},
	{ID: "full_armor", Icon: "🦞⚔️", Name: "全副武装", Desc: "配置完所有安全策略"},
	{ID: "log_addict", Icon: "📊", Name: "数据控", Desc: "查看过 10 次以上的日志分析"},
	{ID: "social", Icon: "🌐", Name: "社交达人", Desc: "同时接入 3 个以上通道"},
	{ID: "king", Icon: "👑", Name: "龙虾之王", Desc: "完成全部任务，达到最高等级"},
}

// AllAchievements returns all achievement definitions.
func AllAchievements() []Achievement {
	result := make([]Achievement, len(allAchievements))
	copy(result, allAchievements)
	return result
}

// Check checks if any new achievements should be unlocked based on current state.
// Parameters:
//   - event: a trigger event string (e.g., "quest_complete", "tool_use", "revive", "rollback", "full_armor", "social")
//   - questStates: map of quest ID -> status
//   - unlocked: currently unlocked achievement IDs
//   - toolUseCount: total tool uses
//   - logViewCount: total log views
//
// Returns a list of newly unlocked Achievement definitions (may be empty).
func Check(event string, questStates map[string]string, unlocked []string, toolUseCount int, logViewCount int) []Achievement {
	// Build a set of already-unlocked IDs for fast lookup
	unlockedSet := make(map[string]bool, len(unlocked))
	for _, id := range unlocked {
		unlockedSet[id] = true
	}

	var newly []Achievement
	for _, a := range allAchievements {
		if unlockedSet[a.ID] {
			continue
		}
		if checkCondition(a.ID, event, questStates, toolUseCount, logViewCount) {
			newly = append(newly, a)
		}
	}
	return newly
}

// IsUnlocked checks if a specific achievement ID is in the unlocked list.
func IsUnlocked(id string, unlocked []string) bool {
	for _, uid := range unlocked {
		if uid == id {
			return true
		}
	}
	return false
}

func checkCondition(id string, event string, questStates map[string]string, toolUseCount int, logViewCount int) bool {
	switch id {
	case "first_install":
		return questStates["Q2"] == "done"
	case "first_tool":
		return toolUseCount >= 1
	case "key_config":
		return questStates["Q3"] == "done"
	case "revive":
		return event == "revive"
	case "rollback":
		return event == "rollback"
	case "full_armor":
		return questStates["Q10"] == "done"
	case "log_addict":
		return logViewCount >= 10
	case "social":
		return event == "social"
	case "king":
		return allQuestsDone(questStates)
	}
	return false
}

func allQuestsDone(quests map[string]string) bool {
	for i := 1; i <= 10; i++ {
		id := fmt.Sprintf("Q%d", i)
		if quests[id] != "done" {
			return false
		}
	}
	return true
}
