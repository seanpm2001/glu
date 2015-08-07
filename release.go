package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/progrium/go-shell"
	"github.com/spf13/cobra"
)

func init() {
	GluCmd.AddCommand(releaseCmd)
}

var releaseCmd = &cobra.Command{
	Use:   "release <github-repo> <version> [release-name] [checksum-hash]",
	Short: "Creates a github release using gh-release",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.Usage()
			os.Exit(1)
		}
		var (
			githubRepo   = args[0]
			projectName  = path.Base(githubRepo)
			version      = args[1]
			releaseName  = optArg(args, 2, version)
			checksumHash = optArg(args, 3, "")
			arch, _      = exec.Command("uname", "-m").Output()
			branch, _    = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		)

		defer shell.ErrExit()
		shell.Trace = true
		shell.Tee = os.Stdout
		sh("go get -u github.com/progrium/gh-release/...")
		sh("rm -rf release")
		sh("mkdir release")

		for _, platform := range []string{"Linux", "Darwin"} {
			if binary := detectBinaryBuild(platform); binary != "" {
				// tar -zcf release/$(NAME)_$(VERSION)_$(PLATFORM)_$(ARCH).tgz -C build/$(PLATFORM) $(BINARYNAME)
				sh("tar -zcf",
					fmt.Sprintf("release/%s_%s_%s_%s.tgz", projectName, version, platform, arch),
					"-C build/"+platform, binary)
			}
		}

		if dir, _ := ioutil.ReadDir("release"); len(dir) == 0 {
			sh("cp build/* release")
		}

		if checksumHash != "" {
			sh("gh-release checksums", checksumHash)
		}

		sh("gh-release create", githubRepo, version, branch, releaseName)
	},
}

func detectBinaryBuild(platform string) string {
	if !exists("build", platform) {
		return ""
	}
	dir, err := ioutil.ReadDir("build/" + platform)
	fatal(err)
	if len(dir) != 1 {
		return ""
	}
	return dir[0].Name()
}