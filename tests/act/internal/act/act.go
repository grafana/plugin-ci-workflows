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
	"path/filepath"
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
	t *testing.T

	// uuid is a unique identifier for this Runner instance.
	uuid uuid.UUID

	// ArtifactsStorage is the storage for artifacts uploaded during the workflow run.
	ArtifactsStorage ArtifactsStorage

	// GCS is the GCS mock storage used during the workflow run.
	GCS GCS

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

	return r, nil
}

// args returns the CLI arguments to pass to act for the given workflow and event payload files.
func (r *Runner) args(eventKind EventKind, actor string, workflowFile string, payloadFile string) ([]string, error) {
	// Get a unique free port for the act artifact server, so multiple act instances can run in parallel
	artifactServerPort, err := getFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port for artifact server: %w", err)
	}

	// Get the repository root so tests can run from any subdirectory
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("get repo root: %w", err)
	}
	mockdataPath := filepath.Join(repoRoot, "tests", "act", "mockdata")

	args := []string{
		// Positional args: event kind
		string(eventKind),

		// Flags
		"-W", workflowFile,
		"-e", payloadFile,
		"--rm",
		"--json",
		// Bind mount the working directory instead of copying, so working-directory paths work
		"--bind",
		// Unique artifact server port and path per act runner instance
		fmt.Sprintf("--artifact-server-port=%d", artifactServerPort),
		"--artifact-server-path=/tmp/act-artifacts/" + r.uuid.String() + "/",
		// Unique action cache path per act runner instance to prevent cache corruption
		// when running multiple tests in parallel or reusing runners with stale cache
		"--action-cache-path=/tmp/act-action-cache/" + r.uuid.String() + "/",

		// Required for cloning private repos
		"--secret", "GITHUB_TOKEN=" + r.gitHubToken,

		// Mounts:
		// - mockdata: for mocked testdata, dist artifacts
		// - GCS: for mocked GCS
		// - /tmp: for temporary files, so the host's /tmp is used
		// Note: act automatically mounts the repo root at the same absolute path
		"--container-options", `"-v ` + mockdataPath + `:/mockdata -v ` + r.GCS.basePath + `:/gcs -v /tmp:/tmp"`,
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
	pciwfRoot, err := getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("get repo root: %w", err)
	}

	// Read release-please config: this contains the tags's prefixes
	var releasePleaseConfig struct {
		Packages map[string]struct {
			PackageName string `json:"package-name"`
		} `json:"packages"`
	}
	cfgF, err := os.Open(filepath.Join(pciwfRoot, "release-please-config.json"))
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
	manifestF, err := os.Open(filepath.Join(pciwfRoot, ".release-please-manifest.json"))
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

	// Get the repository root so act runs from the correct directory.
	// This ensures the entire repository is mounted as the workspace,
	// allowing working-directory specifications like "tests/simple-backend" to work.
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("get repo root: %w", err)
	}

	// TODO: escape args to avoid shell injection
	actCmd := "act " + strings.Join(args, " ")

	// Use a shell otherwise git will not be able to clone anything,
	// not even publis repositories like actions/checkout for some reason.
	cmd := exec.Command("sh", "-c", actCmd)
	cmd.Dir = repoRoot
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

	// Process json logs in stdout stream
	errs := make(chan error, 1)
	go func() {
		if err := r.processStream(stdout, runResult); err != nil {
			errs <- fmt.Errorf("process act stdout: %w", err)
		}
		errs <- nil
	}()

	// Wait for act to finish
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			runResult.Success = false
			return runResult, nil
		}
		return nil, fmt.Errorf("act exit: %w", err)
	}
	runResult.Success = true

	// Wait for stdout processing to complete
	return runResult, <-errs
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
	case "debug", "notice", "warning", "error":
		// Annotations
		runResult.Annotations = append(runResult.Annotations, Annotation{
			Level:   AnnotationLevel(data.Command),
			Title:   data.KvPairs["title"],
			Message: data.Arg,
		})
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

// getRepoRoot returns the absolute path of the root of the git repository.
// This allows tests to run from any subdirectory of the repository.
func getRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current working directory: %w", err)
	}
	for {
		gitPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil && info.IsDir() {
			return dir, nil
		}
		if os.IsNotExist(err) {
			parentDir := filepath.Dir(dir)
			if parentDir == dir {
				break // Reached the root directory
			}
			dir = parentDir
			continue
		}
		return "", fmt.Errorf("stat .git directory: %w", err)
	}
	return "", errors.New(".git directory not found in any parent directories")
}

// actTestPluginsDir is the base directory for temporary plugin copies used during tests.
const actTestPluginsDir = "/tmp/act-test-plugins"

// CopyPluginToTemp copies a plugin source directory to a temporary folder.
// The sourceDir should be the name of a plugin folder inside tests/ (e.g., "simple-frontend").
// It returns the absolute path to the temp folder for use with workflow.WithPluginDirectoryInput().
// The temp folder is automatically cleaned up when the test finishes via t.Cleanup().
func CopyPluginToTemp(t *testing.T, sourceDir string) (string, error) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		return "", fmt.Errorf("get repo root: %w", err)
	}

	// Source directory: tests/{sourceDir}
	srcPath := filepath.Join(repoRoot, "tests", sourceDir)

	// Create unique temp directory: /tmp/act-test-plugins/{test-name}-{uuid}/
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	tempDir := filepath.Join(actTestPluginsDir, testName+"-"+uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	// Register cleanup to remove the temp directory after the test
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("warning: failed to remove temp dir %s: %v", tempDir, err)
		}
	})

	// Copy all files from source to temp directory
	if err := copyDir(srcPath, tempDir); err != nil {
		return "", fmt.Errorf("copy plugin to temp: %w", err)
	}

	return tempDir, nil
}

// skipDirs contains directory names to skip when copying plugin directories.
// These are generated directories that would be recreated during the build.
var skipDirs = map[string]struct{}{
	"node_modules":   {},
	"dist":           {},
	"dist-artifacts": {},
}

// copyDir recursively copies a directory tree from src to dst,
// skipping node_modules, dist, and dist-artifacts directories.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip certain directories entirely
		if _, skip := skipDirs[info.Name()]; info.IsDir() && skip {
			return filepath.SkipDir
		}

		// Calculate the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create the directory
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy the file
		return copyFile(path, dstPath, info.Mode())
	})
}

// copyFile copies a single file from src to dst with the given permissions.
func copyFile(src, dst string, mode os.FileMode) (err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close source file: %w", closeErr)
		}
	}()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close destination file: %w", closeErr)
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy file content: %w", err)
	}

	return nil
}
