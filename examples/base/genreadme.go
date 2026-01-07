package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

type SubfolderContent struct {
	FolderName string
	Content    string
	Order      int
}

type ReadmeData struct {
	Subfolders []SubfolderContent
}

const (
	tmplFileName      = "README.tmpl"
	outputFileName    = "README.md"
	readmeStartMarker = "<!-- README start -->"
)

var orderRegex = regexp.MustCompile(`<!--\s*order:\s*(\d+)\s*-->`)

func _main() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	var subfolders []SubfolderContent

	// Read all subdirectories
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		readmePath := filepath.Join(currentDir, folderName, "README.md")

		// Check if README.md exists in the subfolder
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			continue
		}

		// Extract content after "<!-- README start -->"
		content, order, err := extractContentAfterMarker(readmePath)
		if err != nil {
			log.Printf("Warning: Error processing %s: %v", readmePath, err)
			continue
		}

		if content != "" {
			subfolders = append(subfolders, SubfolderContent{
				FolderName: folderName,
				Content:    content,
				Order:      order,
			})
		}
	}

	// Sort subfolders by order
	sort.Slice(subfolders, func(i, j int) bool {
		return subfolders[i].Order < subfolders[j].Order
	})

	// Load template from file
	tmpl, err := template.ParseFiles(tmplFileName)
	if err != nil {
		return fmt.Errorf("error parsing template file: %w", err)
	}

	// Generate the root README.md
	data := ReadmeData{Subfolders: subfolders}

	outputPath := filepath.Join(currentDir, outputFileName)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", outputFileName, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	fmt.Printf("Generated %s with %d subfolders\n", outputFileName, len(subfolders))
	return nil
}

func extractContentAfterMarker(filePath string) (content string, order int, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	var lines []string
	foundMarker := false
	order = 999 // Default order for files without order comment

	for scanner.Scan() {
		line := scanner.Text()

		// Check for order comment
		if matches := orderRegex.FindStringSubmatch(line); matches != nil {
			if orderNum, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
				order = orderNum
			}
			continue // Don't include the order comment in the output
		}

		if strings.Contains(line, readmeStartMarker) {
			foundMarker = true
			continue
		}

		if foundMarker {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", 0, err
	}

	if !foundMarker {
		return "", 0, fmt.Errorf("marker %q not found", readmeStartMarker)
	}

	// Join lines and trim trailing whitespace
	content = strings.Join(lines, "\n")
	content = strings.TrimSpace(content)

	return content, order, nil
}

func main() {
	if err := _main(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
