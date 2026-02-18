package act

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/go-logfmt/logfmt"
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
	t *testing.T

	// uuid is a unique identifier for this Runner instance.
	uuid uuid.UUID

	// actionsCachePath is the absolute path where GitHub Actions are cached.
	// If empty, a new temporary directory is created for each runner.
	actionsCachePath string

	// ArtifactsStorage is the storage for artifacts uploaded during the workflow run.
	ArtifactsStorage ArtifactsStorage

	// GCS is the GCS mock storage used during the workflow run.
	GCS GCS

	// GCOM is the GCOM API mock used during the workflow run.
	GCOM *GCOM

	// Argo is the Argo mock used during the workflow run.
	// It records inputs from the mocked Argo Workflow trigger step.
	Argo *HTTPSpy

	// gitHubToken is the token used to authenticate with GitHub.
	gitHubToken string

	// ConcurrentJobs defines the number of jobs to run concurrently via act.
	// By default (0), act uses the number of CPU cores.
	ConcurrentJobs int

	// Verbose enables logging of JSON output from act back to stdout.
	Verbose bool

	// ContainerArchitecture is the architecture to use for act containers.
	// By default, act uses the architecture of the host machine.
	// This can be useful to force a specific platform when running on ARM Macs.
	ContainerArchitecture string
}

// actionsCachePathBase is the base path for the action cache.
var actionsCachePathBase = filepath.Join("/tmp", "act-actions-cache")

// TemplateActionsCachePath is the path where GitHub Actions are pre-cached during the warmup workflow.
// This cache is then copied to each runner's actions cache path for the actual workflow runs.
var TemplateActionsCachePath = filepath.Join(actionsCachePathBase, "template")

// RunnerOption is a function that configures a Runner.
type RunnerOption func(r *Runner)

// WithVerbose enables or disables verbose logging of act output.
func WithVerbose(verbose bool) RunnerOption {
	return func(r *Runner) {
		r.Verbose = verbose
	}
}

// WithContainerArchitecture sets the container architecture to use for act.
func WithContainerArchitecture(architecture string) RunnerOption {
	return func(r *Runner) {
		r.ContainerArchitecture = architecture
	}
}

// WithLinuxAMD64ContainerArchitecture sets the container architecture to linux/amd64.
// This is useful when running on ARM Macs to ensure compatibility with x64 images.
func WithLinuxAMD64ContainerArchitecture() RunnerOption {
	return WithContainerArchitecture("linux/amd64")
}

// WithActionsCachePath sets the actions cache path for the runner (absolute path).
func WithActionsCachePath(cachePath string) RunnerOption {
	return func(r *Runner) {
		r.actionsCachePath = cachePath
	}
}

// NewRunner creates a new Runner instance.
func NewRunner(t *testing.T, opts ...RunnerOption) (*Runner, error) {
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
		GCOM:        newGCOM(t),
		Argo: NewHTTPSpy(t, map[string]string{
			"uri": "https://mock-argo-workflows.example.com/workflows/grafana-plugins-cd/mock-workflow-id",
		}),
	}
	if err := r.checkExecutables(); err != nil {
		return nil, err
	}
	r.ArtifactsStorage = newArtifactsStorage(r)
	var err error
	r.GCS, err = newGCS(r)
	if err != nil {
		return nil, fmt.Errorf("new gcs: %w", err)
	}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}
	// Default to a new temporary directory for the actions cache if not set.
	if r.actionsCachePath == "" {
		WithActionsCachePath(filepath.Join(actionsCachePathBase, r.uuid.String()))(r)
	}
	return r, nil
}

// args returns the CLI arguments to pass to act for the given workflow and event payload files.
func (r *Runner) args(eventKind EventKind, actor string, workflowFile string, payloadFile string) ([]string, error) {
	// Get a unique free port for the act artifact server, so multiple act instances can run in parallel
	artifactServerPort, err := getFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port for artifact server: %w", err)
	}

	args := []string{
		// Positional args: event kind
		string(eventKind),

		// Flags
		"-W", workflowFile,
		"-e", payloadFile,
		"--rm",
		"--json",
		// Unique artifact server port and path per act runner instance
		fmt.Sprintf("--artifact-server-port=%d", artifactServerPort),
		"--artifact-server-path=/tmp/act-artifacts/" + r.uuid.String() + "/",

		// Required for cloning private repos
		"--secret", "GITHUB_TOKEN=" + r.gitHubToken,

		// Additional Docker flags
		"--container-options", r.containerOptions(),
	}
	if r.actionsCachePath != "" {
		// Create and use per-runner cache.
		// Do not pre-populate the cache if we are using the shared cache (cache warmup).
		if r.actionsCachePath != TemplateActionsCachePath {
			if err := copyDir(TemplateActionsCachePath, r.actionsCachePath); err != nil {
				return nil, fmt.Errorf("copy action cache: %w", err)
			}
		}
		args = append(args, "--action-cache-path", r.actionsCachePath)
	} else {
		// Use shared cache
		args = append(args, "--action-cache-path", TemplateActionsCachePath)
	}

	// Map local all possible references of plugin-ci-workflows to the local repository
	localRepoArgs, err := r.localRepositoryArgs()
	if err != nil {
		return nil, err
	}
	args = append(args, localRepoArgs...)
	if actor != "" {
		args = append(args, "--actor", actor)
	}
	if r.ConcurrentJobs > 0 {
		args = append(args, "--concurrent-jobs", fmt.Sprint(r.ConcurrentJobs))
	}
	if r.ContainerArchitecture != "" {
		args = append(args, "--container-architecture", r.ContainerArchitecture)
	}
	// Map all self-hosted runners otherwise they don't run in act.
	for _, label := range selfHostedRunnerLabels {
		args = append(args, "-P", label+"="+nektosActRunnerImage)
	}
	return args, nil
}

