//go:build mage
// +build mage

package main

import (
	"fmt"

	// mage:import
	build "github.com/grafana/grafana-plugin-sdk-go/build"
)

// Default configures the default target.
var Default = build.BuildAll

// BuildCustom is a custom build target used to test the backend-build-target CI input.
func BuildCustom() {
	fmt.Printf("::act-debug::msg=%q\n", "custom build target invoked")
	build.BuildAll()
}
