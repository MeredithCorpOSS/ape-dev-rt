package terraform

import (
	"io"
)

type RemoteState struct {
	Backend string
	Config  map[string]string
}

type CmdOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Warnings []string
}

type FreshPlanInput struct {
	RemoteState  *RemoteState
	StdoutWriter io.Writer
	StderrWriter io.Writer
	RootPath     string
	PlanFilePath string
	Variables    map[string]string
	Refresh      bool
	Target       string
	Destroy      bool
}

type PlanInput struct {
	StdoutWriter io.Writer
	StderrWriter io.Writer
	RootPath     string
	PlanFilePath string
	Variables    map[string]string
	Refresh      bool
	Target       string
	Destroy      bool
}

type PlanOutput struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Warnings []string
	Diff     *PlanResourceDiff
}

type PlanResourceDiff struct {
	ToCreate int
	ToRemove int
	ToChange int
}

type FreshApplyInput struct {
	RemoteState  *RemoteState
	StdoutWriter io.Writer
	StderrWriter io.Writer
	RootPath     string
	Target       string
	Refresh      bool
	PlanFilePath string
}

type ApplyInput struct {
	StdoutWriter io.Writer
	StderrWriter io.Writer
	RootPath     string
	Target       string
	Refresh      bool
	Variables    map[string]string
	PlanFilePath string
}

type ApplyOutput struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Warnings []string
	Outputs  map[string]string
	Diff     *ResourceDiff
}

type ResourceDiff struct {
	Created int
	Removed int
	Changed int
}

type FreshDestroyInput struct {
	RemoteState  *RemoteState
	StdoutWriter io.Writer
	StderrWriter io.Writer
	RootPath     string
	Refresh      bool
	Target       string
	Variables    map[string]string
}

type DestroyInput struct {
	StdoutWriter io.Writer
	StderrWriter io.Writer
	RootPath     string
	Refresh      bool
	Target       string
	Variables    map[string]string
}

type DestroyOutput struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Warnings []string
	Diff     *ResourceDiff
}
