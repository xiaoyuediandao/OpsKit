//go:build !windows

package health

import (
	"fmt"
	"opskit/internal/exec"
	"strconv"
	"strings"
)

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

func probeDiskSpace(r *exec.Runner) ProbeResult {
	result := ProbeResult{Name: "磁盘空间", MaxScore: 10}
	out, err := r.Run(`df -h ~ | tail -1 | awk '{print $4}'`)
	out = strings.TrimSpace(out)
	if err != nil || out == "" {
		result.Detail = "无法检测磁盘空间"
		return result
	}
	result.Detail = fmt.Sprintf("可用空间: %s", out)
	upper := strings.ToUpper(out)
	numStr := strings.TrimRight(upper, "BKKMGITP ")
	for len(numStr) > 0 && (numStr[len(numStr)-1] < '0' || numStr[len(numStr)-1] > '9') && numStr[len(numStr)-1] != '.' {
		numStr = numStr[:len(numStr)-1]
	}
	val, parseErr := strconv.ParseFloat(numStr, 64)
	if parseErr != nil {
		return result
	}
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
	total := 0
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
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
