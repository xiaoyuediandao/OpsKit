package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HardenResult holds the result of a single hardening action.
type HardenResult struct {
	Action  string
	Success bool
	Details string
}

// Harden performs all security hardening actions. Idempotent — safe to run multiple times.
func Harden() []HardenResult {
	home, _ := os.UserHomeDir()
	var results []HardenResult

	results = append(results, deploySecurityMD(home)...)
	results = append(results, injectAgentsMD(home)...)
	results = append(results, patchSoulMD(home)...)
	results = append(results, fixPermissions(home)...)

	return results
}

// deploySecurityMD deploys SECURITY.md to all workspaces that are missing it.
func deploySecurityMD(home string) []HardenResult {
	workspaces := findWorkspaces(home)
	if len(workspaces) == 0 {
		return []HardenResult{{Action: "部署 SECURITY.md", Success: false, Details: "未找到 workspace 目录"}}
	}

	var results []HardenResult
	for _, ws := range workspaces {
		secPath := filepath.Join(ws, "SECURITY.md")
		if _, err := os.Stat(secPath); err == nil {
			// Already exists — check if it needs updating by comparing size
			existing, _ := os.ReadFile(secPath)
			if len(existing) > 0 {
				results = append(results, HardenResult{
					Action:  "部署 SECURITY.md",
					Success: true,
					Details: fmt.Sprintf("%s — 已存在，跳过", filepath.Base(ws)),
				})
				continue
			}
		}
		err := os.WriteFile(secPath, []byte(SecurityMDTemplate), 0644)
		if err != nil {
			results = append(results, HardenResult{
				Action:  "部署 SECURITY.md",
				Success: false,
				Details: fmt.Sprintf("%s — 写入失败: %v", filepath.Base(ws), err),
			})
		} else {
			results = append(results, HardenResult{
				Action:  "部署 SECURITY.md",
				Success: true,
				Details: fmt.Sprintf("%s — 已部署", filepath.Base(ws)),
			})
		}
	}
	return results
}

// injectAgentsMD adds security reference line to AGENTS.md files.
func injectAgentsMD(home string) []HardenResult {
	workspaces := findWorkspaces(home)
	var results []HardenResult

	for _, ws := range workspaces {
		agentsPath := filepath.Join(ws, "AGENTS.md")
		data, err := os.ReadFile(agentsPath)
		if err != nil {
			continue // No AGENTS.md in this workspace, skip
		}

		content := string(data)
		if strings.Contains(content, "SECURITY.md") {
			results = append(results, HardenResult{
				Action:  "注入 AGENTS.md",
				Success: true,
				Details: fmt.Sprintf("%s — 已包含安全引用，跳过", filepath.Base(ws)),
			})
			continue
		}
		// Inject the security line after first heading
		lines := strings.SplitN(content, "\n", 2)
		var newContent string
		if len(lines) > 1 && strings.HasPrefix(lines[0], "#") {
			newContent = lines[0] + "\n\n" + AgentsMDSecurityLine + "\n" + lines[1]
		} else {
			newContent = AgentsMDSecurityLine + "\n\n" + content
		}

		err = os.WriteFile(agentsPath, []byte(newContent), 0644)
		if err != nil {
			results = append(results, HardenResult{
				Action:  "注入 AGENTS.md",
				Success: false,
				Details: fmt.Sprintf("%s — 写入失败: %v", filepath.Base(ws), err),
			})
		} else {
			results = append(results, HardenResult{
				Action:  "注入 AGENTS.md",
				Success: true,
				Details: fmt.Sprintf("%s — 已注入安全引用", filepath.Base(ws)),
			})
		}
	}
	return results
}

// patchSoulMD appends security boundary rules to SOUL.md files.
func patchSoulMD(home string) []HardenResult {
	workspaces := findWorkspaces(home)
	var results []HardenResult

	for _, ws := range workspaces {
		soulPath := filepath.Join(ws, "SOUL.md")
		data, err := os.ReadFile(soulPath)
		if err != nil {
			continue
		}

		content := string(data)
		if strings.Contains(content, "🔒") {
			results = append(results, HardenResult{
				Action:  "补丁 SOUL.md",
				Success: true,
				Details: fmt.Sprintf("%s — 已包含安全边界，跳过", filepath.Base(ws)),
			})
			continue
		}

		newContent := content + "\n" + SoulMDBoundaryPatch
		err = os.WriteFile(soulPath, []byte(newContent), 0644)
		if err != nil {
			results = append(results, HardenResult{
				Action:  "补丁 SOUL.md",
				Success: false,
				Details: fmt.Sprintf("%s — 写入失败: %v", filepath.Base(ws), err),
			})
		} else {
			results = append(results, HardenResult{
				Action:  "补丁 SOUL.md",
				Success: true,
				Details: fmt.Sprintf("%s — 已追加安全边界", filepath.Base(ws)),
			})
		}
	}
	return results
}

// fixPermissions sets sensitive config files to 0600.
func fixPermissions(home string) []HardenResult {
	sensitiveFiles := []string{
		filepath.Join(home, ".openclaw", "openclaw.json"),
		filepath.Join(home, ".opskit", "config.json"),
	}

	var results []HardenResult
	for _, f := range sensitiveFiles {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		perm := info.Mode().Perm()
		if perm == 0600 {
			results = append(results, HardenResult{
				Action:  "修复文件权限",
				Success: true,
				Details: fmt.Sprintf("%s — 权限正确 (0600)", filepath.Base(f)),
			})
			continue
		}
		err = os.Chmod(f, 0600)
		if err != nil {
			results = append(results, HardenResult{
				Action:  "修复文件权限",
				Success: false,
				Details: fmt.Sprintf("%s — 修复失败: %v", filepath.Base(f), err),
			})
		} else {
			results = append(results, HardenResult{
				Action:  "修复文件权限",
				Success: true,
				Details: fmt.Sprintf("%s — 已修复 %04o → 0600", filepath.Base(f), perm),
			})
		}
	}
	return results
}
