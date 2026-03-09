package exec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WriteOpenClawConfig writes the full Volcengine provider config JSON to ~/.openclaw/openclaw.json.
// It backs up any existing config before writing.
func WriteOpenClawConfig(apiKey string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home dir: %w", err)
	}

	configDir := filepath.Join(home, ".openclaw")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "openclaw.json")

	// Backup existing config
	if _, err := os.Stat(configPath); err == nil {
		backupPath := fmt.Sprintf("%s.bak.%d", configPath, time.Now().Unix())
		data, readErr := os.ReadFile(configPath)
		if readErr == nil {
			_ = os.WriteFile(backupPath, data, 0600)
		}
	}

	cfg := buildConfig(apiKey)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0600)
}

func buildConfig(apiKey string) map[string]interface{} {
	return map[string]interface{}{
		"models": map[string]interface{}{
			"providers": map[string]interface{}{
				"volcengine-plan": map[string]interface{}{
					"baseUrl": "https://ark.cn-beijing.volces.com/api/coding/v3",
					"apiKey":  apiKey,
					"api":     "openai-completions",
					"models": []map[string]interface{}{
						{
							"id":            "ark-code-latest",
							"name":          "ark-code-latest",
							"api":           "openai-completions",
							"reasoning":     false,
							"input":         []string{"text", "image"},
							"cost":          map[string]interface{}{"input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0},
							"contextWindow": 200000,
							"maxTokens":     32000,
						},
					},
				},
			},
		},
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"model": map[string]interface{}{
					"primary": "volcengine-plan/ark-code-latest",
				},
				"models": map[string]interface{}{
					"volcengine-plan/ark-code-latest": map[string]interface{}{},
				},
			},
		},
		"gateway": map[string]interface{}{
			"mode": "local",
		},
	}
}
