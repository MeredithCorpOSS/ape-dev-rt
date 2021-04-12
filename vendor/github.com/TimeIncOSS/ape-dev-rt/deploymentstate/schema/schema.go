package schema

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/TimeIncOSS/ape-dev-rt/terraform"
)

type _SchemaVersion struct {
	Version int `json:"v"`
}

type ApplicationData struct {
	SchemaVersion int `json:"v"`

	Name string `json:"-"`

	// Allows app-per-app migration off the central repo
	UseCentralGitRepo bool `json:"use_central_git_repo"`

	IsActive             bool              `json:"is_active"`
	InfraOutputs         map[string]string `json:"infra_outputs"`
	LastRtVersion        string            `json:"last_rt_version"`
	LastTerraformVersion string            `json:"last_terraform_version"`
	LastDeploymentTime   time.Time         `json:"last_deployment_time,omitempty"`
	LastInfraChangeTime  time.Time         `json:"last_infra_change_time"`
	SlotCounters         map[string]int64  `json:"slot_counters,omitempty"`
}

func (a *ApplicationData) ToJSON() ([]byte, error) {
	a.SchemaVersion = applicationSchemaVersion
	return json.Marshal(*a)
}

func (a *ApplicationData) FromJSON(data []byte) error {
	sv := &_SchemaVersion{}
	err := json.Unmarshal(data, sv)
	if err != nil {
		return err
	}

	if sv.Version < applicationSchemaVersion {
		if sv.Version == 0 {
			migratedData, err := migrateApplication_v0_to_v1(data)
			if err != nil {
				return fmt.Errorf("Application schema migration from v0 to v1 failed: %s", err)
			}
			return a.FromJSON(migratedData)
		}
		return fmt.Errorf("No migrations available for application schema v%d", sv.Version)
	}

	if sv.Version > applicationSchemaVersion {
		return fmt.Errorf("Failed to process application data (schema v%d). "+
			"Please upgrade RT.", sv.Version)
	}

	// Unmarshall latest schema (after all migrations)
	return json.Unmarshal(data, a)
}

// 1 slot ~= 1 app deployment tfstate file
type SlotData struct {
	SchemaVersion int `json:"v"`

	SlotId   string `json:"-"`
	IsActive bool   `json:"is_active"`

	LastDeploymentStartTime time.Time     `json:"last_deployment_start_time"`
	LastDeployPilot         *DeployPilot  `json:"last_deploy_pilot,omitempty"`
	LastTerraformRun        *TerraformRun `json:"last_terraform_run"`
}

func (s *SlotData) ToJSON() ([]byte, error) {
	s.SchemaVersion = slotSchemaVersion
	return json.Marshal(*s)
}

func (s *SlotData) FromJSON(data []byte) error {
	sv := &_SchemaVersion{}
	err := json.Unmarshal(data, sv)
	if err != nil {
		return err
	}

	if sv.Version < slotSchemaVersion {
		if sv.Version == 0 {
			migratedData, err := migrateSlot_v0_to_v1(data)
			if err != nil {
				return fmt.Errorf("Slot schema migration from v0 to v1 failed: %s", err)
			}
			return s.FromJSON(migratedData)
		}
		return fmt.Errorf("No migrations available for slot schema v%d", sv.Version)
	}

	if sv.Version > slotSchemaVersion {
		return fmt.Errorf("Failed to process slot data (schema v%d). "+
			"Please upgrade RT.", sv.Version)
	}

	// Unmarshall latest schema (after all migrations)
	return json.Unmarshal(data, s)
}

type DeploymentData struct {
	SchemaVersion int    `json:"v"`
	DeploymentId  string `json:"-"`

	DeployPilot *DeployPilot `json:"deploy_pilot,omitempty"`
	StartTime   time.Time    `json:"start_time"`

	Terraform *TerraformRun `json:"terraform,omitempty"`

	RTVersion string `json:"rt_version"`

	// TODO: Data+configuration of/from hooks
	// See https://github.com/TimeIncOSS/ape-dev-rt/issues/138
	// PreDeployHooks  []*Hook
	// PostDeployHooks []*Hook
}

func (d *DeploymentData) ToJSON() ([]byte, error) {
	d.SchemaVersion = deploymentSchemaVersion
	return json.Marshal(*d)
}

func (d *DeploymentData) FromJSON(data []byte) error {
	sv := &_SchemaVersion{}
	err := json.Unmarshal(data, sv)
	if err != nil {
		return err
	}

	if sv.Version < deploymentSchemaVersion {
		if sv.Version == 0 {
			migratedData, err := migrateDeployment_v0_to_v1(data)
			if err != nil {
				return fmt.Errorf("Deployment schema migration from v0 to v1 failed: %s", err)
			}
			return d.FromJSON(migratedData)
		}
		return fmt.Errorf("No migrations available for deployment schema v%d", sv.Version)
	}

	if sv.Version > deploymentSchemaVersion {
		return fmt.Errorf("Failed to process deployment data (schema v%d). "+
			"Please upgrade RT.", sv.Version)
	}

	// Unmarshall latest schema (after all migrations)
	return json.Unmarshal(data, d)
}

type DeployPilot struct {
	AWSApiCaller string `json:"aws_api_caller"` // IAM/STS ARN
	IPAddress    string `json:"ip_address"`
}

type TerraformRun struct {
	PlanStartTime  time.Time `json:"plan_start_time,omitempty"`
	PlanFinishTime time.Time `json:"plan_finish_time,omitempty"`
	StartTime      time.Time `json:"start_time,omitempty"`
	FinishTime     time.Time `json:"finish_time"`
	IsDestroy      bool      `json:"is_destroy"`

	ResourceDiff *terraform.ResourceDiff `json:"resource_diff,omitempty"`
	Variables    map[string]string       `json:"variables"`
	Outputs      map[string]string       `json:"outputs"`

	TerraformVersion string `json:"terraform_version"`

	ExitCode int      `json:"exit_code,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Stderr   string   `json:"stderr,omitempty"`
}

type FinishedTerraformRun struct {
	PlanStartTime  time.Time `json:"plan_start_time"`
	PlanFinishTime time.Time `json:"plan_finish_time"`
	StartTime      time.Time `json:"start_time"`
	FinishTime     time.Time `json:"finish_time"`

	ResourceDiff *terraform.ResourceDiff `json:"resource_diff,omitempty"`
	Outputs      map[string]string       `json:"outputs"`

	ExitCode int      `json:"exit_code,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Stderr   string   `json:"stderr,omitempty"`
}
