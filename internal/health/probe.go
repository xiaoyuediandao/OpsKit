package health

import (
	"fmt"
	"net/http"
	"opskit/internal/exec"
	"os"
	"path/filepath"
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

func probeRPCResponse(_ *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "RPC 探测", MaxScore: 30}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:18789/")
	if err != nil {
		result.Detail = "无法连接 (端口 18789)"
		return result
	}
	defer resp.Body.Close()
	code := resp.StatusCode
	if code >= 200 && code < 300 {
		result.OK = true
		result.Score = 30
		result.Detail = fmt.Sprintf("HTTP %d 响应正常", code)
	} else {
		result.Detail = fmt.Sprintf("HTTP %d 响应异常", code)
	}
	return result
}

func probeAPIConfig(_ *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "API 配置", MaxScore: 20}
	home, err := os.UserHomeDir()
	if err != nil {
		result.Detail = "无法获取用户目录"
		return result
	}
	data, err := os.ReadFile(filepath.Join(home, ".openclaw", "openclaw.json"))
	if err != nil {
		result.Detail = "配置文件不存在或 API Key 未配置"
		return result
	}
	if strings.Contains(string(data), "apiKey") {
		result.OK = true
		result.Score = 20
		result.Detail = "API Key 已配置"
	} else {
		result.Detail = "配置文件不存在或 API Key 未配置"
	}
	return result
}
