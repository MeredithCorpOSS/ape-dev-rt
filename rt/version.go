package rt

import (
	"fmt"
	"os"
	"runtime"

	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/git"
	"github.com/hashicorp/terraform/terraform"
)

// The following will be filled in by the compiler
var GitCommit string
var TerraformCommit string

const TerraformVersion = terraform.Version
const Version = "0.9.3"

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

	fmt.Printf("Terraform v%s (%s)\n", TerraformVersion, TerraformCommit)
	return nil
}
