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
// It doesn't build the plugin, it just prints a debug message that we can assert on
// in our test to confirm that the custom target was invoked.
func BuildCustom() {
	fmt.Printf("::act-debug::msg=%q\n", "custom build target invoked")
}
