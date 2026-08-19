package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string         // items on the to-do list
	cursor   int              // which to-do list item our cursor is pointing at
	selected map[int]struct{} // which to-do items are selected
}

func initModel() model {
	return model{
		choices:  []string{"a: abc", "b: 123"},
		cursor:   0,
		selected: make(map[int]struct{}),
	}
}

func Init() tea.Cmd {
	return nil
}

func update(model model, msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl-c":
			return model, tea.Quit
		case "up", "k":
			if model.cursor > 0 {
				model.cursor--
			}
		case "down", "j":
			if model.cursor < len(model.choices)-1 {
				model.cursor++
			}
		case " ", "enter":
			return model, nil
		}
	}
	return model, nil
}

func View(model model) string {
	s := "This is test"
	for i, choice := range model.choices {
		cursor := " "
		if i == model.cursor {
			cursor = ">"
		}
		s += cursor + " " + choice + "\n"

		selected := " "
		if _, ok := model.selected[i]; ok {
			selected = "x"
		}
		s += selected + " " + choice + "\n"
	}
	return s
}

func main() {
	p := tea.NewProgram(initModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
