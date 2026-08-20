package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// UserProfile は Git config から抽出したユーザープロファイルです。
type UserProfile struct {
	Profile string
	Name    string
	Email   string
	Source  string
}

// Label は TUI 上で表示するラベルを返します。
func (p UserProfile) Label() string {
	if p.Profile == "" || p.Profile == "default" {
		return fmt.Sprintf("%s <%s>", p.Name, p.Email)
	}

	return fmt.Sprintf("%s: %s <%s>", p.Profile, p.Name, p.Email)
}

type profile struct {
	Profile string
	Name    string
	Email   string
	Source  string
}

type configParser struct {
	visited  map[string]bool
	profiles map[string]*profile
}

// LoadUserProfiles は ~/.gitconfig と include/includeIf されたファイルから
// [user] の name/email を抽出します。
func LoadUserProfiles(path string) ([]UserProfile, error) {
	parser := &configParser{
		visited:  make(map[string]bool),
		profiles: make(map[string]*profile),
	}

	if err := parser.parseFile(path, "default"); err != nil {
		return nil, err
	}

	profiles := make([]UserProfile, 0, len(parser.profiles))

	for _, p := range parser.profiles {
		if strings.TrimSpace(p.Name) == "" ||
			strings.TrimSpace(p.Email) == "" {
			continue
		}

		profiles = append(profiles, UserProfile{
			Profile: p.Profile,
			Name:    p.Name,
			Email:   p.Email,
			Source:  p.Source,
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i].Profile == "default" {
			return true
		}
		if profiles[j].Profile == "default" {
			return false
		}

		return profiles[i].Label() < profiles[j].Label()
	})

	return profiles, nil
}

func (p *configParser) parseFile(path string, profileName string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	abs = filepath.Clean(abs)

	if p.visited[abs] {
		return nil
	}

	p.visited[abs] = true

	file, err := os.Open(abs)
	if err != nil {
		return err
	}
	defer file.Close()

	section := ""
	subsection := ""

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" ||
			strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") &&
			strings.HasSuffix(line, "]") {
			section, subsection = parseSection(line)
			continue
		}

		key, value, ok := parseAssignment(line)
		if !ok {
			continue
		}

		switch strings.ToLower(section) {
		case "include", "includeif":
			if strings.EqualFold(key, "path") {
				for _, includePath := range expandIncludePath(
					value,
					filepath.Dir(abs),
				) {
					if _, err := os.Stat(includePath); err != nil {
						continue
					}

					nextProfile := profileNameForPath(
						includePath,
						profileName,
					)

					if err := p.parseFile(
						includePath,
						nextProfile,
					); err != nil {
						return err
					}
				}
			}

		case "user":
			// 通常の [user] は default または include 元の
			// プロファイル名を利用します。
			name := profileName
			if subsection != "" {
				name = subsection
			}

			if name == "" {
				name = "default"
			}

			key := abs + "\x00" + name

			current := p.profiles[key]
			if current == nil {
				current = &profile{
					Profile: name,
					Source:  abs,
				}
				p.profiles[key] = current
			}

			switch strings.ToLower(key) {
			case "name":
				current.Name = value
			case "email":
				current.Email = value
			}
		}
	}

	return scanner.Err()
}

func parseSection(line string) (string, string) {
	body := strings.TrimSpace(line[1 : len(line)-1])

	if index := strings.IndexByte(body, '"'); index >= 0 {
		section := strings.TrimSpace(body[:index])
		subsection := strings.TrimSpace(body[index:])

		if value, err := strconv.Unquote(subsection); err == nil {
			return section, value
		}
	}

	return body, ""
}

func ParseGitConfig() ([]UserProfile, error) {
	nameCmd := exec.Command("git", "config", "--global", "--get", "user.name")
	emailCmd := exec.Command("git", "config", "--global", "--get", "user.email")

	name, err := nameCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git user.name: %w", err)
	}

	email, err := emailCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git user.email: %w", err)
	}

	return []UserProfile{
		{
			Name:  strings.TrimSpace(string(name)),
			Email: strings.TrimSpace(string(email)),
		},
	}, nil
}

func parseAssignment(line string) (string, string, bool) {
	index := strings.IndexByte(line, '=')

	if index < 0 {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return "", "", false
		}

		return fields[0], strings.Join(fields[1:], " "), true
	}

	key := strings.TrimSpace(line[:index])
	value := strings.TrimSpace(line[index+1:])

	if key == "" {
		return "", "", false
	}

	return key, cleanValue(value), true
}

func cleanValue(value string) string {
	value = strings.TrimSpace(value)

	if value == "" {
		return ""
	}

	if strings.HasPrefix(value, `"`) {
		if parsed, err := strconv.Unquote(value); err == nil {
			return parsed
		}
	}

	inQuote := false

	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '"':
			inQuote = !inQuote

		case '#', ';':
			if !inQuote &&
				(i == 0 ||
					value[i-1] == ' ' ||
					value[i-1] == '\t') {
				return strings.TrimSpace(value[:i])
			}
		}
	}

	return strings.TrimSpace(value)
}

func expandIncludePath(value string, baseDir string) []string {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	if strings.HasPrefix(value, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			value = filepath.Join(home, value[2:])
		}
	}

	if !filepath.IsAbs(value) {
		value = filepath.Join(baseDir, value)
	}

	matches, err := filepath.Glob(value)
	if err != nil {
		return nil
	}

	if len(matches) == 0 {
		return []string{value}
	}

	return matches
}

func profileNameForPath(path string, fallback string) string {
	base := filepath.Base(path)

	// ~/.gitconfig は default として扱います。
	if base == ".gitconfig" {
		return fallback
	}

	// 例:
	// ~/.gitconfig-personal -> personal
	// ~/.config/git/work -> work
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.TrimPrefix(base, ".gitconfig-")

	if base == "" {
		return fallback
	}

	return base
}
