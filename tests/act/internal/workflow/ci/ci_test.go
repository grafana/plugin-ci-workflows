package ci

import (
	"testing"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
	"github.com/stretchr/testify/require"
)

// newJobWithEmptyInputs returns a fresh Job with an initialized With map.
// SetCIInputs writes into job.With, so callers must ensure it is non-nil.
func newJobWithEmptyInputs() *workflow.Job {
	return &workflow.Job{With: map[string]any{}}
}

func TestSetCIInputs_AllowUnsignedInDev(t *testing.T) {
	t.Parallel()

	const inputKey = "allow-unsigned-in-dev"

	t.Run("nil leaves input unset (workflow default applies)", func(t *testing.T) {
		t.Parallel()

		job := newJobWithEmptyInputs()
		SetCIInputs(job, WorkflowInputs{})

		require.NotContains(t, job.With, inputKey,
			"unset AllowUnsignedInDev should not propagate %q so the workflow's default (false) applies", inputKey)
	})

	t.Run("true is forwarded as the escape-hatch opt-in", func(t *testing.T) {
		t.Parallel()

		job := newJobWithEmptyInputs()
		SetCIInputs(job, WorkflowInputs{
			AllowUnsignedInDev: workflow.Input(true),
		})

		require.Equal(t, true, job.With[inputKey],
			"AllowUnsignedInDev=true must propagate as %q=true to opt back into the pre-v8 untrusted-dev fallback", inputKey)
	})

	t.Run("false is forwarded explicitly", func(t *testing.T) {
		t.Parallel()

		job := newJobWithEmptyInputs()
		SetCIInputs(job, WorkflowInputs{
			AllowUnsignedInDev: workflow.Input(false),
		})

		require.Equal(t, false, job.With[inputKey],
			"AllowUnsignedInDev=false must propagate as %q=false to keep the hard-fail default", inputKey)
	})

	t.Run("does not collide with allow-unsigned", func(t *testing.T) {
		t.Parallel()

		job := newJobWithEmptyInputs()
		SetCIInputs(job, WorkflowInputs{
			AllowUnsigned:      workflow.Input(true),
			AllowUnsignedInDev: workflow.Input(false),
		})

		require.Equal(t, true, job.With["allow-unsigned"], "AllowUnsigned should map to %q", "allow-unsigned")
		require.Equal(t, false, job.With[inputKey], "AllowUnsignedInDev should map to %q", inputKey)
	})
}
