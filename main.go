package main

import (
	"fmt"
	"os"
	"strings"

	"tui-go/commands"
	"tui-go/core"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	cursor int
	active core.Tool
	result string
	isErr  bool
}

func initialModel() model { return model{} }

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.active != nil {
		if done, ok := msg.(core.ToolDoneMsg); ok {
			m.active = nil
			if done.Err != nil {
				m.result = done.Err.Error()
				m.isErr = true
			} else if done.Result != "" {
				m.result = done.Result
				m.isErr = false
			}
			return m, nil
		}
		newTool, cmd := m.active.Update(msg)
		m.active = newTool
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(commands.Registry)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.result = ""
			m.active = commands.Registry[m.cursor].New()
			return m, m.active.Init()
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	if m.active != nil {
		v := m.active.View()
		v.AltScreen = true
		return v
	}

	var body strings.Builder
	body.WriteString(core.TitleStyle.Render("🛠  ツールランチャー"))
	body.WriteString("\n\n")

	for i, t := range commands.Registry {
		body.WriteString(core.RenderCursor(i == m.cursor, t.Label))
		body.WriteString("\n")
	}

	if m.result != "" {
		body.WriteString("\n")
		if m.isErr {
			body.WriteString(core.ResultErrStyle.Render("✗ ") + m.result)
		} else {
			body.WriteString(core.ResultOKStyle.Render("✓ ") + m.result)
		}
		body.WriteString("\n")
	}

	body.WriteString(core.RenderHelp("↑/↓ 移動", "Enter 選択", "q 終了"))

	return tea.NewView(core.BoxStyle.Render(body.String()))
}

func main() {

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
