package health

import (
	"fmt"
	"opskit/internal/exec"
	"strconv"
	"strings"
	"time"
)

// ProbeResult holds the outcome of a single health probe.
type ProbeResult struct {
	Name     string // probe name in Chinese
	OK       bool
	Score    int // points earned (0 if not OK)
	MaxScore int // max possible points
	Detail   string
}

// HealthReport aggregates all probe results.
type HealthReport struct {
	Probes  []ProbeResult
	TotalHP int // sum of all scores (0-100)
	MaxHP   int // always 100
}

// RunProbes executes all health probes and returns a report.
func RunProbes() HealthReport {
	runner := exec.NewRunner(5 * time.Second)

	probes := []func(*exec.Runner) ProbeResult{
		probeGatewayProcess,
		probeRPCResponse,
		probeAPIConfig,
		probeDiskSpace,
		probeErrorLogs,
	}

	report := HealthReport{MaxHP: 100}
	for _, p := range probes {
		result := p(runner)
		report.Probes = append(report.Probes, result)
		report.TotalHP += result.Score
	}
	return report
}

func probeGatewayProcess(r *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "网关进程", MaxScore: 30}
	out, _ := r.Run(`pgrep -f "openclaw" 2>/dev/null || launchctl list 2>/dev/null | grep openclaw`)
	out = strings.TrimSpace(out)
	if out != "" {
		pid := strings.Split(out, "\n")[0]
		result.OK = true
		result.Score = 30
		result.Detail = fmt.Sprintf("进程存活 (PID %s)", strings.TrimSpace(pid))
	} else {
		result.Detail = "进程未找到"
	}
	return result
}

func probeRPCResponse(r *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "RPC 探测", MaxScore: 30}
	out, _ := r.Run(`curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 http://127.0.0.1:18789/ 2>/dev/null`)
	out = strings.TrimSpace(out)
	if strings.HasPrefix(out, "2") {
		result.OK = true
		result.Score = 30
		result.Detail = fmt.Sprintf("HTTP %s 响应正常", out)
	} else {
		result.Detail = "无法连接 (端口 18789)"
	}
	return result
}

func probeAPIConfig(r *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "API 配置", MaxScore: 20}
	out, err := r.Run(`cat ~/.openclaw/openclaw.json 2>/dev/null`)
	if err == nil && strings.Contains(out, "apiKey") {
		result.OK = true
		result.Score = 20
		result.Detail = "API Key 已配置"
	} else {
		result.Detail = "配置文件不存在或 API Key 未配置"
	}
	return result
}

func probeDiskSpace(r *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "磁盘空间", MaxScore: 10}
	out, err := r.Run(`df -h ~ | tail -1 | awk '{print $4}'`)
	out = strings.TrimSpace(out)
	if err != nil || out == "" {
		result.Detail = "无法检测磁盘空间"
		return result
	}
	result.Detail = fmt.Sprintf("可用空间: %s", out)
	// Parse available space — value like "50Gi", "500Mi", "2.5Ti", "800Ki"
	upper := strings.ToUpper(out)
	numStr := strings.TrimRight(upper, "BKKMGITP ")
	// Remove trailing unit letter(s) to get the number
	for len(numStr) > 0 && (numStr[len(numStr)-1] < '0' || numStr[len(numStr)-1] > '9') && numStr[len(numStr)-1] != '.' {
		numStr = numStr[:len(numStr)-1]
	}
	val, parseErr := strconv.ParseFloat(numStr, 64)
	if parseErr != nil {
		result.Detail = fmt.Sprintf("可用空间: %s", out)
		return result
	}
	// Convert to GB
	gb := val
	switch {
	case strings.Contains(upper, "T"):
		gb = val * 1024
	case strings.Contains(upper, "G"):
		gb = val
	case strings.Contains(upper, "M"):
		gb = val / 1024
	case strings.Contains(upper, "K"):
		gb = val / (1024 * 1024)
	}
	if gb > 1.0 {
		result.OK = true
		result.Score = 10
	} else {
		result.Detail = fmt.Sprintf("磁盘空间不足 (%s)", out)
	}
	return result
}

func probeErrorLogs(r *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "错误日志", MaxScore: 10}
	out, _ := r.Run(`grep -c "ERROR" ~/.openclaw/logs/*.log 2>/dev/null || echo "0"`)
	out = strings.TrimSpace(out)
	// grep -c may return multiple lines (one per file) like "file:3"
	// Sum all counts; if output is just "0", total is 0
	total := 0
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		// Handle "filename:count" format
		if idx := strings.LastIndex(line, ":"); idx >= 0 {
			line = line[idx+1:]
		}
		if n, err := strconv.Atoi(line); err == nil {
			total += n
		}
	}
	if total == 0 {
		result.OK = true
		result.Score = 10
		result.Detail = "最近无错误日志"
	} else {
		result.Detail = fmt.Sprintf("发现 %d 条错误日志", total)
	}
	return result
}
