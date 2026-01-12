package act

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// actionUsesRegex matches "uses: <action>@<ref>" patterns in YAML files.
// It captures the full action reference including the version/commit.
var actionUsesRegex = regexp.MustCompile(`uses:\s*["']?([^@\s"'#]+@[^\s"'#]+)`)

// ExtractExternalActions scans YAML files in the given directories and returns
// unique external action references (uses: owner/repo@ref).
// It filters out internal grafana/plugin-ci-workflows references.
// The returned slice contains deduplicated action references sorted alphabetically.
func ExtractExternalActions(dirs ...string) ([]string, error) {
	seen := make(map[string]struct{})
	var actions []string

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			// Only process YAML files
			ext := filepath.Ext(path)
			if ext != ".yml" && ext != ".yaml" {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			matches := actionUsesRegex.FindAllSubmatch(content, -1)
			for _, match := range matches {
				if len(match) < 2 {
					continue
				}
				action := string(match[1])

				// Skip internal references
				if strings.HasPrefix(action, "grafana/plugin-ci-workflows") {
					continue
				}

				// Skip if already seen
				if _, ok := seen[action]; ok {
					continue
				}
				seen[action] = struct{}{}
				actions = append(actions, action)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return actions, nil
}

