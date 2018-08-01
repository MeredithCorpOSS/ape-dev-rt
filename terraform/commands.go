package terraform

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"github.com/TimeIncOSS/ape-dev-rt/ui"
)

type Meta struct {
	Color            bool             // True if output should be colored
	Ui               *ui.StreamedUi           // Ui for output
}

type TfCommand struct {
	Meta
}

func (c *TfCommand) Execute(args []string) int {
	binary, err := CheckTerraform()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Correct version of terraform not available\n")
		return 1
	}

	cmd := exec.Command(binary, args...)
	stdoutPipe, _ := cmd.StdoutPipe()
	stdout, doneChan := c.Meta.Ui.AttachOutputReadCloser(stdoutPipe)
	//cmd.Stdout = c.Meta.Ui.OutputWriter

	stderrPipe, _ := cmd.StderrPipe()
	c.Meta.Ui.AttachErrorReadCloser(stderrPipe)
	//cmd.Stderr = c.Meta.Ui.ErrorWriter
	var stderr *bytes.Buffer
	stderr = c.Meta.Ui.ErrorBuffer

	c.Meta.Ui.FlushBuffers()
	cmd.Start()

	err = cmd.Wait()
	if err != nil {
		//fmt.Fprintf(os.Stderr, "Terraform error with %q, %s\n\n%s\n", args, err.Error(), "stderr")
		fmt.Fprintf(os.Stderr, "Terraform error with %q, %s\n\n%s\n", args, err.Error(), stderr.String())
		return 1
	}
	<-doneChan
	fmt.Fprintf(os.Stdout, stdout.String())
	//fmt.Fprintf(os.Stdout, "hi")
	return 0
}

type ApplyCommand struct {
	TfCommand
}

func (c *ApplyCommand) Run(args []string) int {
	fmt.Fprintf(os.Stderr, "trying to run ApplyCommand: %q\n", args)
	args = append([]string{"apply"}, args...)

	return c.Execute(args)
}

type PlanCommand struct {
	TfCommand
}

func (c *PlanCommand) Run(args []string) int {
	fmt.Fprintf(os.Stderr, "trying to run PlanCommand: %q\n", args)
	args = append([]string{"plan"}, args...)

	return c.Execute(args)
}

type InitCommand struct {
	TfCommand
}

func (c *InitCommand) Run(args []string) int {
	fmt.Fprintf(os.Stderr, "trying to run InitCommand: %q\n", args)
	args = append([]string{"init"}, args...)

	return c.Execute(args)
}

// This shouldn't be seen, but is required by github.com/mitchellh/cli
func (c *TfCommand) Help() string {
	helpText := `
	This is help text
	`
	return strings.TrimSpace(helpText)
}

// This shouldn't be seen, but is required by github.com/mitchellh/cli
func (c *TfCommand) Synopsis() string {
	return "Builds or changes infrastructure"
}

func CheckTerraform() (string, error) {
	return "/Users/zshepherd1271/devwork/ape-dev-terraform/bin108/terraform", nil
}
