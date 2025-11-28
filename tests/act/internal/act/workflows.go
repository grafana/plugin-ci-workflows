package act

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
)

// EventPayload represents the event payload to pass to act.
// It is a map of string keys to arbitrary values.
// It should mimic a GitHub event payload. By default, it includes an "act": true key-value pair
// which makes it possible to detect when the workflow is running under act in the workflow itself.
type EventPayload map[string]any

// NewEventPayload creates a new EventPayload with the given data.
// It always includes an "act": true key-value pair.
func NewEventPayload(data map[string]any) EventPayload {
	// Default data that should always be present in the payload
	data["act"] = true
	return EventPayload(data)
}

// NewEmptyEventPayload creates a new EventPayload with only the default "act": true key-value pair.
func NewEmptyEventPayload() EventPayload {
	return NewEventPayload(map[string]any{})
}

// CreateTempEventFile creates a temporary file in a temporary folder
// containing the given event payload in JSON format.
// The function returns the path to the created file.
// The caller is responsible for deleting the file when no longer needed.
func CreateTempEventFile(payload EventPayload) (string, error) {
	f, err := os.CreateTemp("", "act-*-event.json")
	if err != nil {
		return "", fmt.Errorf("create temp event file: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(payload); err != nil {
		return "", fmt.Errorf("encode event to temp file: %w", err)
	}
	return f.Name(), nil
}

// CreateTempWorkflowFile creates a temporary workflow file inside .github/workflows
// containing the given workflow marshaled to YAML, with the file name returned by workflow.FileName().
// The function returns the path to the created file.
// The caller is responsible for deleting the file when no longer needed.
func CreateTempWorkflowFile(workflow workflow.Workflow) (string, error) {
	content, err := workflow.Marshal()
	if err != nil {
		return "", fmt.Errorf("marshal workflow: %w", err)
	}
	fn := filepath.Join(".github", "workflows", workflow.FileName())
	if err := os.WriteFile(fn, content, 0o644); err != nil {
		return "", fmt.Errorf("write temp workflow file: %w", err)
	}
	// Create temporary child workflows if any
	for _, child := range workflow.Children() {
		if _, err := CreateTempWorkflowFile(child); err != nil {
			return "", fmt.Errorf("create child workflow file: %w", err)
		}
	}
	return fn, nil
}

// CleanupTempWorkflowFiles removes all temporary workflow files created for act tests
// that were created inside .github/workflows by CreateTempWorkflowFile.
func CleanupTempWorkflowFiles() error {
	files, err := filepath.Glob(filepath.Join(".github", "workflows", "act-*.yml"))
	if err != nil {
		return fmt.Errorf("glob old test workflow files: %w", err)
	}
	err = nil
	for _, f := range files {
		fmt.Printf("removing %q\n", f)
		err = errors.Join(err, os.Remove(f))
	}
	return err
}
