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

// Name returns the name of the event payload, e.g. "push" or "pull_request".
// If the name is not set, it returns an empty string.
func (e EventPayload) Name() string {
	if name, ok := e["event_name"].(string); ok {
		return name
	}
	return ""
}

// IsPush returns true if the event payload represents a `push` event.
func (e EventPayload) IsPush() bool {
	return e.Name() == "push"
}

// IsPullRequest returns true if the event payload represents a `pull_request` event.
func (e EventPayload) IsPullRequest() bool {
	return e.Name() == "pull_request"
}

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

// NewPushEventPayload creates a new EventPayload for a push event on the given branch.
func NewPushEventPayload(branch string) EventPayload {
	return NewEventPayload(map[string]any{
		"event_name": "push",
		"ref":        "refs/heads/" + branch,
	})
}

// NewPullRequestEventPayload creates a new EventPayload for a pull request event
// from a branch with the given name.
func NewPullRequestEventPayload(prBranch string) EventPayload {
	return NewEventPayload(map[string]any{
		"event_name": "pull_request",
		"head_ref":   prBranch,
		// "ref":        "refs/pull/1/merge",
	})
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
