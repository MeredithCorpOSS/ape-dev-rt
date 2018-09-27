package terraform

import (
	"fmt"
	"github.com/TimeIncOSS/ape-dev-rt/ui"
	"os"
	"os/exec"
	"strings"
)

type Meta struct {
	Color bool           // True if output should be colored
	Ui    *ui.StreamedUi // Ui for output
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
	_, doneChan := c.Meta.Ui.AttachOutputReadCloser(stdoutPipe)
	cmd.Stdout = c.Meta.Ui.OutputWriter

	stderrPipe, _ := cmd.StderrPipe()
	c.Meta.Ui.AttachErrorReadCloser(stderrPipe)
	cmd.Stderr = c.Meta.Ui.ErrorWriter

	c.Meta.Ui.FlushBuffers()
	cmd.Start()

	err = cmd.Wait()
	if err != nil {
		return 1
	}
	<-doneChan
	return 0
}

type ApplyCommand struct {
	TfCommand
}

func (c *ApplyCommand) Run(args []string) int {
	args = append([]string{"apply"}, args...)
	return c.Execute(args)
}

type PlanCommand struct {
	TfCommand
}

func (c *PlanCommand) Run(args []string) int {
	args = append([]string{"plan"}, args...)
	return c.Execute(args)
}

type InitCommand struct {
	TfCommand
}

func (c *InitCommand) Run(args []string) int {
	args = append([]string{"init"}, args...)
	return c.Execute(args)
}

type GetCommand struct {
	TfCommand
}

func (c *GetCommand) Run(args []string) int {
	args = append([]string{"get"}, args...)
	return c.Execute(args)
}

type OutputCommand struct {
	TfCommand
}

func (c *OutputCommand) Run(args []string) int {
	args = append([]string{"output"}, args...)
	return c.Execute(args)
}

type ShowCommand struct {
	TfCommand
}

func (c *ShowCommand) Run(args []string) int {
	args = append([]string{"show"}, args...)
	return c.Execute(args)
}

type DestroyCommand struct {
	TfCommand
}

func (c *DestroyCommand) Run(args []string) int {
	args = append([]string{"destroy"}, args...)
	return c.Execute(args)
}

type TaintCommand struct {
	TfCommand
}

func (c *TaintCommand) Run(args []string) int {
	args = append([]string{"taint"}, args...)
	return c.Execute(args)
}

type UntaintCommand struct {
	TfCommand
}

func (c *UntaintCommand) Run(args []string) int {
	args = append([]string{"untaint"}, args...)
	return c.Execute(args)
}

type ValidateCommand struct {
	TfCommand
}

func (c *ValidateCommand) Run(args []string) int {
	args = append([]string{"validate"}, args...)
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
