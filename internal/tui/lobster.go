package tui

import (
	"fmt"
	"opskit/internal/health"
	"opskit/internal/state"
)

// Lobster represents the AI companion state.
type Lobster struct {
	Stage  int
	Level  int
	HP     int
	MaxHP  int
	XP     int
	Status string // "active", "sick", "idle"
}

// NewLobster creates a Lobster from persisted state.
func NewLobster(s *state.State) Lobster {
	return Lobster{
		Stage:  s.LobsterStage,
		Level:  s.LobsterLevel,
		HP:     s.LobsterHP,
		MaxHP:  s.LobsterMaxHP,
		XP:     s.LobsterXP,
		Status: s.LobsterStatus,
	}
}

// ToState converts the lobster to a state struct for persistence.
func (l *Lobster) ToState() *state.State {
	return &state.State{
		LobsterStage:  l.Stage,
		LobsterLevel:  l.Level,
		LobsterXP:     l.XP,
		LobsterHP:     l.HP,
		LobsterMaxHP:  l.MaxHP,
		LobsterStatus: l.Status,
	}
}

// XPToNextLevel returns the XP needed to reach the next level.
func (l *Lobster) XPToNextLevel() int {
	return l.Level * 100
}

// AddXP adds XP and handles leveling up.
func (l *Lobster) AddXP(xp int) {
	l.XP += xp
	for l.XP >= l.XPToNextLevel() {
		l.XP -= l.XPToNextLevel()
		l.Level++
		l.MaxHP += 10
		l.HP = l.MaxHP
	}
}

// StageEmoji returns the emoji for the current stage.
func (l *Lobster) StageEmoji() string {
	switch l.Stage {
	case 0:
		return "🥚"
	case 1:
		return "🦐"
	case 2:
		return "🦞"
	case 3:
		return "🦞⚡"
	case 4:
		return "🦞⚔️"
	case 5:
		return "🦞👑"
	default:
		return "🦞"
	}
}

// StageName returns the name for the current stage.
func (l *Lobster) StageName() string {
	switch l.Stage {
	case 0:
		return "虾卵"
	case 1:
		return "小虾苗"
	case 2:
		return "幼年龙虾"
	case 3:
		return "成年龙虾"
	case 4:
		return "战士龙虾"
	case 5:
		return "龙虾之王"
	default:
		return "龙虾"
	}
}

// HPBar returns an ASCII HP bar.
func (l *Lobster) HPBar() string {
	if l.MaxHP == 0 {
		return "[----------]"
	}
	filled := (l.HP * 10) / l.MaxHP
	bar := "["
	for i := 0; i < 10; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += fmt.Sprintf("] %d/%d", l.HP, l.MaxHP)
	return bar
}

// XPBar returns an ASCII XP bar.
func (l *Lobster) XPBar() string {
	next := l.XPToNextLevel()
	if next == 0 {
		return "[----------]"
	}
	filled := (l.XP * 10) / next
	bar := "["
	for i := 0; i < 10; i++ {
		if i < filled {
			bar += "▪"
		} else {
			bar += "·"
		}
	}
	bar += fmt.Sprintf("] %d/%d", l.XP, next)
	return bar
}

// StatusIcon returns an icon for the current status.
func (l *Lobster) StatusIcon() string {
	return health.StatusEmoji(l.Status)
}

// StatusLabel returns the Chinese label for the current status.
func (l *Lobster) StatusLabel() string {
	return health.StatusLabel(l.Status)
}

// UpdateStatusFromHP updates the lobster's status based on current HP.
func (l *Lobster) UpdateStatusFromHP() {
	l.Status = health.HPStatus(l.HP, l.MaxHP)
}