// localRepositoryArgs returns act CLI arguments to map local references of plugin-ci-workflows
// to the local repository based on release-please configuration and manifest.
// It adds a CLI flag for each release-please component and the main branch.
func (r *Runner) localRepositoryArgs() (args []string, err error) {
	// Get local repository path
	pciwfRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	// Read release-please config: this contains the tags's prefixes
	var releasePleaseConfig struct {
		Packages map[string]struct {
			PackageName string `json:"package-name"`
		} `json:"packages"`
	}
	cfgF, err := os.Open("release-please-config.json")
	if err != nil {
		return nil, fmt.Errorf("open release-please-config.json: %w", err)
	}
	defer func() {
		if closeErr := cfgF.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if err := json.NewDecoder(cfgF).Decode(&releasePleaseConfig); err != nil {
		return nil, fmt.Errorf("decode release-please-config.json: %w", err)
	}

	// Read release-please manifest: this contains the actual semver versions (suffixes)
	manifestF, err := os.Open(".release-please-manifest.json")
	if err != nil {
		return nil, fmt.Errorf("open .release-please-manifest.json: %w", err)
	}
	defer func() {
		if closeErr := manifestF.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	var releasePleaseManifest map[string]string
	if err := json.NewDecoder(manifestF).Decode(&releasePleaseManifest); err != nil {
		return nil, fmt.Errorf("decode release-please-config.json: %w", err)
	}

	// For each component in the manifest, map its tag to the local repository
	for componentName, tag := range releasePleaseManifest {
		releasePleasePackage, ok := releasePleaseConfig.Packages[componentName]
		if !ok {
			continue
		}
		tag := releasePleasePackage.PackageName + "/v" + tag
		args = append(args, "--local-repository=grafana/plugin-ci-workflows@"+tag+"="+pciwfRoot)
	}
	args = append(args, "--local-repository=grafana/plugin-ci-workflows@main="+pciwfRoot)
	return args, nil
}

// Run runs the given workflow with the given event payload using act.
func (r *Runner) Run(workflow workflow.Workflow, event Event) (runResult *RunResult, err error) {
	result := newRunResult()
	runResult = &result

	// Create temp workflow file inside .github/workflows or act won't
	// map the repo to the workflow correctly.
	workflowFile, err := CreateTempWorkflowFile(workflow)
	if err != nil {
		return nil, fmt.Errorf("create temp workflow file: %w", err)
	}
	// TODO: enable again and also remove child workflows
	// defer os.Remove(workflowFile)

	// Create temp event payload file to simulate a GitHub event
	payloadFile, err := CreateTempEventFile(event)
	if err != nil {
		return nil, fmt.Errorf("create temp event file: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(payloadFile); removeErr != nil && err == nil {
			err = removeErr
		}
	}()

	args, err := r.args(event.Kind, event.Actor, workflowFile, payloadFile)
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

	// Just pipe stderr as nothing to parse there
	cmd.Stderr = os.Stderr

	// Run act in the background
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start act: %w", err)
	}

	// Process json logs in stdout stream.
	// This must complete BEFORE cmd.Wait() is called, because cmd.Wait()
	// closes the stdout pipe. If the goroutine is still reading when the
	// pipe is closed, it will get a "file already closed" error.
	errs := make(chan error, 1)
	go func() {
		if err := r.processStream(stdout, runResult); err != nil {
			errs <- fmt.Errorf("process act stdout: %w", err)
		}
		errs <- nil
	}()

	// Wait for stdout processing to complete FIRST.
	// The scanner will return EOF when the process exits and closes its stdout.
	if streamErr := <-errs; streamErr != nil {
		// Still call Wait to clean up the process
		_ = cmd.Wait()
		return nil, streamErr
	}

	// Now wait for the process to fully exit and get its exit status.
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			runResult.Success = false
			return runResult, nil
		}
		return nil, fmt.Errorf("act exit: %w", err)
	}
	runResult.Success = true

	return runResult, nil
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
	// Intercept custom "act-debug" command and treat it as a normal debug annotation.
	// Normally, debug annotations are very verbose and not shown in act output unless --verbose is provided.
	// We use this custom "act" command to selectively log debug information in our workflows whenever we need to,
	// and then we assert on these annotations.
	if data.Command == "act-debug" {
		data.Command = "debug"
	}

	switch data.Command {
	case "set-output":
		if data.Name == "" {
			fmt.Printf("%s: [%s]: WARNING: received GHA set-output command without name, ignoring output", r.t.Name(), data.Job)
			break
		}
		// Store the output value. StepID can be an array in case of composite actions,
		// group all composite action outputs under the first step ID for simplicity.
		runResult.Outputs.Set(data.JobID, data.StepID[0], data.Name, data.Arg)
	case "debug", "notice", "warning", "error":
		// Annotations
		runResult.Annotations = append(runResult.Annotations, Annotation{
			Level:   AnnotationLevel(data.Command),
			Title:   data.KvPairs["title"],
			Message: data.Arg,
		})
	case "summary":
		// Summary
		runResult.Summary = append(runResult.Summary, data.Content)
	default:
		// Nothing special to do
		if r.Verbose && data.Command != "" {
			fmt.Printf("%s: [%s]: unhandled GHA command %q, ignoring", r.t.Name(), data.Job, data.Command)
		}
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

// RunResult represents the result of a test workflow that was run via act.
type RunResult struct {
	// Success indicates whether the workflow run was successful.
	Success bool

	// Outputs contains the outputs for each job + step of the workflow run.
	Outputs Outputs

	// Annotations contains the GitHub Actions annotations generated during the workflow run.
	Annotations []Annotation

	// Summary contains the summary of the workflow run.
	Summary []string
}

// AnnotationLevel represents the level of a GitHub Actions annotation.
type AnnotationLevel string

// Annotation levels
const (
	AnnotationLevelDebug   AnnotationLevel = "debug"
	AnnotationLevelNotice  AnnotationLevel = "notice"
	AnnotationLevelWarning AnnotationLevel = "warning"
	AnnotationLevelError   AnnotationLevel = "error"
)

// Annotation represents a single GitHub Actions annotation.
type Annotation struct {
	// Level is the level of the annotation.
	Level AnnotationLevel

	// Title is the optional title of the annotation.
	Title string

	// Message is the message of the annotation itself.
	Message string
}

// ParseLogFmtMessage parses the annotation message as a logfmt string and returns a map of key-value pairs.
func (a Annotation) ParseLogFmtMessage() map[string]string {
	decoder := logfmt.NewDecoder(strings.NewReader(a.Message))
	result := make(map[string]string)
	for decoder.ScanRecord() {
		for decoder.ScanKeyval() {
			result[string(decoder.Key())] = string(decoder.Value())
		}
	}
	return result
}

// newRunResult creates a new empty RunResult instance.
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

// containerOptions returns the Docker container options for act.
// On Linux, it adds --add-host to enable host.docker.internal (already works on Docker Desktop for macOS/Windows).
func (r *Runner) containerOptions() string {

	opts := []string{
		// mocked testdata, dist artifacts
		"-v $PWD/tests/act/mockdata:/mockdata",
		// mocked GCS
		"-v " + r.GCS.basePath + ":/gcs",
	}

	// On Linux, add --add-host for host.docker.internal (Docker Desktop handles this automatically)
	// This is needed for local mock HTTP servers (e.g.: GCOM)
	if runtime.GOOS == "linux" {
		if hostIP := getDockerHostIP(); hostIP != "" {
			opts = append([]string{"--add-host=host.docker.internal:" + hostIP}, opts...)
		}
	}

	return `"` + strings.Join(opts, " ") + `"`
}

// getDockerHostIP returns the IP address that Docker containers can use to reach the host.
// On Linux, this is typically the docker0 bridge IP (172.17.0.1) or the IP from the default route.
// Returns empty string if detection fails (the mock servers won't work, but at least act will start).
func getDockerHostIP() string {
	// Try to get the IP from the default route (most reliable method)
	cmd := exec.Command("ip", "route", "get", "1")
	output, err := cmd.Output()
	if err == nil {
		// Output looks like: "1.0.0.0 via 192.168.1.1 dev eth0 src 192.168.1.100 uid 1000"
		// We want the "src" IP
		fields := strings.Fields(string(output))
		for i, field := range fields {
			if field == "src" && i+1 < len(fields) {
				return fields[i+1]
			}
		}
	}

	// Fallback: try docker0 bridge IP (common default)
	iface, err := net.InterfaceByName("docker0")
	if err == nil {
		addrs, err := iface.Addrs()
		if err == nil && len(addrs) > 0 {
			// Get the first IPv4 address
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	// Last resort: use the common docker bridge IP
	return "172.17.0.1"
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer func() {
				if closeErr := l.Close(); closeErr != nil && err == nil {
					err = closeErr
				}
			}()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

// copyDir recursively copies a directory tree from src to dst.
// If src does not exist, it creates an empty dst directory.
func copyDir(src, dst string) error {
	// Check if source exists
	srcInfo, err := os.Stat(src)
	if os.IsNotExist(err) {
		// Source doesn't exist, just create an empty destination directory
		return os.MkdirAll(dst, 0755)
	}
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	// Walk through source directory and copy all files
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path and destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf("get dir info: %w", err)
			}
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file from src to dst, preserving permissions.
func copyFile(src, dst string) (err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy file contents: %w", err)
	}

	return nil
}
