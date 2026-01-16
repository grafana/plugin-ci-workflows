package act

import "time"

// logLine represents a single line in act's JSON log output.
type logLine struct {
	DryRun bool   `json:"dryrun"`
	Job    string `json:"job"`
	JobID  string `json:"jobID"`
	Level  string `json:"level"`
	// TODO: type
	Matrix  any       `json:"matrix"`
	Message string    `json:"msg"`
	Stage   string    `json:"stage"`
	Step    string    `json:"step"`
	StepID  []string  `json:"stepID"`
	Time    time.Time `json:"time"`

	// Intercepted GHA commands

	Command string `json:"command,omitempty"`
	Name    string `json:"name,omitempty"`

	// GHA annotation command
	Arg     string            `json:"arg,omitempty"`
	KvPairs map[string]string `json:"kvPairs,omitempty"`

	// GHA summary command
	Content string `json:"content,omitempty"`
}
