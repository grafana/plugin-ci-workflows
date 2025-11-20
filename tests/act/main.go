package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/grafana/plugin-ci-workflows/tests/act/internal/act"
)

func _main() error {
	runner, err := act.NewRunner()
	if err != nil {
		return err
	}
	if err := runner.Run(filepath.Join(".github", "workflows", "act.yml")); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
