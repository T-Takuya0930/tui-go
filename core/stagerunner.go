package core

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// StageRunner は複数の Stage を順番に進めていく共通ロジックを提供する。
// 各 Tool はこれを保持し、Update/View から委譲するだけでよい。
type StageRunner struct {
	Title  string
	Idx    int
	Stages []Stage
}

func NewStageRunner(title string, stages []Stage) StageRunner {
	return StageRunner{Title: title, Stages: stages}
}

func (r *StageRunner) Init() tea.Cmd { return textinput.Blink }

// Done は全ステージが完了したかどうかを返す。
func (r *StageRunner) Done() bool { return r.Idx >= len(r.Stages) }

// Advance は次のステージへ進む。入力ステージならフォーカスする。
func (r *StageRunner) Advance() tea.Cmd {
	r.Idx++
	if r.Done() {
		return nil
	}
	if r.Stages[r.Idx].Kind == StageInput {
		r.Stages[r.Idx].TextInput.Focus()
		return textinput.Blink
	}
	return nil
}

// HandleKey は現在のステージに対する共通のキー操作
// (上下移動・選択トグル・確定・入力編集・Escでの中断) を処理する。
//
// advanced: Enterで確定し次のステージへ進んだ場合 true
// canceled: Escが押された場合 true(呼び出し側で ToolDone("", nil) などを返すこと)
func (r *StageRunner) HandleKey(km tea.KeyMsg) (advanced bool, canceled bool, cmd tea.Cmd) {
	if r.Done() {
		return false, false, nil
	}
	if km.String() == "esc" {
		return false, true, nil
	}

	s := &r.Stages[r.Idx]
	switch s.Kind {
	case StageInput:
		if km.String() == "enter" {
			s.Result = s.TextInput.Value()
			return true, false, r.Advance()
		}
		var c tea.Cmd
		s.TextInput, c = s.TextInput.Update(km)
		return false, false, c

	case StageChoice:
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
			if s.SelectType == SelectMultiple {
				if _, ok := s.Selected[s.Cursor]; ok {
					delete(s.Selected, s.Cursor)
				} else {
					s.Selected[s.Cursor] = struct{}{}
				}
			}
		case "enter":
			switch s.SelectType {
			case SelectSingle:
				s.Result = s.Options[s.Cursor]
			case SelectMultiple:
				chosen := make([]string, 0, len(s.Selected))
				for i, opt := range s.Options {
					if _, ok := s.Selected[i]; ok {
						chosen = append(chosen, opt)
					}
				}
				s.Result = strings.Join(chosen, ",")
			}
			return true, false, r.Advance()
		}
	}
	return false, false, nil
}

// View は現在のステージを共通レイアウト(タイトル・進捗・本体・ヘルプ)で描画する。
func (r *StageRunner) View() tea.View {
	if r.Done() {
		return tea.NewView("")
	}
	s := &r.Stages[r.Idx]

	var body strings.Builder
	body.WriteString(TitleStyle.Render(r.Title))
	body.WriteString("\n")
	body.WriteString(ProgressStyle.Render(fmt.Sprintf("ステップ %d/%d ・ %s", r.Idx+1, len(r.Stages), s.StepName)))
	body.WriteString("\n\n")

	switch s.Kind {
	case StageInput:
		body.WriteString(s.TextInput.View())
		body.WriteString("\n")
		body.WriteString(RenderHelp("Enter 次へ", "Esc 中止"))

	case StageChoice:
		for i, opt := range s.Options {
			active := i == s.Cursor
			mark := ""
			if s.SelectType == SelectMultiple {
				if _, ok := s.Selected[i]; ok {
					mark = CheckedStyle.Render("[x] ")
				} else {
					mark = "[ ] "
				}
			}
			body.WriteString(RenderCursor(active, mark+opt))
			body.WriteString("\n")
		}
		if s.SelectType == SelectMultiple {
			body.WriteString(RenderHelp("↑/↓ 移動", "Space トグル", "Enter 確定", "Esc 中止"))
		} else {
			body.WriteString(RenderHelp("↑/↓ 移動", "Enter 確定", "Esc 中止"))
		}
	}

	return tea.NewView(BoxStyle.Render(body.String()))
}
