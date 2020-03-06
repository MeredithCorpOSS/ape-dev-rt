package terraform

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/TimeIncOSS/ape-dev-rt/ui"
	m_cli "github.com/mitchellh/cli"
)

const AppName = "app"

func FreshPlan(input *FreshPlanInput) (*PlanOutput, error) {
	_, err := ReenableRemoteState(input.RemoteState, input.RootPath)
	if err != nil {
		return nil, err
	}

	err = Get(input.RootPath)
	if err != nil {
		return nil, err
	}

	return Plan(&PlanInput{
		RootPath:     input.RootPath,
		PlanFilePath: input.PlanFilePath,
		Variables:    input.Variables,
		Refresh:      input.Refresh,
		Target:       input.Target,
		Destroy:      input.Destroy,
		StderrWriter: input.StderrWriter,
		StdoutWriter: input.StdoutWriter,
		XLegacy:      input.XLegacy,
	})
}

func Plan(input *PlanInput) (*PlanOutput, error) {
	// Arguments
	var tfArguments []string
	for key, val := range input.Variables {
		_var := fmt.Sprintf("-var=%s=%s", key, val)
		tfArguments = append(tfArguments, _var)
	}
	tfArguments = append(tfArguments, "-input=false")
	tfArguments = append(tfArguments, "-module-depth=-1")
	tfArguments = append(tfArguments, fmt.Sprintf("-out=%s", input.PlanFilePath))
	tfArguments = append(tfArguments, fmt.Sprintf("-refresh=%t", input.Refresh))
	if input.Destroy {
		tfArguments = append(tfArguments, "-destroy")
	}
	if input.Target != "" {
		tfArguments = append(tfArguments, fmt.Sprintf("-target=%s", input.Target))
	}
	if input.XLegacy {
		tfArguments = append(tfArguments, "-Xlegacy-graph")
	}

	out, err := Cmd("plan", tfArguments, input.RootPath,
		input.StdoutWriter, input.StderrWriter)
	if err != nil {
		return nil, err
	}

	diff, err := parseDiffFromPlanOutput(out.Stdout)
	if err != nil {
		return nil, err
	}

	return &PlanOutput{
		ExitCode: out.ExitCode,
		Stdout:   out.Stdout,
		Stderr:   out.Stderr,
		Diff:     diff,
		Warnings: out.Warnings,
	}, nil
}

func FreshApply(input *FreshApplyInput) (*ApplyOutput, error) {
	_, err := ReenableRemoteState(input.RemoteState, input.RootPath)
	if err != nil {
		return nil, err
	}

	err = Get(input.RootPath)
	if err != nil {
		return nil, err
	}

	return Apply(&ApplyInput{
		RootPath:     input.RootPath,
		Target:       input.Target,
		Refresh:      input.Refresh,
		PlanFilePath: input.PlanFilePath,
		StderrWriter: input.StderrWriter,
		StdoutWriter: input.StdoutWriter,
		XLegacy:      input.XLegacy,
	})
}

func Apply(input *ApplyInput) (*ApplyOutput, error) {
	// Arguments
	tfArguments := []string{
		"-input=false",
		fmt.Sprintf("-refresh=%t", input.Refresh),
	}
	if input.PlanFilePath != "" {
		tfArguments = append(tfArguments, input.PlanFilePath)
	}
	for key, val := range input.Variables {
		_var := fmt.Sprintf("-var=%s=%s", key, val)
		tfArguments = append(tfArguments, _var)
	}

	out, err := Cmd("apply", tfArguments, input.RootPath,
		input.StdoutWriter, input.StderrWriter)
	if err != nil {
		return nil, err
	}

	diff, err := parseDiffFromApplyOutput(out.Stdout)
	if err != nil {
		return nil, err
	}

	outputs := parseOutputsFromApplyOutput(out.Stdout)

	return &ApplyOutput{
		ExitCode: out.ExitCode,
		Stdout:   out.Stdout,
		Stderr:   out.Stderr,
		Warnings: out.Warnings,
		Diff:     diff,
		Outputs:  outputs,
	}, nil
}

