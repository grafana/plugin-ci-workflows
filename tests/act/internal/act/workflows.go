package act

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type EventPayload map[string]any

func NewEventPayload(data map[string]any) EventPayload {
	// Default data that should always be present in the payload
	data["act"] = true
	return EventPayload(data)
}

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

func CreateTempWorkflowFile(content []byte) (string, error) {
	fn := "act-" + uuid.NewString() + ".yml"
	fn = filepath.Join(".github", "workflows", fn)
	if err := os.WriteFile(fn, content, 0o644); err != nil {
		return "", fmt.Errorf("write temp workflow file: %w", err)
	}
	return fn, nil
}

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
