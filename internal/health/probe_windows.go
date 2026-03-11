//go:build windows

package health

import (
	"bufio"
	"fmt"
	"opskit/internal/exec"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func probeGatewayProcess(r *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "网关进程", MaxScore: 30}
	out, _ := r.Run(`tasklist /FI "IMAGENAME eq node.exe" /FO CSV /NH`)
	out = strings.TrimSpace(out)
	if strings.Contains(strings.ToLower(out), "node.exe") {
		// Extract PID from CSV: "node.exe","1234",...
		parts := strings.Split(out, ",")
		pid := "unknown"
		if len(parts) >= 2 {
			pid = strings.Trim(parts[1], "\" ")
		}
		result.OK = true
		result.Score = 30
		result.Detail = fmt.Sprintf("进程存活 (PID %s)", pid)
	} else {
		result.Detail = "进程未找到"
	}
	return result
}

func probeDiskSpace(r *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "磁盘空间", MaxScore: 10}
	out, err := r.Run(`(Get-PSDrive C).Free`)
	out = strings.TrimSpace(out)
	if err != nil || out == "" {
		result.Detail = "无法检测磁盘空间"
		return result
	}
	freeBytes, parseErr := strconv.ParseFloat(out, 64)
	if parseErr != nil {
		result.Detail = fmt.Sprintf("可用空间: %s (无法解析)", out)
		return result
	}
	gb := freeBytes / (1024 * 1024 * 1024)
	result.Detail = fmt.Sprintf("可用空间: %.1f GB", gb)
	if gb > 1.0 {
		result.OK = true
		result.Score = 10
	} else {
		result.Detail = fmt.Sprintf("磁盘空间不足 (%.1f GB)", gb)
	}
	return result
}

func probeErrorLogs(_ *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "错误日志", MaxScore: 10}
	home, err := os.UserHomeDir()
	if err != nil {
		result.Detail = "无法获取用户目录"
		return result
	}
	logDir := filepath.Join(home, ".openclaw", "logs")
	matches, err := filepath.Glob(filepath.Join(logDir, "*.log"))
	if err != nil || len(matches) == 0 {
		result.OK = true
		result.Score = 10
		result.Detail = "最近无错误日志"
		return result
	}
	total := 0
	for _, path := range matches {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "ERROR") {
				total++
			}
		}
		f.Close()
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
