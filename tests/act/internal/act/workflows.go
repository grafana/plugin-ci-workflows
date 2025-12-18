package act

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
)

// EventKind represents the kind of GitHub event, e.g., "push" or "pull_request".
type EventKind string

// EventKind enum values

const (
	EventKindPush        EventKind = "push"
	EventKindPullRequest EventKind = "pull_request"
)

// Event represents the event with a name and payload to pass to act.
// It should mimic a GitHub event payload. For example, a pull_request
// event payload should follow the GitHub pull_request webhook event structure:
// https://docs.github.com/en/webhooks/webhook-events-and-payloads#pull_request
//
// By default, the payload includes an additional `{ "act": true }` key-value pair in the payload,
// which makes it possible to detect when the workflow is running under act in the workflow itself:
//
// ```yaml
//
//	if: ${{ github.event.act == true }}
//
// ```
type Event struct {
	// Kind is the type of the event, e.g., "push" or "pull_request".
	Kind EventKind

	// Actor is the GitHub username of the actor that triggered the event.
	// It is optional and can be used to simulate different users triggering the event.
	// The default (empty) will use `nektos/act`.
	Actor string

	// Payload is the event payload data (JSON serializable).
	// See the GitHub "webhooks and events payload" documentation
	// for the schema of different event payloads:
	// https://docs.github.com/en/webhooks/webhook-events-and-payloads
	Payload map[string]any
}

// EventOption is a function that configures an Event.
type EventOption func(e *Event)

// WithEventActor sets the actor of the Event, in order to impersonate
// different users triggering the event when running the workflow with act.
func WithEventActor(actor string) EventOption {
	return func(e *Event) {
		e.Actor = actor
	}
}

// NewEventPayload creates a new EventPayload with the given data.
// It always includes an "act": true key-value pair.
func NewEventPayload(kind EventKind, data map[string]any, opts ...EventOption) Event {
	if data == nil {
		data = map[string]any{}
	}
	e := Event{
		Kind:    kind,
		Payload: data,
	}
	// Default data that should always be present in the payload
	e.Payload["act"] = true
	// Apply options
	for _, opt := range opts {
		opt(&e)
	}
	return e
}

// NewEmptyEventPayload creates a new default "push" EventPayload with only the default "act": true key-value pair.
//
// Deprecated: use NewEventPayload instead.
func NewEmptyEventPayload(opts ...EventOption) Event {
	return NewEventPayload(EventKindPush, map[string]any{}, opts...)
}

// NewPushEventPayload creates a new EventPayload for a push event on the given branch.
func NewPushEventPayload(branch string, opts ...EventOption) Event {
	return NewEventPayload(EventKindPush, map[string]any{
		"ref": "refs/heads/" + branch,
	}, opts...)
}

// NewPullRequestEventPayload creates a new EventPayload for a pull request event
// from a branch with the given name.
func NewPullRequestEventPayload(prBranch string, opts ...EventOption) Event {
	return NewEventPayload(EventKindPullRequest, map[string]any{
		"action": "opened",
		"pull_request": map[string]any{
			"head": map[string]any{
				"ref": prBranch,
			},
			"base": map[string]any{
				"ref": "main",
			},
		},
	}, opts...)
}

// CreateTempEventFile creates a temporary file in a temporary folder
// containing the payload from the given event in JSON format.
// The function returns the path to the created file.
// The caller is responsible for deleting the file when no longer needed.
func CreateTempEventFile(event Event) (string, error) {
	f, err := os.CreateTemp("", "act-*-event.json")
	if err != nil {
		return "", fmt.Errorf("create temp event file: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(event.Payload); err != nil {
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
