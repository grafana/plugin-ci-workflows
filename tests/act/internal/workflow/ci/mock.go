package ci

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
)

// mockWorkflowContextStep returns a Step that mocks the "workflow-context" step
// to return the given mocked Context.
func mockWorkflowContextStep(ctx Context) (workflow.Step, error) {
	ctxJSON, err := json.Marshal(ctx)
	if err != nil {
		return workflow.Step{}, fmt.Errorf("marshal workflow context to json: %w", err)
	}
	return workflow.Step{
		Name: "Determine workflow context (mocked)",
		Run: workflow.Commands{
			`echo "result=$RESULT" >> "$GITHUB_OUTPUT"`,
		}.String(),
		Env: map[string]string{
			"RESULT": string(ctxJSON),
		},
		Shell: "bash",
	}, nil
}
