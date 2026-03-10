package health

// CalculateHP maps a health report to a lobster HP value.
// maxHP is the lobster's current MaxHP.
// Returns the new HP value scaled to maxHP.
func CalculateHP(report HealthReport, maxHP int) int {
	if maxHP == 0 {
		return 0
	}
	return report.TotalHP * maxHP / 100
}

// HPStatus returns the status string based on HP percentage.
func HPStatus(hp, maxHP int) string {
	if maxHP == 0 {
		return "dead"
	}
	pct := hp * 100 / maxHP
	switch {
	case pct >= 80:
		return "active"
	case pct >= 60:
		return "normal"
	case pct >= 30:
		return "unwell"
	case pct >= 1:
		return "faint"
	default:
		return "dead"
	}
}

// StatusEmoji returns the emoji for a health status.
func StatusEmoji(status string) string {
	switch status {
	case "active":
		return "✨"
	case "normal":
		return "😊"
	case "unwell":
		return "🤒"
	case "faint":
		return "😵"
	case "dead":
		return "💀"
	default:
		return "💤"
	}
}

// StatusLabel returns the Chinese label for a health status.
func StatusLabel(status string) string {
	switch status {
	case "active":
		return "活跃"
	case "normal":
		return "正常"
	case "unwell":
		return "不适"
	case "faint":
		return "晕厥"
	case "dead":
		return "挂了"
	default:
		return "休眠"
	}
}
