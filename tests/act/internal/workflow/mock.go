// This file contains functions to create mocked jobs and steps for testing purposes.
package workflow

// CopyMockFilesStep returns a Step that copies mock files from a source folder to a destination folder.
// The mock files are present in tests/act/mockdata in the repo, which is mounted into the act container at /mockdata.
// The sourceFolder is relative to /mockdata, e.g., "dist/simple-frontend".
// The destFolder is the destination path inside the GitHub Actions runner workspace.
// You can use GitHub Actions expressions in destFolder, e.g., "${{ github.workspace }}/plugins/my-plugin/dist".
func CopyMockFilesStep(sourceFolder string, destFolder string) Step {
	return Step{
		Name: "Copy mock dist files",
		Run: Commands{
			"set -x",
			"mkdir -p " + destFolder,
			"cp -r /mockdata/" + sourceFolder + "/. " + destFolder,
			"cd " + destFolder,
			"ls -la",
		}.String(),
		Shell: "bash",
	}
}
