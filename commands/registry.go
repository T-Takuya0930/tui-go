package commands

import (
	"tui-go/core"
)

type Entry struct {
	Label string
	New   func() core.Tool
}

// Registry: 新しいツールはここに1行足すだけで登録できる
var Registry = []Entry{
	{"1: Ping command", NewToolPing},
	{"2: Git Init", NewToolGitInit},
}
