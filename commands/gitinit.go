package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"tui-go/core"
	"tui-go/core/git"

	tea "charm.land/bubbletea/v2"
)

const (
	gitInitTargetStage = iota
	gitInitUserStage
	gitInitUserNameStage
	gitInitUserEmailStage
	gitInitSSHStage
	gitInitSSHCustomStage
	gitInitRepoStage
	gitInitRepoCustomStage
	gitInitConfirmStage
)

const (
	newInputOption = "新規入力"
	skipOption     = "スキップ"
	executeOption  = "実行"
	cancelOption   = "中止"
)

type ToolGitInit struct {
	runner core.StageRunner

	userProfiles []git.UserProfile
	sshHosts     []string
	repositories []git.GHRepo

	selectedUser git.UserProfile
	selectedSSH  string
	selectedRepo string

	remoteURL string
	targetDir string
}

func NewToolGitInit() core.Tool {
	return &ToolGitInit{}
}

func (t *ToolGitInit) Init() tea.Cmd {
	t.userProfiles = nil
	t.sshHosts = nil
	t.repositories = nil

	t.selectedUser = git.UserProfile{}
	t.selectedSSH = ""
	t.selectedRepo = ""
	t.remoteURL = ""
	t.targetDir = "."

	// ~/.gitconfig からユーザープロファイルを取得
	if profiles, err := git.ParseGitConfig(); err == nil {
		t.userProfiles = profiles
	}

	// ~/.ssh/config から SSH Host を取得
	if hosts, err := git.ParseSSHConfig(); err == nil {
		t.sshHosts = hosts
	}

	// gh repo list から Repository を取得
	if repos, err := git.FetchGHRepos(); err == nil {
		t.repositories = repos
	}

	t.runner = core.NewStageRunner(
		"Git Init",
		t.buildStages(),
	)

	return t.runner.Init()
}

func (t *ToolGitInit) buildStages() []core.Stage {
	// -------------------------
	// User
	// -------------------------

	userOptions := make([]string, 0, len(t.userProfiles)+1)

	for _, profile := range t.userProfiles {
		userOptions = append(
			userOptions,
			formatUserProfile(profile),
		)
	}

	userOptions = append(
		userOptions,
		newInputOption,
	)

	// -------------------------
	// SSH
	// -------------------------

	sshOptions := make(
		[]string,
		0,
		len(t.sshHosts)+1,
	)

	sshOptions = append(
		sshOptions,
		t.sshHosts...,
	)

	sshOptions = append(
		sshOptions,
		newInputOption,
	)

	// -------------------------
	// Repository
	// -------------------------

	repoOptions := make(
		[]string,
		0,
		len(t.repositories)+2,
	)

	seen := make(map[string]bool)

	for _, repo := range t.repositories {
		name := repo.NameWithOwner

		if name == "" {
			continue
		}

		if seen[name] {
			continue
		}

		seen[name] = true
		repoOptions = append(
			repoOptions,
			name,
		)
	}

	repoOptions = append(
		repoOptions,
		newInputOption,
	)

	repoOptions = append(
		repoOptions,
		skipOption,
	)

	return []core.Stage{
		core.NewTextInputStage(
			"Target directory",
			".",
			50,
		),

		core.NewChoiceStage(
			"Git user",
			userOptions,
			core.SelectSingle,
		),

		core.NewTextInputStage(
			"User name",
			"user.name",
			50,
		),

		core.NewTextInputStage(
			"User email",
			"user.email",
			50,
		),

		core.NewChoiceStage(
			"SSH Host",
			sshOptions,
			core.SelectSingle,
		),

		core.NewTextInputStage(
			"SSH Host",
			"例: github-work",
			50,
		),

		core.NewChoiceStage(
			"Repository",
			repoOptions,
			core.SelectSingle,
		),

		core.NewTextInputStage(
			"Repository",
			"owner/repository",
			60,
		),

		core.NewChoiceStage(
			"Confirm",
			[]string{
				executeOption,
				cancelOption,
			},
			core.SelectSingle,
		),
	}
}

