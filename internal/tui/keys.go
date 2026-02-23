package tui

import tea "github.com/charmbracelet/bubbletea"

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionReview
	ActionUnreview
	ActionIgnore
	ActionExpand
	ActionBack
	ActionCopy
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
	case "h", "left":
		return ActionLeft
	case "l", "right":
		return ActionRight
	case "r":
		return ActionReview
	case "u":
		return ActionUnreview
	case "i":
		return ActionIgnore
	case "enter":
		return ActionExpand
	case "esc":
		return ActionBack
	case "c":
		return ActionCopy
	case "tab":
		return ActionRight
	case "shift+tab":
		return ActionLeft
	case "n":
		return ActionNote
	default:
		return ActionNone
	}
}
