package git

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ParseSSHConfig は ~/.ssh/config をパースし、定義されている Host の一覧を抽出します。
func ParseSSHConfig() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".ssh", "config")
	file, err := os.Open(configPath)
	if err != nil {
		return []string{}, nil // ファイルがない場合は空のリストを返す（フォールバック）
	}
	defer file.Close()

	var hosts []string
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.EqualFold(fields[0], "Host") {
			for _, host := range fields[1:] {
				// ワイルドカード指定は除外
				if strings.Contains(host, "*") || strings.Contains(host, "?") {
					continue
				}
				if !seen[host] {
					seen[host] = true
					hosts = append(hosts, host)
				}
			}
		}
	}

	return hosts, nil
}
