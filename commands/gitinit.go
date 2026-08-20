package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"tui-go/core"
	"tui-go/core/git"

	tea "github.com/charmbracelet/bubbletea"
)

type Phase int

const (
	PhaseTargetDir Phase = iota
	PhaseUserSelect
	PhaseUserNameInput
	PhaseUserEmailInput
	PhaseRemoteSelect
	PhaseRemoteCustomInput
	PhaseConfirm
	PhaseExecute
)

type ToolGitInit struct {
	phase        Phase
	targetDir    string
	selectedUser git.UserProfile
	newUser      git.UserProfile
	userProfiles []git.UserProfile
	remoteURL    string
	remoteOpts   []string
	commands     []string
}

// NewToolGitInit は core.Tool インターフェースを返す構造体生成関数です
func NewToolGitInit() core.Tool {
	return &ToolGitInit{
		phase:     PhaseTargetDir,
		targetDir: ".",
	}
}

// Init はツールの初期化コマンドを返します（core.Tool インターフェースの実装）
func (t *ToolGitInit) Init() tea.Cmd {
	t.phase = PhaseTargetDir
	t.targetDir = "."
	t.selectedUser = git.UserProfile{}
	t.newUser = git.UserProfile{}
	t.userProfiles = nil
	t.remoteURL = ""
	t.remoteOpts = nil
	t.commands = nil
	return nil
}

// Run は対話型ステージ管理とコマンド実行を行うメイン処理です
func (t *ToolGitInit) Run() error {
	reader := bufio.NewReader(os.Stdin)

	for t.phase != PhaseExecute {
		switch t.phase {
		case PhaseTargetDir:
			fmt.Print("ターゲットディレクトリを入力 (デフォルト: .): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				t.targetDir = input
			}
			t.phase = PhaseUserSelect

		case PhaseUserSelect:
			profiles, _ := git.ParseGitConfig()
			t.userProfiles = profiles

			fmt.Println("\n--- ユーザー設定の選択 ---")
			for i, p := range profiles {
				fmt.Printf("[%d] %s <%s>\n", i+1, p.Name, p.Email)
			}
			fmt.Printf("[%d] 新規入力\n", len(profiles)+1)
			fmt.Print("選択してください: ")

			var choice int
			fmt.Scanln(&choice)

			if choice > 0 && choice <= len(profiles) {
				t.selectedUser = profiles[choice-1]
				t.phase = PhaseRemoteSelect
			} else if choice == len(profiles)+1 {
				t.phase = PhaseUserNameInput
			} else {
				fmt.Println("無効な選択肢です。")
			}

		case PhaseUserNameInput:
			fmt.Print("ユーザー名を入力 (user.name): ")
			name, _ := reader.ReadString('\n')
			t.newUser.Name = strings.TrimSpace(name)
			t.phase = PhaseUserEmailInput

		case PhaseUserEmailInput:
			fmt.Print("メールアドレスを入力 (user.email): ")
			email, _ := reader.ReadString('\n')
			t.newUser.Email = strings.TrimSpace(email)
			t.selectedUser = t.newUser
			t.phase = PhaseRemoteSelect

		case PhaseRemoteSelect:
			t.collectRemoteOptions()
			fmt.Println("\n--- リモートURLの選択 ---")
			for i, opt := range t.remoteOpts {
				fmt.Printf("[%d] %s\n", i+1, opt)
			}
			fmt.Printf("[%d] 新規入力 (SSH URL)\n", len(t.remoteOpts)+1)
			fmt.Printf("[%d] スキップ\n", len(t.remoteOpts)+2)
			fmt.Print("選択してください: ")

			var choice int
			fmt.Scanln(&choice)

			if choice > 0 && choice <= len(t.remoteOpts) {
				t.remoteURL = t.remoteOpts[choice-1]
				t.phase = PhaseConfirm
			} else if choice == len(t.remoteOpts)+1 {
				t.phase = PhaseRemoteCustomInput
			} else if choice == len(t.remoteOpts)+2 {
				t.remoteURL = ""
				t.phase = PhaseConfirm
			} else {
				fmt.Println("無効な選択肢です。")
			}

		case PhaseRemoteCustomInput:
			fmt.Print("SSH URLを手入力してください: ")
			url, _ := reader.ReadString('\n')
			t.remoteURL = strings.TrimSpace(url)
			t.phase = PhaseConfirm

		case PhaseConfirm:
			t.buildCommands()
			fmt.Println("\n--- 実行予定コマンド ---")
			for _, cmd := range t.commands {
				fmt.Println("  $", cmd)
			}
			fmt.Print("\n実行しますか？ (Enter: 実行 / Esc・q: 中止): ")
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(confirm)
			if confirm == "" || strings.ToLower(confirm) == "y" {
				t.phase = PhaseExecute
			} else {
				fmt.Println("処理を中止しました。")
				return nil
			}
		}
	}

	return t.executeCommands()
}

func (t *ToolGitInit) collectRemoteOptions() {
	var opts []string
	seen := make(map[string]bool)

	if repos, err := git.FetchGHRepos(); err == nil {
		for _, r := range repos {
			if r.SSHCloneURL != "" && !seen[r.SSHCloneURL] {
				seen[r.SSHCloneURL] = true
				opts = append(opts, r.SSHCloneURL)
			}
		}
	}

	if hosts, err := git.ParseSSHConfig(); err == nil {
		for _, h := range hosts {
			if !seen[h] {
				seen[h] = true
				opts = append(opts, h)
			}
		}
	}

	t.remoteOpts = opts
}

func (t *ToolGitInit) buildCommands() {
	t.commands = []string{
		fmt.Sprintf("git init %s", t.targetDir),
	}
	if t.selectedUser.Name != "" {
		t.commands = append(t.commands, fmt.Sprintf("git -C %s config user.name \"%s\"", t.targetDir, t.selectedUser.Name))
	}
	if t.selectedUser.Email != "" {
		t.commands = append(t.commands, fmt.Sprintf("git -C %s config user.email \"%s\"", t.targetDir, t.selectedUser.Email))
	}
	if t.remoteURL != "" {
		t.commands = append(t.commands, fmt.Sprintf("git -C %s remote add origin %s", t.targetDir, t.remoteURL))
	}
}

func (t *ToolGitInit) executeCommands() error {
	fmt.Println("\n--- 実行中 ---")
	for _, cmdStr := range t.commands {
		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

		var args []string
		for _, p := range parts[1:] {
			args = append(args, strings.Trim(p, "\""))
		}

		cmd := exec.Command(parts[0], args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("コマンド失敗: %s, error: %w", cmdStr, err)
		}
	}
	fmt.Println("正常に完了しました。")
	return nil
}