func FreshDestroy(input *FreshDestroyInput) (*DestroyOutput, error) {
	_, err := ReenableRemoteState(input.RemoteState, input.RootPath)
	if err != nil {
		return nil, err
	}

	err = Get(input.RootPath)
	if err != nil {
		return nil, err
	}

	return Destroy(&DestroyInput{
		RootPath:     input.RootPath,
		Refresh:      input.Refresh,
		Target:       input.Target,
		Variables:    input.Variables,
		StderrWriter: input.StderrWriter,
		StdoutWriter: input.StdoutWriter,
		XLegacy:      input.XLegacy,
	})
}

func Destroy(input *DestroyInput) (*DestroyOutput, error) {
	// Arguments
	tfArguments := []string{
		"-input=false",
		"-force",
		fmt.Sprintf("-refresh=%t", input.Refresh),
	}
	for key, val := range input.Variables {
		_var := fmt.Sprintf("-var=%s=%s", key, val)
		tfArguments = append(tfArguments, _var)
	}
	if input.Target != "" {
		tfArguments = append(tfArguments, fmt.Sprintf("-target=%s", input.Target))
	}
	if input.XLegacy {
		tfArguments = append(tfArguments, "-Xlegacy-graph")
	}

	out, err := Cmd("destroy", tfArguments, input.RootPath,
		input.StdoutWriter, input.StderrWriter)
	if err != nil {
		return nil, err
	}

	diff, err := parseDiffFromApplyOutput(out.Stdout)
	if err != nil {
		return nil, err
	}

	return &DestroyOutput{
		ExitCode: out.ExitCode,
		Stdout:   out.Stdout,
		Stderr:   out.Stderr,
		Warnings: out.Warnings,
		Diff:     diff,
	}, nil
}

func IsStateEmpty(rootPath string) (bool, error) {
	// TODO: Use `terraform state list` instead to filter out false negatives due to Outputs section
	state, err := Show(rootPath)
	if err != nil {
		return false, err
	}
	if len(state) == 0 {
		return true, nil
	}
	if regexp.MustCompile("^Outputs:\n").MatchString(state) {
		return true, nil
	}
	return false, nil
}

func FreshShow(remoteState *RemoteState, rootPath string) (string, error) {
	_, err := ReenableRemoteState(remoteState, rootPath)
	if err != nil {
		return "", err
	}
	return Show(rootPath)
}

func Show(rootPath string) (string, error) {
	var args []string
	args = append(args, "-no-color")
	out, err := Cmd("show", args, rootPath, ioutil.Discard, ioutil.Discard)
	if err != nil {
		return "", err
	}
	if out.ExitCode != 0 {
		return "", fmt.Errorf("Error(s) occured (exit code %d). Stderr:\n%s",
			out.ExitCode, out.Stderr)
	}

	// Ignoring warnings here as we don't expect any from show cmd

	return strings.TrimSpace(out.Stdout), nil
}

func GetBackendConfigFilename(rootPath string) string {
	var config_file_name string
	config_file_name = "backend-config.tf.json"
	return path.Join(rootPath, config_file_name)
}

func GenerateBackendConfig(remoteState *RemoteState, rootPath string) (string, error) {
	// create backend config file from remoteState object
	// TODO: check for existing file before writing

	var s3backend = S3Backend{remoteState.Config}
	var backends []S3Backend
	backends = make([]S3Backend, 1)
	backends[0] = s3backend
	var backendObject = BackendObj{backends}

	var backendList []BackendObj
	backendList = make([]BackendObj, 1)
	backendList[0] = backendObject
	var backendConfig = BackendConfig{backendList}

	b, err := json.Marshal(backendConfig)
	if err != nil {
		return "", err
	}

	file, err := os.Create(GetBackendConfigFilename(rootPath))
	if err != nil {
		return "", fmt.Errorf("Unable to create backend config file %s", err)
	}
	defer file.Close()

	var bytes_written int
	bytes_written, err = fmt.Fprintf(file, string(b))
	if err != nil {
		return "", fmt.Errorf("Unable to write backend config into %s", err)
	}
	if bytes_written == 0 {
		return "", fmt.Errorf("Zero bytes written to backend config file")
	}
	log.Printf("[DEBUG] Wrote backend config: %s", GetBackendConfigFilename(rootPath))

	return "", nil
}

func ReenableRemoteState(remoteState *RemoteState, rootPath string) (string, error) {
	os.RemoveAll(path.Join(rootPath, ".terraform"))

	_, err := GenerateBackendConfig(remoteState, rootPath)
	if err != nil {
		return "", err
	}

	var output string

	out, err := Cmd("init", nil, rootPath, os.Stdout, os.Stderr)
	if err != nil {
		return "", err
	}

	if out.ExitCode != 0 {
		return "", fmt.Errorf("Error(s) occured when initialising backend (exit code %d). Stderr:\n%s",
			out.ExitCode, out.Stderr)
	}
	output += out.Stdout

	return output, nil
}

