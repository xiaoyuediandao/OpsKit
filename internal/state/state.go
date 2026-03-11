package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// State represents persistent lobster/app state stored at ~/.opskit/state.json
type State struct {
	LobsterStage  int    `json:"lobster_stage"`
	LobsterLevel  int    `json:"lobster_level"`
	LobsterXP     int    `json:"lobster_xp"`
	LobsterHP     int    `json:"lobster_hp"`
	LobsterMaxHP  int    `json:"lobster_max_hp"`
	LobsterStatus string `json:"lobster_status"`

	Quests       map[string]string `json:"quests"`         // quest ID -> status ("locked","available","done")
	Achievements []string          `json:"achievements"`   // list of unlocked achievement IDs
	ToolUseCount int               `json:"tool_use_count"` // total successful tool executions
	LogViewCount int               `json:"log_view_count"` // times user viewed logs
}

func DefaultState() *State {
	return &State{
		LobsterStage:  0,
		LobsterLevel:  1,
		LobsterXP:     0,
		LobsterHP:     100,
		LobsterMaxHP:  100,
		LobsterStatus: "idle",
		Quests: map[string]string{
			"Q1": "available",
			"Q2": "locked",
			"Q3": "locked",
			"Q4": "locked",
			"Q5": "locked",
			"Q6": "locked",
			"Q7": "locked",
			"Q8": "locked",
			"Q9":  "locked",
			"Q10": "locked",
		},
		Achievements: []string{},
		ToolUseCount: 0,
		LogViewCount: 0,
	}
}

func statePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".opskit", "state.json"), nil
}

func Load() (*State, error) {
	path, err := statePath()
	if err != nil {
		return DefaultState(), nil
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultState(), nil
	}
	if err != nil {
		return DefaultState(), nil
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultState(), nil
	}
	return &s, nil
}

func Save(s *State) error {
	path, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
