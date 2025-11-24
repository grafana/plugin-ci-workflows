package workflow

import "github.com/goccy/go-yaml"

const (
	pciwfMain = "grafana/plugin-ci-workflows/.github/workflows/ci.yml@main"
)

// SimpleCI is a predefined GitHub Actions workflow for testing plugins using act.
// It uses the plugin-ci-workflows CI workflow as a base, with sane default values
// and allows customization through options.
// It implements the Marshalable interface to allow conversion to YAML format.
// Instances must be created using NewSimpleCI.
type SimpleCI struct {
	Workflow
}

// NewSimpleCI creates a new SimpleCI workflow instance with default settings.
// The returned instance can be further customized using WithOption functions.
// For example:
//
// ```go
//
//	simpleCI := NewSimpleCI().With(
//	    WithJobName("My CI Job"),
//	    WithPluginDirectory("path/to/plugin"),
//	    WithDistArtifactPrefix("my-plugin-"),
//	    WithPlaywright(true),
//	)
//
// ```
func NewSimpleCI() SimpleCI {
	return SimpleCI{
		Workflow: Workflow{
			Name: "act",
			On: On{
				Push: OnPush{
					Branches: []string{"main"},
				},
				PullRequest: OnPullRequest{
					Branches: []string{"main"},
				},
			},
			Jobs: yaml.MapSlice{
				yaml.MapItem{
					Key: "ci",
					Value: &Job{
						Name: "CI",
						Uses: pciwfMain,
						Permissions: Permissions{
							"contents": "read",
							"id-token": "write",
						},
						With: map[string]any{
							"plugin-version-suffix": "${{ github.event_name == 'pull_request' && github.event.pull_request.head.sha || '' }}",
							"testing":               true,
						},
						Secrets: Secrets{
							"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
						},
					},
				},
			},
		},
	}
}

// With applies the given WithOption functions to the SimpleCI instance,
// returning the modified instance.
func (w SimpleCI) With(opts ...WithOption) SimpleCI {
	for _, opt := range opts {
		w = opt(w)
	}
	return w
}

// WithOption is a function that modifies a SimpleCI instance during its construction.
type WithOption func(SimpleCI) SimpleCI

// WithPluginDirectory sets the plugin-directory input for the CI job in the SimpleCI workflow.
func WithPluginDirectory(dir string) WithOption {
	return func(w SimpleCI) SimpleCI {
		w.FirstJob().With["plugin-directory"] = dir
		return w
	}
}

// WithDistArtifactPrefix sets the dist-artifacts-prefix input for the CI job in the SimpleCI workflow.
func WithDistArtifactPrefix(prefix string) WithOption {
	return func(w SimpleCI) SimpleCI {
		w.FirstJob().With["dist-artifacts-prefix"] = prefix
		return w
	}
}

// WithPlaywright sets the run-playwright input for the CI job in the SimpleCI workflow.
func WithPlaywright(enabled bool) WithOption {
	return func(w SimpleCI) SimpleCI {
		w.FirstJob().With["run-playwright"] = enabled
		return w
	}
}

// Static checks

var _ Marshalable = SimpleCI{}