func Get(rootPath string) error {
	out, err := Cmd("get", []string{"-update"}, rootPath, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	if out.ExitCode != 0 {
		return fmt.Errorf("Error(s) occured when updating modules (exit code %d). Stderr:\n%s",
			out.ExitCode, out.Stderr)
	}
	return nil
}

func FreshOutput(remoteState *RemoteState, rootPath string) (map[string]string, error) {
	_, err := ReenableRemoteState(remoteState, rootPath)
	if err != nil {
		return nil, err
	}

	return Output(rootPath)
}

func Output(rootPath string) (map[string]string, error) {
	args := []string{
		"-no-color",
		// TODO: Terraform v0.7+
		// "-json",
	}

	out, err := Cmd("output", args, rootPath, os.Stdout, os.Stderr)
	if err != nil {
		return nil, err
	}

	if out.ExitCode != 0 {
		if out.Stdout == "The module root could not be found. There is nothing to output.\n" {
			return nil, nil
		}
		if strings.HasPrefix(out.Stdout, "The state file has no outputs defined.") {
			return nil, nil
		}

		return nil, fmt.Errorf("Error(s) occured with output (exit code %d). Stderr:\n%s",
			out.ExitCode, out.Stderr)
	}

	var outputs = make(map[string]string, 0)
	// TODO: Parse JSON output instead (Terraform v0.7)
	lines := strings.Split(out.Stdout, "\n")
	for _, l := range lines {
		items := strings.Split(l, " = ")
		if len(items) == 2 {
			key, value := items[0], items[1]
			outputs[key] = value
		}
	}

	return outputs, nil

}

func Validate(rootpath string) (*CmdOutput, error) {
	out, err := Cmd("validate", []string{}, rootpath, os.Stdout, os.Stderr)
	if err != nil {
		return nil, err
	}

	if out.ExitCode != 0 {
		return nil, fmt.Errorf("Error(s) occured (exit code %d). Stderr:\n%s",
			out.ExitCode, out.Stderr)
	}
	return out, nil

}

func Cmd(cmdName string, args []string, basePath string, stdoutW, stderrW io.Writer) (*CmdOutput, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	os.Chdir(basePath)
	defer os.Chdir(workDir)

	streamedUi := new(ui.StreamedUi)
	streamedUi.OutputWriter = stdoutW
	streamedUi.ErrorWriter = stderrW

	meta := Meta{
		Ui:    streamedUi,
		Color: true,
	}

	commands := map[string]m_cli.Command{
		"apply": &ApplyCommand{
			TfCommand{Meta: meta}},
		"get": &GetCommand{
			TfCommand{Meta: meta}},
		"output": &OutputCommand{
			TfCommand{Meta: meta}},
		"plan": &PlanCommand{
			TfCommand{Meta: meta}},
		"init": &InitCommand{
			TfCommand{Meta: meta}},
		"show": &ShowCommand{
			TfCommand{Meta: meta}},
		"destroy": &DestroyCommand{
			TfCommand{Meta: meta}},
		"taint": &TaintCommand{
			TfCommand{Meta: meta}},
		"untaint": &UntaintCommand{
			TfCommand{Meta: meta}},
		"validate": &ValidateCommand{
			TfCommand{Meta: meta}},
	}

	cmd, ok := commands[cmdName]
	if !ok {
		return nil, fmt.Errorf("Unknown Terraform command: %s", cmdName)
	}

	log.Printf("[DEBUG] Executing: terraform %s %q in path %s", cmdName, args, basePath)
	exitCode := cmd.Run(args)
	bufferedStdout := streamedUi.OutputBuffer.String()
	bufferedStderr := streamedUi.ErrorBuffer.String()
	streamedUi.FlushBuffers()
	streamedUi.OutputBuffer.Reset()
	streamedUi.ErrorBuffer.Reset()

	warns := parseOutWarnings(bufferedStdout)

	return &CmdOutput{
		Stdout:   bufferedStdout,
		Stderr:   bufferedStderr,
		ExitCode: exitCode,
		Warnings: warns,
	}, nil
}

func parseDiffFromPlanOutput(output string) (planDiff *PlanResourceDiff, err error) {
	sumsRegexp := regexp.MustCompile(
		"Plan:[^ ]{0,6} " + // Coloring
			"([0-9]+) to add, ([0-9]+) to change, ([0-9]+) to destroy.")
	matches := sumsRegexp.FindStringSubmatch(output)

	if matches == nil {
		return &PlanResourceDiff{
			ToCreate: 0,
			ToChange: 0,
			ToRemove: 0,
		}, nil
	}

	toCreate, err := strconv.ParseInt(matches[1], 0, 0)
	if err != nil {
		return
	}
	toChange, err := strconv.ParseInt(matches[2], 0, 0)
	if err != nil {
		return
	}
	toRemove, err := strconv.ParseInt(matches[3], 0, 0)
	if err != nil {
		return
	}

	return &PlanResourceDiff{
		ToCreate: int(toCreate),
		ToChange: int(toChange),
		ToRemove: int(toRemove),
	}, nil
}

func parseDiffFromApplyOutput(output string) (*ResourceDiff, error) {
	re := regexp.MustCompile("(?:Apply|Destroy) complete! Resources: " +
		"(([0-9]+) added, )?(([0-9]+) changed, )?([0-9]+) destroyed.")
	matches := re.FindStringSubmatch(output)

	if matches == nil {
		return &ResourceDiff{
			Created: 0,
			Changed: 0,
			Removed: 0,
		}, nil
	}

	created, err := strconv.ParseInt(matches[2], 0, 0)
	if err != nil {
		created = 0
	}
	changed, err := strconv.ParseInt(matches[4], 0, 0)
	if err != nil {
		changed = 0
	}
	removed, err := strconv.ParseInt(matches[5], 0, 0)
	if err != nil {
		return nil, err
	}

	return &ResourceDiff{
		Created: int(created),
		Changed: int(changed),
		Removed: int(removed),
	}, nil
}

func parseOutWarnings(input string) []string {
	var warnings = make([]string, 0)
	re := regexp.MustCompile("(?s)Warnings:\n\x1b\\[0m\x1b\\[0m\n(.+)\n\n")
	matches := re.FindStringSubmatch(input)

	if len(matches) == 2 {
		prefix := "\x1b[33m  * "
		suffix := "\x1b[0m\x1b[0m"
		warnings = strings.Split(matches[1], "\n")
		for i, w := range warnings {
			warnings[i] = strings.TrimPrefix(w, prefix)
			warnings[i] = strings.TrimSuffix(warnings[i], suffix)
		}
	}

	return warnings
}

func parseOutputsFromApplyOutput(input string) map[string]string {
	outputs := make(map[string]string, 0)
	re := regexp.MustCompile("(?s)Outputs:\n\n(.+)\\[0m\n")
	matches := re.FindStringSubmatch(input)
	if len(matches) == 2 {
		lines := strings.Split(matches[1], "\n")
		for _, l := range lines {
			m := strings.Split(l, " = ")
			if len(m) != 2 {
				// Skip anything that's not parseable (map, list)
				continue
			}
			trimmedKey := strings.TrimSpace(m[0])
			if m[0] != trimmedKey {
				// Skip elements of a map (those are indented)
				continue
			}
			v := strings.TrimSpace(m[1])
			if v == "{" || v == "[" {
				// Skip anything that looks like a map or list
				log.Printf("[WARN] Parser is skipping %q because it looks like a list or map", m[0])
				continue
			}

			key, value := trimmedKey, v
			outputs[key] = value
		}
	}
	return outputs
}

func GetRemoteStateForApp(rs *RemoteState, namespace, appName string) (*RemoteState, error) {
	if rs.Backend == "s3" {
		rs.Config["key"] = fmt.Sprintf("%s/%s/terraform.tfstate", namespace, appName)
		return rs, nil
	}

	return nil, fmt.Errorf("Unable to construct app remote state cfg for %q backend", rs.Backend)
}

func GetRemoteStateForSlotId(rs *RemoteState, namespace, appName, slotId string) (*RemoteState, error) {
	if rs.Backend == "s3" {
		rs.Config["key"] = fmt.Sprintf("%s/%s/slots/%s.tfstate", namespace, appName, slotId)
		return rs, nil
	}

	return nil, fmt.Errorf("Unable to construct slot remote state cfg for %q backend", rs.Backend)
}
