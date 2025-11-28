package act

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/grafana/plugin-ci-workflows/tests/act/internal/workflow"
)

var (
	logUUIDRegex           = regexp.MustCompile(`-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	selfHostedRunnerLabels = [...]string{
		"ubuntu-x64-small",
		"ubuntu-x64",
		"ubuntu-x64-large",
		"ubuntu-x64-xlarge",
		"ubuntu-x64-2xlarge",
		"ubuntu-arm64-small",
		"ubuntu-arm64",
		"ubuntu-arm64-large",
		"ubuntu-arm64-xlarge",
		"ubuntu-arm64-2xlarge",
	}
)

const nektosActRunnerImage = "ghcr.io/catthehacker/ubuntu:act-latest"

// Runner is a test runner that can execute GitHub Actions workflows using act.
type Runner struct {
	// t is the testing.T instance for the current test.
	t                *testing.T
	uuid             uuid.UUID
	ArtifactsStorage ArtifactsStorage

	// gitHubToken is the token used to authenticate with GitHub.
	gitHubToken string

	// ConcurrentJobs defines the number of jobs to run concurrently via act.
	// By default (0), act uses the number of CPU cores.
	ConcurrentJobs int

	// Verbose enables logging of JSON output from act back to stdout.
	Verbose bool
}

// NewRunner creates a new Runner instance.
func NewRunner(t *testing.T) (*Runner, error) {
	// Get GitHub token from environment (GHA) or gh CLI (local)
	ghToken, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok || ghToken == "" {
		cmd := exec.Command("gh", "auth", "token")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("exec 'gh auth token': %w", err)
		}
		ghToken = strings.TrimSpace(string(output))
	}
	r := &Runner{
		t:           t,
		uuid:        uuid.New(),
		gitHubToken: ghToken,
	}
	if err := r.checkExecutables(); err != nil {
		return nil, err
	}
	r.ArtifactsStorage = newArtifactsStorage(r)
	return r, nil
}

// args returns the CLI arguments to pass to act for the given workflow and event payload files.
func (r *Runner) args(workflowFile string, payloadFile string) ([]string, error) {
	artifactServerPort, err := getFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port for artifact server: %w", err)
	}

	pciwfRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	f, err := os.Open(".release-please-manifest.json")
	if err != nil {
		return nil, fmt.Errorf("open .release-please-manifest.json: %w", err)
	}
	defer f.Close()
	var releasePleaseManifest map[string]string
	if err := json.NewDecoder(f).Decode(&releasePleaseManifest); err != nil {
		return nil, fmt.Errorf("decode release-please-config.json: %w", err)
	}
	releasePleaseTag := "ci-cd-workflows/v" + releasePleaseManifest[".github/workflows"]
	args := []string{
		"-W", workflowFile,
		"-e", payloadFile,
		"--rm",
		"--json",

		// Per-runner unique cache and toolcache paths to avoid conflicts when running tests in parallel
		"--action-cache-path=/tmp/act-cache/" + r.uuid.String(),
		"--env RUNNER_TOOL_CACHE=/opt/hostedtoolcache/" + r.uuid.String(),

		// Spin up artifact server on a different port per runner instance, so tests can run in parallel
		fmt.Sprintf("--artifact-server-port=%d", artifactServerPort),
		"--artifact-server-path=/tmp/act-artifacts/" + r.uuid.String() + "/",

		// Replace the plugin-ci-workflows action with the local checkout
		"--local-repository=grafana/plugin-ci-workflows@main=" + pciwfRoot,
		"--local-repository=grafana/plugin-ci-workflows@" + releasePleaseTag + "=" + pciwfRoot,

		// Required for cloning private repos
		"--secret", "GITHUB_TOKEN=" + r.gitHubToken,

		// Mount mockdata (for mocks)
		// and unique toolcache volume, so that multiple runners don't clash when running in parallel
		"--container-options", `"-v $PWD/tests/act/mockdata:/mockdata -v act-toolcache-` + r.uuid.String() + `:/opt/hostedtoolcache/` + r.uuid.String() + `"`,
	}
	if r.ConcurrentJobs > 0 {
		args = append(args, "--concurrent-jobs", fmt.Sprint(r.ConcurrentJobs))
	}
	// Map all self-hosted runners otherwise they don't run in act.
	for _, label := range selfHostedRunnerLabels {
		args = append(args, "-P", label+"="+nektosActRunnerImage)
	}
	return args, nil
}

// Run runs the given workflow with the given event payload using act.
func (r *Runner) Run(workflow workflow.Workflow, eventPayload EventPayload) (*RunResult, error) {
	runResult := newRunResult()

	// Create temp workflow file inside .github/workflows or act won't
	// map the repo to the workflow correctly.
	workflowFile, err := CreateTempWorkflowFile(workflow)
	if err != nil {
		return nil, fmt.Errorf("create temp workflow file: %w", err)
	}
	// TODO: enable again and also remove child workflows
	// defer os.Remove(workflowFile)

	// Create temp event payload file to simulate a GitHub event
	payloadFile, err := CreateTempEventFile(eventPayload)
	if err != nil {
		return nil, fmt.Errorf("create temp event file: %w", err)
	}
	defer os.Remove(payloadFile)

	args, err := r.args(workflowFile, payloadFile)
	if err != nil {
		return nil, fmt.Errorf("get act args: %w", err)
	}

	// TODO: escape args to avoid shell injection
	actCmd := "act " + strings.Join(args, " ")

	// Use a shell otherwise git will not be able to clone anything,
	// not even publis repositories like actions/checkout for some reason.
	cmd := exec.Command("sh", "-c", actCmd)
	cmd.Env = os.Environ()

	// Get stdout pipe to parse act output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("get act stdout pipe: %w", err)
	}
	defer stdout.Close()

	// Just pipe stderr as nothing to parse there
	cmd.Stderr = os.Stderr

	// Run act in the background
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start act: %w", err)
	}

	// Process json logs in stdout stream
	errs := make(chan error, 1)
	go func() {
		if err := r.processStream(stdout, &runResult); err != nil {
			errs <- fmt.Errorf("process act stdout: %w", err)
		}
		errs <- nil
	}()

	// Wait for act to finish
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			runResult.Success = false
			return &runResult, nil
		}
		return nil, fmt.Errorf("act exit: %w", err)
	}
	runResult.Success = true

	// Wait for stdout processing to complete
	return &runResult, <-errs
}

// processStream processes the given reader line by line as JSON log lines generated by act.
func (r *Runner) processStream(reader io.Reader, runResult *RunResult) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var data logLine
		line := scanner.Bytes()
		err := json.Unmarshal(line, &data)
		if r.Verbose {
			fmt.Println(string(line))
		}
		if err != nil {
			continue
		}

		// Print back to stdout in a human-readable format for now
		// Clean up uuids from data.Job for cleaner output
		data.Job = logUUIDRegex.ReplaceAllString(data.Job, "")
		fmt.Printf("%s: [%s] %s\n", r.t.Name(), data.Job, strings.TrimSpace(data.Message))

		// Parse GHA commands (outputs, annotations, etc.)
		r.parseGHACommand(data, runResult)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}
	return nil
}

// parseGHACommand parses intercepted GHA commands from act log lines and
// updates the RunResult accordingly. If the log line does not contain a
// recognized command, it is ignored.
func (r *Runner) parseGHACommand(data logLine, runResult *RunResult) {
	switch data.Command {
	case "set-output":
		if data.Name == "" {
			fmt.Printf("%s: [%s]: WARNING: received GHA set-output command without name, ignoring output", r.t.Name(), data.Job)
			break
		}
		// Store the output value. StepID can be an array in case of composite actions,
		// group all composite action outputs under the first step ID for simplicity.
		runResult.Outputs.Set(data.JobID, data.StepID[0], data.Name, data.Arg)
	default:
		// Nothing special to do, ignore silently
		break
	}
}

// checkExecutable checks if the given executable is available in PATH.
func checkExecutable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// checkExecutables checks if all executables required by the Runner are available in PATH.
func (r *Runner) checkExecutables() error {
	for _, name := range []string{"act", "gh"} {
		if !checkExecutable(name) {
			return fmt.Errorf("%q executable not found", name)
		}
	}
	return nil
}

// Outputs represents the outputs of jobs in a workflow run.
// Callers should use the Get and Set methods to access outputs.
type Outputs struct {
	// data is a map of job id -> step id -> output name (keys) -> output value (value)
	data map[string]map[string]map[string]string
}

// newOutputs creates a new Outputs instance.
func newOutputs() Outputs {
	return Outputs{
		data: make(map[string]map[string]map[string]string),
	}
}

// Get retrieves the output value for the given job ID, step ID, and output name.
func (o Outputs) Get(jobID, stepID, outputName string) (string, bool) {
	if steps, ok := o.data[jobID]; ok {
		if outputs, ok := steps[stepID]; ok {
			if value, ok := outputs[outputName]; ok {
				return value, true
			}
		}
	}
	return "", false
}

// Set sets the output value for the given job ID, step ID, and output name.
func (o Outputs) Set(jobID, stepID, outputName, value string) {
	steps, ok := o.data[jobID]
	if !ok {
		steps = map[string]map[string]string{}
		o.data[jobID] = steps
	}
	outputs, ok := steps[stepID]
	if !ok {
		outputs = map[string]string{}
		steps[stepID] = outputs
	}
	outputs[outputName] = value
}

type RunResult struct {
	Success bool
	Outputs Outputs
}

func newRunResult() RunResult {
	return RunResult{Outputs: newOutputs()}
}

// GetTestingWorkflowRunID retrieves the GitHub Actions workflow run ID.
// For this to work, the workflow must have been created via NewTestingWorkflow,
// which adds a job to get the run ID and expose it as an output.
func (r *RunResult) GetTestingWorkflowRunID() (string, error) {
	runID, ok := r.Outputs.Get("get-workflow-run-id", "run-id", "run-id")
	if !ok {
		return "", errors.New("could not get workflow run id. make sure you created the testing workflow via NewTestingWorkflow")
	}
	return runID, nil
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}
