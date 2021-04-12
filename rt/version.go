package rt

import (
	"fmt"
	"os"
	"runtime"

	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/git"
)

// GitCommit ... The following will be filled in by the compiler
var GitCommit string

// TerraformVersion ... terraform version needed to run
const TerraformVersion = "0.13.6"

// Version ... current ape-dev-rt version
const Version = "0.12.1"

// GetVersion ... Return version information for CLI
func GetVersion(c *commons.Context) error {
	fmt.Printf("rt %s (%s)\n", Version, GitCommit)
	fmt.Printf("go %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	g := git.NewGit("", "")
	version, err := g.Version()
	if err == nil {
		fmt.Print(version)
	} else {
		fmt.Fprintf(os.Stderr, "git - Unable to get version (%s)\n", err.Error())
	}

	fmt.Printf("requires Terraform v%s\n", TerraformVersion)
	return nil
}
