package security

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ScanResult holds the result of a single security check.
type ScanResult struct {
	Name    string
	Passed  bool
	Score   int
	Details string
}

// ScanReport holds the full security scan report.
type ScanReport struct {
	Results    []ScanResult
	TotalScore int
	MaxScore   int
	Grade      string
}

// Scan runs all security checks and returns a report.
func Scan() ScanReport {
	home, _ := os.UserHomeDir()
	results := []ScanResult{
		checkSecurityMD(home),
		checkAgentsMD(home),
		checkSoulMD(home),
		checkFilePermissions(home),
		checkKeyLeaks(home),
	}

	total := 0
	for _, r := range results {
		total += r.Score
	}

	grade := "F"
	switch {
	case total >= 90:
		grade = "A"
	case total >= 80:
		grade = "B"
	case total >= 60:
		grade = "C"
	case total >= 40:
		grade = "D"
	}
	return ScanReport{
		Results:    results,
		TotalScore: total,
		MaxScore:   100,
		Grade:      grade,
	}
}

func findWorkspaces(home string) []string {
	pattern := filepath.Join(home, ".openclaw", "workspace*")
	matches, _ := filepath.Glob(pattern)
	// Filter to directories only
	var dirs []string
	for _, m := range matches {
		info, err := os.Stat(m)
		if err == nil && info.IsDir() {
			dirs = append(dirs, m)
		}
	}
	return dirs
}

// Check 1: SECURITY.md deployed to all workspaces (20 points)
func checkSecurityMD(home string) ScanResult {
	workspaces := findWorkspaces(home)
	if len(workspaces) == 0 {
		return ScanResult{
			Name:    "SECURITY.md 部署",
			Passed:  false,
			Score:   0,
			Details: "未找到任何 workspace 目录",
		}
	}

	missing := []string{}
	for _, ws := range workspaces {
		secPath := filepath.Join(ws, "SECURITY.md")
		if _, err := os.Stat(secPath); os.IsNotExist(err) {
			missing = append(missing, filepath.Base(ws))
		}
	}

	if len(missing) == 0 {
		return ScanResult{
			Name:    "SECURITY.md 部署",
			Passed:  true,
			Score:   20,
			Details: fmt.Sprintf("全部 %d 个 workspace 已部署", len(workspaces)),
		}
	}
	return ScanResult{
		Name:    "SECURITY.md 部署",
		Passed:  false,
		Score:   0,
		Details: fmt.Sprintf("缺失: %s", strings.Join(missing, ", ")),
	}
}
// Check 2: AGENTS.md references SECURITY.md (20 points)
func checkAgentsMD(home string) ScanResult {
	workspaces := findWorkspaces(home)
	if len(workspaces) == 0 {
		return ScanResult{Name: "AGENTS.md 安全引用", Passed: false, Score: 0, Details: "未找到 workspace"}
	}

	checked := 0
	passed := 0
	for _, ws := range workspaces {
		agentsPath := filepath.Join(ws, "AGENTS.md")
		data, err := os.ReadFile(agentsPath)
		if err != nil {
			continue
		}
		checked++
		if strings.Contains(string(data), "SECURITY.md") {
			passed++
		}
	}

	if checked == 0 {
		return ScanResult{Name: "AGENTS.md 安全引用", Passed: false, Score: 0, Details: "未找到 AGENTS.md 文件"}
	}
	if passed == checked {
		return ScanResult{Name: "AGENTS.md 安全引用", Passed: true, Score: 20, Details: fmt.Sprintf("%d/%d 已配置", passed, checked)}
	}
	return ScanResult{Name: "AGENTS.md 安全引用", Passed: false, Score: 0, Details: fmt.Sprintf("%d/%d 已配置", passed, checked)}
}

// Check 3: SOUL.md contains security boundary (20 points)
func checkSoulMD(home string) ScanResult {
	workspaces := findWorkspaces(home)
	if len(workspaces) == 0 {
		return ScanResult{Name: "SOUL.md 安全边界", Passed: false, Score: 0, Details: "未找到 workspace"}
	}

	checked := 0
	passed := 0
	for _, ws := range workspaces {
		soulPath := filepath.Join(ws, "SOUL.md")
		data, err := os.ReadFile(soulPath)
		if err != nil {
			continue
		}
		checked++
		if strings.Contains(string(data), "🔒") {
			passed++
		}
	}

	if checked == 0 {
		return ScanResult{Name: "SOUL.md 安全边界", Passed: false, Score: 0, Details: "未找到 SOUL.md 文件"}
	}
	if passed == checked {
		return ScanResult{Name: "SOUL.md 安全边界", Passed: true, Score: 20, Details: fmt.Sprintf("%d/%d 已配置", passed, checked)}
	}
	return ScanResult{Name: "SOUL.md 安全边界", Passed: false, Score: 0, Details: fmt.Sprintf("%d/%d 已配置", passed, checked)}
}
// Check 4: Sensitive file permissions are 0600 (20 points, skip on Windows)
func checkFilePermissions(home string) ScanResult {
	if runtime.GOOS == "windows" {
		return ScanResult{Name: "文件权限检查", Passed: true, Score: 20, Details: "Windows 平台跳过"}
	}

	sensitiveFiles := []string{
		filepath.Join(home, ".openclaw", "openclaw.json"),
		filepath.Join(home, ".opskit", "config.json"),
	}

	bad := []string{}
	checked := 0
	for _, f := range sensitiveFiles {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		checked++
		perm := info.Mode().Perm()
		if perm != 0600 {
			bad = append(bad, fmt.Sprintf("%s (%04o)", filepath.Base(f), perm))
		}
	}

	if checked == 0 {
		return ScanResult{Name: "文件权限检查", Passed: true, Score: 20, Details: "无敏感文件需要检查"}
	}
	if len(bad) == 0 {
		return ScanResult{Name: "文件权限检查", Passed: true, Score: 20, Details: fmt.Sprintf("%d 个文件权限正确 (0600)", checked)}
	}
	return ScanResult{Name: "文件权限检查", Passed: false, Score: 0, Details: fmt.Sprintf("权限不正确: %s", strings.Join(bad, ", "))}
}

// Check 5: No API key leaks in non-config workspace files (20 points)
func checkKeyLeaks(home string) ScanResult {
	workspaces := findWorkspaces(home)
	if len(workspaces) == 0 {
		return ScanResult{Name: "密钥泄露扫描", Passed: true, Score: 20, Details: "无 workspace 需要扫描"}
	}

	leakPatterns := []string{
		`apiKey`,
		`appSecret`,
		`app_secret`,
	}

	leaks := []string{}
	for _, ws := range workspaces {
		entries, err := os.ReadDir(ws)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Skip config files and SECURITY.md itself
			if name == "openclaw.json" || name == "SECURITY.md" || strings.HasSuffix(name, ".json") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(ws, name))
			if err != nil {
				continue
			}
			content := string(data)
			for _, pat := range leakPatterns {
				if strings.Contains(content, pat) {
					leaks = append(leaks, fmt.Sprintf("%s/%s", filepath.Base(ws), name))
					break
				}
			}
		}
	}

	if len(leaks) == 0 {
		return ScanResult{Name: "密钥泄露扫描", Passed: true, Score: 20, Details: "未发现泄露"}
	}
	return ScanResult{Name: "密钥泄露扫描", Passed: false, Score: 0, Details: fmt.Sprintf("疑似泄露: %s", strings.Join(leaks, ", "))}
}
