package commands

import (
	"strings"

	"tui-go/core"

	tea "charm.land/bubbletea/v2"
)

type toolPing struct {
	runner core.StageRunner
}

func NewToolPing() core.Tool {
	return &toolPing{
		runner: core.NewStageRunner("Ping", []core.Stage{
			core.NewTextInputStage("Message: ", "Input message that you want to ping", 40),
			core.NewChoiceStage("Mode: ", []string{"English", "Japanese"}, core.SelectSingle),
			core.NewChoiceStage("Additional: ", []string{"Bold", "Italic", "Underline"}, core.SelectMultiple),
		}),
	}
}

func (t *toolPing) Init() tea.Cmd { return t.runner.Init() }

func (t *toolPing) Update(msg tea.Msg) (core.Tool, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return t, nil
	}

	advanced, canceled, cmd := t.runner.HandleKey(km)
	if canceled {
		return t, core.ToolDone("", nil)
	}
	if advanced && t.runner.Done() {
		text := t.runner.Stages[0].Result
		language := t.runner.Stages[1].Result
		var chosen []string
		if t.runner.Stages[2].Result != "" {
			chosen = strings.Split(t.runner.Stages[2].Result, ",")
		}
		return t, core.ToolDone(ping(text, language, chosen), nil)
	}
	return t, cmd
}

func (t *toolPing) View() tea.View { return t.runner.View() }

func ping(text string, language string, option []string) string {
	var result string
	language = strings.ReplaceAll(language, " ", "")
	switch language {
	case "Japanese":
		result = "pingメッセージです: " + text
	case "English":
		result = "ping message: " + text
	default:
		return "Error: unknown language"
	}
	return core.RenderStyled(result, option)
}