func (t *ToolGitInit) Update(
	msg tea.Msg,
) (core.Tool, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)

	if !ok {
		return t, nil
	}

	advanced, canceled, cmd :=
		t.runner.HandleKey(km)

	if canceled {
		return t, core.ToolDone(
			"",
			nil,
		)
	}

	if !advanced {
		return t, cmd
	}

	switch t.runner.Idx - 1 {
	case gitInitTargetStage:
		t.handleTargetDir()

	case gitInitUserStage:
		return t.handleUserSelection()

	case gitInitUserNameStage:
		t.handleUserName()

	case gitInitUserEmailStage:
		t.handleUserEmail()

	case gitInitSSHStage:
		return t.handleSSHSelection()

	case gitInitSSHCustomStage:
		t.handleSSHCustom()

	case gitInitRepoStage:
		return t.handleRepositorySelection()

	case gitInitRepoCustomStage:
		t.handleRepositoryCustom()

	case gitInitConfirmStage:
		return t.handleConfirm()
	}

	return t, cmd
}

func (t *ToolGitInit) View() tea.View {
	return t.runner.View()
}

// -------------------------
// User
// -------------------------

func (t *ToolGitInit) handleTargetDir() {
	value :=
		strings.TrimSpace(
			t.runner.Stages[gitInitTargetStage].Result,
		)

	if value == "" {
		value = "."
	}

	t.targetDir = value
}

func (t *ToolGitInit) handleUserSelection() (core.Tool, tea.Cmd) {
	result :=
		t.runner.Stages[gitInitUserStage].Result

	if result == newInputOption {
		// User name 入力へ進む
		t.runner.Idx = gitInitUserNameStage
		t.runner.Stages[gitInitUserNameStage].TextInput.Focus()

		return t, nil
	}

	for _, profile := range t.userProfiles {
		if formatUserProfile(profile) == result {
			t.selectedUser = profile
			break
		}
	}

	// 既存ユーザーなら User name / email 入力をスキップ
	t.runner.Idx = gitInitSSHStage

	return t, nil
}

func (t *ToolGitInit) handleUserName() {
	t.selectedUser.Name =
		strings.TrimSpace(
			t.runner.Stages[gitInitUserNameStage].Result,
		)
}

func (t *ToolGitInit) handleUserEmail() {
	t.selectedUser.Email =
		strings.TrimSpace(
			t.runner.Stages[gitInitUserEmailStage].Result,
		)
}

func formatUserProfile(
	profile git.UserProfile,
) string {
	return fmt.Sprintf(
		"%s <%s>",
		profile.Name,
		profile.Email,
	)
}

// -------------------------
// SSH
// -------------------------

func (t *ToolGitInit) handleSSHSelection() (core.Tool, tea.Cmd) {
	result :=
		t.runner.Stages[gitInitSSHStage].Result

	if result == newInputOption {
		t.runner.Idx = gitInitSSHCustomStage
		t.runner.Stages[gitInitSSHCustomStage].TextInput.Focus()

		return t, nil
	}

	t.selectedSSH = result

	// SSH を選択してから Repository へ進む
	t.runner.Idx = gitInitRepoStage

	return t, nil
}

func (t *ToolGitInit) handleSSHCustom() {
	t.selectedSSH =
		strings.TrimSpace(
			t.runner.Stages[gitInitSSHCustomStage].Result,
		)

	// SSH Host を決定してから Repository 選択へ
	t.runner.Idx = gitInitRepoStage
}

// -------------------------
// Repository
// -------------------------

func (t *ToolGitInit) handleRepositorySelection() (core.Tool, tea.Cmd) {
	result :=
		t.runner.Stages[gitInitRepoStage].Result

	switch result {
	case newInputOption:
		t.runner.Idx = gitInitRepoCustomStage
		t.runner.Stages[gitInitRepoCustomStage].TextInput.Focus()

		return t, nil

	case skipOption:
		t.selectedRepo = ""
		t.remoteURL = ""

		t.runner.Idx = gitInitConfirmStage

		return t, nil
	}

	t.selectedRepo = result

	t.remoteURL =
		buildSSHRemoteURL(
			t.selectedSSH,
			t.selectedRepo,
		)

	t.runner.Idx = gitInitConfirmStage

	return t, nil
}

