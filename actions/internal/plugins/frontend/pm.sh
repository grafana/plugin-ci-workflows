#!/bin/bash

# Check if command argument is provided
if [ "$1" = "" ]; then
	echo "Please provide a command to run."
	exit 1
fi

install_pnpm_if_not_present() {
    if ! command -v pnpm &> /dev/null
    then
        echo "pnpm could not be found, installing..."
        npm install -g pnpm
    fi
}

# Use provided package manager if set in PACKAGE_MANAGER environment variable
if [ -n "$PACKAGE_MANAGER" ]; then
	pm="$PACKAGE_MANAGER"
# Detect the package manager
elif [ -f yarn.lock ]; then
	pm="yarn"
elif [ -f pnpm-lock.yaml ]; then
	install_pnpm_if_not_present
	pm="pnpm"
elif [ -f package-lock.json ]; then
	pm="npm"
else
	echo "No recognized package manager found in this project."
	exit 1
fi

# Run the provided command with the detected package manager
echo "Running '$1' with $pm..."
if [ "$1" = "install" ]; then
	# Use CI-appropriate commands that fail if lock files are out of sync
	case "$pm" in
		npm)
			npm ci
			;;
		pnpm)
			pnpm install --frozen-lockfile
			;;
		yarn)
			yarn install --frozen-lockfile
			;;
		*)
			"$pm" install
			;;
	esac
else
	"$pm" run "$1"
fi
