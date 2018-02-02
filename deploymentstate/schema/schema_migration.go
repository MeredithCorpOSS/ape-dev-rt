package schema

import (
	"encoding/json"
	"log"
	"time"
)

const (
	applicationSchemaVersion = 1
	slotSchemaVersion        = 1
	deploymentSchemaVersion  = 1
)

type ApplicationData_v0 struct {
	SchemaVersion        int               `json:"v"`
	Name                 string            `json:"-"`
	UseCentralGitRepo    bool              `json:"use_central_git_repo"`
	IsActive             bool              `json:"is_active"`
	InfraOutputs         map[string]string `json:"infra_outputs"`
	LastRtVersion        string            `json:"last_rt_version"`
	LastTerraformVersion string            `json:"last_terraform_version"`
	LastDeploymentTime   time.Time         `json:"last_deployment_time,omitempty"`
	LastInfraChangeTime  time.Time         `json:"last_infra_change_time"`
	SlotCounters         map[string]int64  `json:"slot_counters,omitempty"`
}

// Just an example of migration (no-op, just bumps the version)
func migrateApplication_v0_to_v1(sourceData []byte) ([]byte, error) {
	source := ApplicationData_v0{}
	err := json.Unmarshal(sourceData, &source)
	if err != nil {
		return nil, err
	}

	// Any potential modifications would go here
	source.SchemaVersion = 1

	log.Println("Migrated application from v0 to v1.")

	return json.Marshal(&source)
}

type SlotData_v0 struct {
	SchemaVersion           int           `json:"v"`
	SlotId                  string        `json:"-"`
	IsActive                bool          `json:"is_active"`
	LastDeploymentStartTime time.Time     `json:"last_deployment_start_time"`
	LastDeployPilot         *DeployPilot  `json:"last_deploy_pilot,omitempty"`
	LastTerraformRun        *TerraformRun `json:"last_terraform_run"`
}

// Just an example of migration (no-op, just bumps the version)
func migrateSlot_v0_to_v1(sourceData []byte) ([]byte, error) {
	source := SlotData_v0{}
	err := json.Unmarshal(sourceData, &source)
	if err != nil {
		return nil, err
	}

	// Any potential modifications would go here
	source.SchemaVersion = 1

	return json.Marshal(&source)
}

type DeploymentData_v0 struct {
	SchemaVersion int           `json:"v"`
	DeploymentId  string        `json:"-"`
	DeployPilot   *DeployPilot  `json:"deploy_pilot,omitempty"`
	StartTime     time.Time     `json:"start_time"`
	Terraform     *TerraformRun `json:"terraform,omitempty"`
	RTVersion     string        `json:"rt_version"`
}

// Just an example of migration (no-op, just bumps the version)
func migrateDeployment_v0_to_v1(sourceData []byte) ([]byte, error) {
	source := DeploymentData_v0{}
	err := json.Unmarshal(sourceData, &source)
	if err != nil {
		return nil, err
	}

	// Any potential modifications would go here
	source.SchemaVersion = 1

	return json.Marshal(&source)
}
