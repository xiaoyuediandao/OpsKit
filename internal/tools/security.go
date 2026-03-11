package tools

import (
	"fmt"
	"strings"

	"opskit/internal/security"
)

// RunSecurityScan runs the security scanner and returns a formatted markdown report.
func RunSecurityScan() ToolResult {
	report := security.Scan()

	var sb strings.Builder
	sb.WriteString("# 🔒 OpsKit 安全扫描报告\n\n")
	sb.WriteString(fmt.Sprintf("**安全评分: %d/%d (Grade: %s)**\n\n", report.TotalScore, report.MaxScore, report.Grade))
	sb.WriteString("| 检查项 | 状态 | 得分 | 详情 |\n")
	sb.WriteString("|--------|------|------|------|\n")

	for _, r := range report.Results {
		status := "❌"
		if r.Passed {
			status = "✅"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d/20 | %s |\n", r.Name, status, r.Score, r.Details))
	}

	sb.WriteString(fmt.Sprintf("\n**总分: %d/100 — 等级: %s**\n", report.TotalScore, report.Grade))

	return ToolResult{Output: sb.String()}
}

// RunSecurityFix runs the security hardener and returns a formatted report.
func RunSecurityFix() ToolResult {
	results := security.Harden()

	var sb strings.Builder
	sb.WriteString("# 🛡️ OpsKit 安全加固报告\n\n")

	successCount := 0
	for _, r := range results {
		status := "❌"
		if r.Success {
			status = "✅"
			successCount++
		}
		sb.WriteString(fmt.Sprintf("- %s **%s**: %s\n", status, r.Action, r.Details))
	}

	sb.WriteString(fmt.Sprintf("\n**完成: %d/%d 项操作成功**\n", successCount, len(results)))
	sb.WriteString("\n> 💡 运行 `security_scan` 验证加固效果\n")

	return ToolResult{Output: sb.String()}
}
