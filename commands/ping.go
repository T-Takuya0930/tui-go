package commands

import (
	"fmt"
	"strings"

	"tui-go/core"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type toolPing struct {
	title  string
	idx    int
	stages []core.Stage
}

func NewToolPing() core.Tool {
	return &toolPing{
		title: "Ping",
		stages: []core.Stage{
			core.NewTextInputStage("名前", "名前を入力してください", 20),
			core.NewChoiceStage("モード選択", []string{"Option 1", "Option 2"}, core.SelectSingle),
			core.NewChoiceStage("機能選択", []string{"Option A", "Option B", "Option C"}, core.SelectMultiple),
		},
	}
}

func (t *toolPing) Init() tea.Cmd { return textinput.Blink }

func (t *toolPing) advance() tea.Cmd {
	t.idx++
	if t.idx >= len(t.stages) {
		return core.ToolDone(t.summary(), nil)
	}
	if t.stages[t.idx].Kind == core.StageInput {
		t.stages[t.idx].TextInput.Focus()
		return textinput.Blink
	}
	return nil
}

func (t *toolPing) summary() string {
	parts := make([]string, 0, len(t.stages))
	for _, s := range t.stages {
		if s.Result != "" {
			parts = append(parts, s.Result)
		}
	}
	return strings.Join(parts, " / ")
}

func (t *toolPing) Update(msg tea.Msg) (core.Tool, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return t, nil
	}
	if t.idx >= len(t.stages) {
		return t, nil
	}
	s := &t.stages[t.idx]

	if km.String() == "esc" {
		return t, core.ToolDone("", nil)
	}

	switch s.Kind {
	case core.StageInput:
		if km.String() == "enter" {
			s.Result = s.TextInput.Value()
			return t, t.advance()
		}
		var cmd tea.Cmd
		s.TextInput, cmd = s.TextInput.Update(msg)
		return t, cmd

	case core.StageChoice:
		switch km.String() {
		case "up", "k":
			if s.Cursor > 0 {
				s.Cursor--
			}
		case "down", "j":
			if s.Cursor < len(s.Options)-1 {
				s.Cursor++
			}
		case " ", "space":
			if s.SelectType == core.SelectMultiple {
				if _, ok := s.Selected[s.Cursor]; ok {
					delete(s.Selected, s.Cursor)
				} else {
					s.Selected[s.Cursor] = struct{}{}
				}
			}
		case "enter":
			switch s.SelectType {
			case core.SelectSingle:
				s.Result = fmt.Sprintf("%s: %s", s.StepName, s.Options[s.Cursor])
			case core.SelectMultiple:
				chosen := make([]string, 0, len(s.Selected))
				for i, opt := range s.Options {
					if _, ok := s.Selected[i]; ok {
						chosen = append(chosen, opt)
					}
				}
				s.Result = fmt.Sprintf("%s: %s", s.StepName, strings.Join(chosen, ","))
			}
			return t, t.advance()
		}
	}
	return t, nil
}

func (t *toolPing) View() tea.View {
	if t.idx >= len(t.stages) {
		return tea.NewView("")
	}
	s := &t.stages[t.idx]

	var body strings.Builder
	body.WriteString(core.TitleStyle.Render(t.title))
	body.WriteString("\n")
	body.WriteString(core.ProgressStyle.Render(fmt.Sprintf("ステップ %d/%d ・ %s", t.idx+1, len(t.stages), s.StepName)))
	body.WriteString("\n\n")

	switch s.Kind {
	case core.StageInput:
		body.WriteString(s.TextInput.View())
		body.WriteString("\n")
		body.WriteString(core.RenderHelp("Enter 次へ", "Esc 中止"))

	case core.StageChoice:
		for i, opt := range s.Options {
			active := i == s.Cursor
			mark := ""
			if s.SelectType == core.SelectMultiple {
				if _, ok := s.Selected[i]; ok {
					mark = core.CheckedStyle.Render("[x] ")
				} else {
					mark = "[ ] "
				}
			}
			body.WriteString(core.RenderCursor(active, mark+opt))
			body.WriteString("\n")
		}
		if s.SelectType == core.SelectMultiple {
			body.WriteString(core.RenderHelp("↑/↓ 移動", "Space トグル", "Enter 確定", "Esc 中止"))
		} else {
			body.WriteString(core.RenderHelp("↑/↓ 移動", "Enter 確定", "Esc 中止"))
		}
	}

	return tea.NewView(core.BoxStyle.Render(body.String()))
}
