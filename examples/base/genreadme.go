// Generate README.md files for each subdirectory based on the root README.md
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const generatedHeader = `<!-- THIS FILE IS AUTO-GENERATED. DO NOT EDIT MANUALLY. RUN: "go run genreadme.go" TO RE-GENERATE -->`

const fixedSection = `
> [!WARNING]
>
> Please [read the docs](https://enghub.grafana-ops.net/docs/default/component/grafana-plugins-platform/plugins-ci-github-actions/010-plugins-ci-github-actions) before using any of these workflows in your repository.
` +
	"The `yaml` files should be put in your repository's `.github/workflows` folder, and customized depending on your needs."

func main() {
	// Read the root README.md file
	rootReadme, err := os.ReadFile("README.md")
	if err != nil {
		log.Fatalf("Error reading README.md: %v", err)
	}

	// Parse sections from the root README
	sections := parseSections(string(rootReadme))

	// Get all subdirectories
	entries, err := os.ReadDir(".")
	if err != nil {
		log.Fatalf("Error reading current directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			folderName := entry.Name()

			// Check if there's a section for this folder
			if sectionContent, exists := sections[folderName]; exists {
				generateReadme(folderName, sectionContent)
				fmt.Printf("Generated README.md for folder: %s\n", folderName)
			}
		}
	}
}

func parseSections(content string) map[string]string {
	sections := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Regex to match section headers like ## [`simple`](./simple/)
	headerRegex := regexp.MustCompile(`^## \[` + "`" + `([^` + "`" + `]+)` + "`" + `\]\(\.\/[^\/]+\/\)`)

	var currentSection string
	var currentContent strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this line is a section header
		if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			// Save previous section if it exists
			if currentSection != "" {
				sections[currentSection] = strings.TrimSpace(currentContent.String())
			}

			// Start new section
			currentSection = matches[1]
			currentContent.Reset()
			continue
		}

		// If we're in a section, collect content until we hit another ## header or end
		if currentSection != "" {
			// Stop if we hit another ## header that's not our target format
			if strings.HasPrefix(line, "## ") && !headerRegex.MatchString(line) {
				sections[currentSection] = strings.TrimSpace(currentContent.String())
				currentSection = ""
				currentContent.Reset()
				continue
			}

			// Add line to current section content
			if currentContent.Len() > 0 {
				currentContent.WriteString("\n")
			}
			currentContent.WriteString(line)
		}
	}

	// Don't forget the last section
	if currentSection != "" {
		sections[currentSection] = strings.TrimSpace(currentContent.String())
	}

	return sections
}

func generateReadme(folderName, sectionContent string) {
	readmePath := filepath.Join(folderName, "README.md")

	// Create the content
	var builder strings.Builder
	builder.WriteString(generatedHeader)
	builder.WriteString("\n\n# ")
	builder.WriteString(folderName)
	builder.WriteString("\n")
	builder.WriteString(fixedSection)
	builder.WriteString("\n\n")
	builder.WriteString(sectionContent)

	// Write to file
	// Open file for writing
	file, err := os.Create(readmePath)
	if err != nil {
		log.Printf("Error creating README.md for %s: %v", folderName, err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(builder.String())
	if err != nil {
		log.Printf("Error writing README.md for %s: %v", folderName, err)
	}
}
