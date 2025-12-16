#!/bin/bash
set -euo pipefail

# Prints all testdata plugins found in the tests/ folder to stdout, one per line.
# The "act" folder is excluded since it contains the tests themselves.
# Can be used to loop over all test plugins, like this:
#
#   for tc in $(./scripts/find-tests.sh); do
#     echo $tc
#   done

cd tests
find . -maxdepth 1 -type d ! -name '.' ! -name 'act' -print