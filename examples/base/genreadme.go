package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type SubfolderContent struct {
	FolderName string
	Content    string
}

type ReadmeData struct {
	Subfolders []SubfolderContent
}

const (
	tmplFileName      = "README.tmpl"
	outputFileName    = "README.md"
	readmeStartMarker = "<!-- README start -->"
)

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current directory:", err)
	}

	var subfolders []SubfolderContent

	// Read all subdirectories
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		log.Fatal("Error reading directory:", err)
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
		content, err := extractContentAfterMarker(readmePath)
		if err != nil {
			log.Printf("Warning: Error processing %s: %v", readmePath, err)
			continue
		}

		if content != "" {
			subfolders = append(subfolders, SubfolderContent{
				FolderName: folderName,
				Content:    content,
			})
		}
	}

	// Load template from file
	tmpl, err := template.ParseFiles(tmplFileName)
	if err != nil {
		log.Fatal("Error parsing template file:", err)
	}

	// Generate the root README.md
	data := ReadmeData{Subfolders: subfolders}

	outputPath := filepath.Join(currentDir, outputFileName)
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Error creating %s: %v", outputFileName, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Warning: error closing file %s: %v", outputPath, cerr)
		}
	}()

	err = tmpl.Execute(file, data)
	if err != nil {
		log.Fatal("Error executing template:", err)
	}

	fmt.Printf("Generated %s with %d subfolders\n", outputFileName, len(subfolders))
}

func extractContentAfterMarker(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	foundMarker := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, readmeStartMarker) {
			foundMarker = true
			continue
		}

		if foundMarker {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if !foundMarker {
		return "", fmt.Errorf("marker %q not found", readmeStartMarker)
	}

	// Join lines and trim trailing whitespace
	content := strings.Join(lines, "\n")
	content = strings.TrimSpace(content)

	return content, nil
}
