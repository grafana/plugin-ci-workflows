package workflow

const (
	pciwfMain = "grafana/plugin-ci-workflows/.github/workflows/ci.yml@main"
	ciJob     = "ci"
)

type SimpleCI struct {
	Workflow
}

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
			Jobs: map[string]Job{
				ciJob: {
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
	}
}

type WithOption func(SimpleCI) SimpleCI

func WithJobName(name string) WithOption {
	return func(w SimpleCI) SimpleCI {
		job := w.Jobs[ciJob]
		job.Name = name
		w.Jobs[ciJob] = job
		return w
	}
}

func WithPluginDirectory(dir string) WithOption {
	return func(w SimpleCI) SimpleCI {
		w.Jobs[ciJob].With["plugin-directory"] = dir
		return w
	}
}

func WithDistArtifactPrefix(prefix string) WithOption {
	return func(w SimpleCI) SimpleCI {
		w.Jobs[ciJob].With["dist-artifacts-prefix"] = prefix
		return w
	}
}

func WithPlaywright(enabled bool) WithOption {
	return func(w SimpleCI) SimpleCI {
		w.Jobs[ciJob].With["run-playwright"] = enabled
		return w
	}
}

func (w SimpleCI) With(opts ...WithOption) SimpleCI {
	for _, opt := range opts {
		w = opt(w)
	}
	return w
}

// Static checks

var _ Marshalable = SimpleCI{}