func (t *ToolGitInit) handleRepositoryCustom() {
	repository :=
		strings.TrimSpace(
			t.runner.Stages[gitInitRepoCustomStage].Result,
		)

	t.selectedRepo = repository

	if repository == "" {
		t.remoteURL = ""
	} else {
		t.remoteURL =
			buildSSHRemoteURL(
				t.selectedSSH,
				repository,
			)
	}

	t.runner.Idx = gitInitConfirmStage
}

func buildSSHRemoteURL(
	sshHost string,
	repository string,
) string {
	sshHost =
		strings.TrimSpace(sshHost)

	repository =
		strings.TrimSpace(repository)

	repository =
		strings.TrimPrefix(
			repository,
			"/",
		)

	repository =
		strings.TrimSuffix(
			repository,
			".git",
		)

	if sshHost == "" ||
		repository == "" {
		return ""
	}

	return fmt.Sprintf(
		"git@%s:%s.git",
		sshHost,
		repository,
	)
}

// -------------------------
// Confirm
// -------------------------

func (t *ToolGitInit) handleConfirm() (core.Tool, tea.Cmd) {
	result :=
		t.runner.Stages[gitInitConfirmStage].Result

	if result == cancelOption {
		return t, core.ToolDone(
			"",
			nil,
		)
	}

	if result != executeOption {
		return t, nil
	}

	return t, t.execute()
}

// -------------------------
// Execute
// -------------------------

func (t *ToolGitInit) execute() tea.Cmd {
	targetDir := t.targetDir
	user := t.selectedUser
	remoteURL := t.remoteURL

	return func() tea.Msg {
		if err := runGitInit(
			targetDir,
			user,
			remoteURL,
		); err != nil {
			return core.ToolDoneMsg{
				Result: "",
				Err:    err,
			}
		}

		return core.ToolDoneMsg{
			Result: buildSuccessMessage(
				targetDir,
				remoteURL,
			),
			Err: nil,
		}
	}
}

func runGitInit(
	targetDir string,
	user git.UserProfile,
	remoteURL string,
) error {
	if err := runCommand(
		"git",
		"init",
		targetDir,
	); err != nil {
		return err
	}

	if user.Name != "" {
		if err := runCommand(
			"git",
			"-C",
			targetDir,
			"config",
			"user.name",
			user.Name,
		); err != nil {
			return err
		}
	}

	if user.Email != "" {
		if err := runCommand(
			"git",
			"-C",
			targetDir,
			"config",
			"user.email",
			user.Email,
		); err != nil {
			return err
		}
	}

	if remoteURL != "" {
		if err := runCommand(
			"git",
			"-C",
			targetDir,
			"remote",
			"add",
			"origin",
			remoteURL,
		); err != nil {
			return err
		}
	}

	return nil
}

func runCommand(
	name string,
	args ...string,
) error {
	cmd :=
		exec.Command(
			name,
			args...,
		)

	output, err :=
		cmd.CombinedOutput()

	if err != nil {
		message :=
			strings.TrimSpace(
				string(output),
			)

		if message != "" {
			return fmt.Errorf(
				"%s: %w: %s",
				name,
				err,
				message,
			)
		}

		return fmt.Errorf(
			"%s: %w",
			name,
			err,
		)
	}

	return nil
}

func buildSuccessMessage(
	targetDir string,
	remoteURL string,
) string {
	if remoteURL == "" {
		return fmt.Sprintf(
			"Git Init completed: %s",
			targetDir,
		)
	}

	return fmt.Sprintf(
		"Git Init completed: %s → %s",
		targetDir,
		remoteURL,
	)
}

var _ core.Tool = (*ToolGitInit)(nil)
