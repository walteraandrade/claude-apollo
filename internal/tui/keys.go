package tui

import tea "github.com/charmbracelet/bubbletea"

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionUp
	ActionDown
	ActionReview
	ActionUnreview
	ActionIgnore
	ActionDetail
	ActionBack
	ActionCycleFilter
	ActionNote
)

func MapKey(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "q", "ctrl+c":
		return ActionQuit
	case "k", "up":
		return ActionUp
	case "j", "down":
		return ActionDown
	case "r":
		return ActionReview
	case "u":
		return ActionUnreview
	case "i":
		return ActionIgnore
	case "enter":
		return ActionDetail
	case "esc":
		return ActionBack
	case "tab":
		return ActionCycleFilter
	case "n":
		return ActionNote
	default:
		return ActionNone
	}
}
