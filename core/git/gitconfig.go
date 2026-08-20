package git

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// UserProfile は gitconfig から抽出した名前とメールのペアを保持します。
type UserProfile struct {
	Name  string
	Email string
}

// ParseGitConfig は ~/.gitconfig および include/includeIf で参照されている設定ファイルをパースし、
// 検出されたユーザープロファイルの一覧を返します。
func ParseGitConfig() ([]UserProfile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	mainConfigPath := filepath.Join(home, ".gitconfig")
	profiles := make([]UserProfile, 0)
	visited := make(map[string]bool)

	parseConfigFile(mainConfigPath, &profiles, visited)

	// 重複プロファイルの除外処理
	uniqueProfiles := make([]UserProfile, 0, len(profiles))
	seen := make(map[string]bool)
	for _, p := range profiles {
		if p.Name == "" && p.Email == "" {
			continue
		}
		key := p.Name + "<" + p.Email + ">"
		if !seen[key] {
			seen[key] = true
			uniqueProfiles = append(uniqueProfiles, p)
		}
	}

	return uniqueProfiles, nil
}

func parseConfigFile(filePath string, profiles *[]UserProfile, visited map[string]bool) {
	cleanPath := filepath.Clean(filePath)
	if visited[cleanPath] {
		return
	}
	visited[cleanPath] = true

	file, err := os.Open(cleanPath)
	if err != nil {
		return // ファイルが存在しない、または読み込めない場合はスキップ（フォールバック）
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentSection string
	var currentProfile UserProfile
	dir := filepath.Dir(cleanPath)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 空行やコメントをスキップ
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// セクションの判定
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if currentProfile.Name != "" || currentProfile.Email != "" {
				*profiles = append(*profiles, currentProfile)
				currentProfile = UserProfile{}
			}
			currentSection = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}

		// キーと値の分割
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"") // クォーテーションの削除

		// [user] セクション内の解析
		if currentSection == "user" {
			if key == "name" {
				currentProfile.Name = val
			} else if key == "email" {
				currentProfile.Email = val
			}
		}

		// [include] または [includeIf ...] セクション内の path 属性のパース
		if strings.HasPrefix(currentSection, "include") && key == "path" {
			incPath := resolvePath(val, dir)
			parseConfigFile(incPath, profiles, visited)
		}
	}

	if currentProfile.Name != "" || currentProfile.Email != "" {
		*profiles = append(*profiles, currentProfile)
	}
}

func resolvePath(pathStr, baseDir string) string {
	if strings.HasPrefix(pathStr, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, pathStr[2:])
		}
	}
	if !filepath.IsAbs(pathStr) {
		return filepath.Join(baseDir, pathStr)
	}
	return pathStr
}
