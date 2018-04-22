package backendstest

import (
	"errors"

	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate/schema"
)

type FixtureBackend struct{}

func (fb *FixtureBackend) Configure(config map[string]interface{}) (interface{}, error) {
	if _, ok := config["bucket"]; !ok {
		return nil, errors.New("Missing bucket field in config")
	}
	return 42, nil
}

func (fb *FixtureBackend) SupportsWriteLock() bool {
	return false
}

func (fb *FixtureBackend) IsReady(meta interface{}) (bool, error) {
	return false, nil
}

func (fb *FixtureBackend) ListApplications(meta interface{}) ([]*schema.ApplicationData, error) {
	return nil, nil
}

func (fb *FixtureBackend) SaveApplication(meta interface{}, name string, data *schema.ApplicationData) error {
	return nil
}

func (fb *FixtureBackend) GetApplication(meta interface{}, name string) (*schema.ApplicationData, error) {
	return nil, nil
}

func (fb *FixtureBackend) DeleteSlot(meta interface{}, appName, slotId string) error {
	return nil
}

func (fb *FixtureBackend) ListSlots(meta interface{}, appName string) ([]*schema.SlotData, error) {
	return nil, nil
}

func (fb *FixtureBackend) SaveSlot(meta interface{}, appName, slotId string, slot *schema.SlotData) error {
	return nil
}

func (fb *FixtureBackend) GetSlot(meta interface{}, appName, slotId string) (*schema.SlotData, error) {
	return nil, nil
}

func (fb *FixtureBackend) ListSortedDeploymentsForSlotId(meta interface{}, appName, slotId string, limitPerSlot int) ([]*schema.DeploymentData, error) {
	return nil, nil
}

func (fb *FixtureBackend) SaveDeployment(meta interface{}, appName, slotId, deploymentId string, data *schema.DeploymentData) error {
	return nil
}

func (fb *FixtureBackend) GetDeployment(meta interface{}, appName, slotId, deploymentId string) (*schema.DeploymentData, error) {
	return nil, nil
}
