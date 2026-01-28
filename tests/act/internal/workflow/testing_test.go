package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const testJobID = "test-job"

// newTestWorkflow creates a TestingWorkflow with a job containing steps A, B, C, D for testing.
func newTestWorkflow(opts ...TestingWorkflowOption) *TestingWorkflow {
	return NewTestingWorkflow("test", BaseWorkflow{
		Jobs: map[string]*Job{
			testJobID: {
				Steps: Steps{
					{ID: "step-a", Name: "Step A"},
					{ID: "step-b", Name: "Step B"},
					{ID: "step-c", Name: "Step C"},
					{ID: "step-d", Name: "Step D"},
				},
			},
		},
	}, opts...)
}

// stepIDs extracts the IDs from the steps for easy comparison.
func stepIDs(steps Steps) []string {
	ids := make([]string, len(steps))
	for i, s := range steps {
		ids[i] = s.ID
	}
	return ids
}

func TestWithInjectedSteps(t *testing.T) {
	t.Run("PositionBefore with InjectionStepID", func(t *testing.T) {
		twf := newTestWorkflow(
			WithInjectedSteps(t, testJobID, InjectedStepsOptions{
				Position:        InjectedStepsOptionsPositionBefore,
				InjectionStepID: "step-c",
				Steps: Steps{
					{ID: "new-1", Name: "New Step 1"},
					{ID: "new-2", Name: "New Step 2"},
				},
			}),
		)
		require.Equal(t, []string{"step-a", "step-b", "new-1", "new-2", "step-c", "step-d"}, stepIDs(twf.BaseWorkflow.Jobs[testJobID].Steps))
	})

	t.Run("PositionAfter with InjectionStepID", func(t *testing.T) {
		twf := newTestWorkflow(
			WithInjectedSteps(t, testJobID, InjectedStepsOptions{
				Position:        InjectedStepsOptionsPositionAfter,
				InjectionStepID: "step-b",
				Steps: Steps{
					{ID: "new-1", Name: "New Step 1"},
					{ID: "new-2", Name: "New Step 2"},
				},
			}),
		)
		require.Equal(t, []string{"step-a", "step-b", "new-1", "new-2", "step-c", "step-d"}, stepIDs(twf.BaseWorkflow.Jobs[testJobID].Steps))
	})

	t.Run("PositionBefore with InjectionStepIndex", func(t *testing.T) {
		twf := newTestWorkflow(
			WithInjectedSteps(t, testJobID, InjectedStepsOptions{
				Position:           InjectedStepsOptionsPositionBefore,
				InjectionStepIndex: 2, // step-c
				Steps: Steps{
					{ID: "new-1", Name: "New Step 1"},
					{ID: "new-2", Name: "New Step 2"},
				},
			}),
		)
		require.Equal(t, []string{"step-a", "step-b", "new-1", "new-2", "step-c", "step-d"}, stepIDs(twf.BaseWorkflow.Jobs[testJobID].Steps))
	})

	t.Run("PositionAfter with InjectionStepIndex", func(t *testing.T) {
		twf := newTestWorkflow(
			WithInjectedSteps(t, testJobID, InjectedStepsOptions{
				Position:           InjectedStepsOptionsPositionAfter,
				InjectionStepIndex: 1, // step-b
				Steps: Steps{
					{ID: "new-1", Name: "New Step 1"},
					{ID: "new-2", Name: "New Step 2"},
				},
			}),
		)
		require.Equal(t, []string{"step-a", "step-b", "new-1", "new-2", "step-c", "step-d"}, stepIDs(twf.BaseWorkflow.Jobs[testJobID].Steps))
	})

	t.Run("inject before first step", func(t *testing.T) {
		twf := newTestWorkflow(
			WithInjectedSteps(t, testJobID, InjectedStepsOptions{
				Position:        InjectedStepsOptionsPositionBefore,
				InjectionStepID: "step-a",
				Steps: Steps{
					{ID: "new-1", Name: "New Step 1"},
					{ID: "new-2", Name: "New Step 2"},
				},
			}),
		)
		require.Equal(t, []string{"new-1", "new-2", "step-a", "step-b", "step-c", "step-d"}, stepIDs(twf.BaseWorkflow.Jobs[testJobID].Steps))
	})

	t.Run("inject after last step", func(t *testing.T) {
		twf := newTestWorkflow(
			WithInjectedSteps(t, testJobID, InjectedStepsOptions{
				Position:        InjectedStepsOptionsPositionAfter,
				InjectionStepID: "step-d",
				Steps: Steps{
					{ID: "new-1", Name: "New Step 1"},
					{ID: "new-2", Name: "New Step 2"},
				},
			}),
		)
		require.Equal(t, []string{"step-a", "step-b", "step-c", "step-d", "new-1", "new-2"}, stepIDs(twf.BaseWorkflow.Jobs[testJobID].Steps))
	})

	t.Run("inject after last step using -1 index", func(t *testing.T) {
		twf := newTestWorkflow(
			WithInjectedSteps(t, testJobID, InjectedStepsOptions{
				Position:           InjectedStepsOptionsPositionAfter,
				InjectionStepIndex: -1,
				Steps: Steps{
					{ID: "new-1", Name: "New Step 1"},
					{ID: "new-2", Name: "New Step 2"},
				},
			}),
		)
		require.Equal(t, []string{"step-a", "step-b", "step-c", "step-d", "new-1", "new-2"}, stepIDs(twf.BaseWorkflow.Jobs[testJobID].Steps))
	})
}
