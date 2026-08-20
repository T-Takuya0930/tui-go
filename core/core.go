package core

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ===== Tool 共通インターフェース =====

type Tool interface {
	Init() tea.Cmd
	Update(tea.Msg) (Tool, tea.Cmd)
	View() tea.View
}

type ToolDoneMsg struct {
	Result string
	Err    error
}

func ToolDone(result string, err error) tea.Cmd {
	return func() tea.Msg { return ToolDoneMsg{Result: result, Err: err} }
}

// ===== ステージ定義 =====

type StageKind int

const (
	StageInput StageKind = iota
	StageChoice
)

const (
	SelectSingle   = "single"
	SelectMultiple = "multiple"
)

type Stage struct {
	StepName   string
	Kind       StageKind
	TextInput  textinput.Model
	Options    []string
	SelectType string

	Cursor   int
	Selected map[int]struct{}
	Result   string
}

func NewTextInputStage(stepName, placeholder string, width int) Stage {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.SetWidth(width)
	return Stage{StepName: stepName, Kind: StageInput, TextInput: ti}
}

func NewChoiceStage(stepName string, options []string, selectType string) Stage {
	return Stage{
		StepName:   stepName,
		Kind:       StageChoice,
		Options:    options,
		SelectType: selectType,
		Selected:   make(map[int]struct{}),
	}
}

// ===== スタイル =====

var (
	ColorPrimary = lipgloss.Color("205")
	ColorSubtle  = lipgloss.Color("240")
	ColorSuccess = lipgloss.Color("42")
	ColorError   = lipgloss.Color("196")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(1, 2)

	CursorStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	CheckedStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	ProgressStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			Italic(true).
			MarginBottom(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			MarginTop(1)

	ResultOKStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	ResultErrStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
)

func RenderCursor(active bool, mark string) string {
	if active {
		return CursorStyle.Render("❯ " + mark)
	}
	return NormalItemStyle.Render("  " + mark)
}

func RenderHelp(parts ...string) string {
	return HelpStyle.Render(strings.Join(parts, "  ·  "))
}
