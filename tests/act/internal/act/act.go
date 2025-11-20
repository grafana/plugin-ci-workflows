package act

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
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
	_, isRunningInGHA = os.LookupEnv("CI")
)

const nektosActRunnerImage = "ghcr.io/catthehacker/ubuntu:act-latest"

type Runner struct {
	gitHubToken string
}

func NewRunner() (*Runner, error) {
	if err := checkExecutables(); err != nil {
		return nil, err
	}

	ghToken, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok || ghToken == "" {
		cmd := exec.Command("gh", "auth", "token")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("exec 'gh auth token': %w", err)
		}
		ghToken = strings.TrimSpace(string(output))
	}

	return &Runner{
		gitHubToken: ghToken,
	}, nil
}

func (r *Runner) args(workflow string) []string {
	pciwfRoot, err := os.Getwd()
	if err != nil {
		// TODO: do not fail silently
		pciwfRoot = ""
	}
	args := []string{
		"-W", workflow,
		"--rm",
		"--json",
		"--artifact-server-path=/tmp/artifacts",
		"--local-repository=grafana/plugin-ci-workflows@main=" + pciwfRoot,
		// "--no-skip-checkout",
		"--secret", "GITHUB_TOKEN=" + r.gitHubToken,
		// TODO: remove
		"--concurrent-jobs", "1",
	}
	for _, label := range selfHostedRunnerLabels {
		args = append(args, "-P", label+"="+nektosActRunnerImage)
	}
	return args
}

func (r *Runner) Run(workflow string) error {
	args := r.args(workflow)

	// TODO: escape args to avoid shell injection
	actCmd := "act " + strings.Join(args, " ")
	fmt.Println(actCmd)

	// Use a shell otherwise git will not be able to clone anything,
	// not even publis repositories like actions/checkout for some reason.
	cmd := exec.Command("sh", "-c", actCmd)
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("get act stdout pipe: %w", err)
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("get act stderr pipe: %w", err)
	}
	defer stderr.Close()

	// TODO: combine readers together

	stdoutTee := io.TeeReader(stdout, os.Stdout)
	stderrTee := io.TeeReader(stderr, os.Stderr)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start act: %w", err)
	}

	errs := make(chan error, 2)
	go func() {
		if err := r.processStream(stdoutTee); err != nil {
			errs <- fmt.Errorf("process act stdout: %w", err)
		}
		errs <- nil
	}()
	go func() {
		if err := r.processStream(stderrTee); err != nil {
			errs <- fmt.Errorf("process act stderr: %w", err)
		}
		errs <- nil
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("act exit: %w", err)
	}
	// Wait for stdout and stderr processing to complete
	var finalErr error
	for i := 0; i < 2; i++ {
		finalErr = errors.Join(finalErr, <-errs)
	}
	return finalErr
}

func (r *Runner) processStream(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var data map[string]any
		err := json.Unmarshal(scanner.Bytes(), &data)
		if err != nil {
			// Ignore silently
			continue
		}
		// TODO: parse
		// fmt.Printf("%+v\n", data)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}
	fmt.Println("scanner closed")
	return nil
}

func checkExecutable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func checkExecutables() error {
	for _, name := range []string{"act", "gh"} {
		if !checkExecutable(name) {
			return fmt.Errorf("%q executable not found", name)
		}
	}
	return nil
}
