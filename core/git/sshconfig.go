package git

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ParseSSHConfig は ~/.ssh/config を読み込み、
// SSH Host alias の一覧を返します。
func ParseSSHConfig() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(
		home,
		".ssh",
		"config",
	)

	file, err := os.Open(configPath)
	if err != nil {
		// config が存在しない場合は空リスト。
		return []string{}, nil
	}
	defer file.Close()

	var hosts []string

	seen := make(map[string]bool)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(
			scanner.Text(),
		)

		if line == "" ||
			strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		if !strings.EqualFold(
			fields[0],
			"Host",
		) {
			continue
		}

		for _, host := range fields[1:] {
			// ワイルドカードは選択候補にしない。
			if strings.ContainsAny(
				host,
				"*?!",
			) {
				continue
			}

			if seen[host] {
				continue
			}

			seen[host] = true
			hosts = append(
				hosts,
				host,
			)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}
