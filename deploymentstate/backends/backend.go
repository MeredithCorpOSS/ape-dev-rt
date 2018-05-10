package backends

import (
	"fmt"

	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate/schema"
)

type BackendFactory struct {
	Name    string
	Backend Backend
	Meta    interface{}
}

func (b *BackendFactory) Initialize(config map[string]interface{}) error {
	meta, err := b.Backend.Configure(config)
	if err != nil {
		return fmt.Errorf("Unable to initalize backend with config: %q: %s",
			config, err)
	}

	b.Meta = meta

	return nil
}

type SlotNotFound struct {
	SlotName    string
	OriginalErr error
}

func (s *SlotNotFound) Error() string {
	return fmt.Sprintf("Slot %q was not found.", s.SlotName)
}

type AppNotFound struct {
	AppName     string
	OriginalErr error
}

func (s *AppNotFound) Error() string {
	return fmt.Sprintf("Application %q was not found.", s.AppName)
}

type Backend interface {
	Configure(config map[string]interface{}) (interface{}, error)

	SupportsWriteLock() bool

	// IsReady can perform any kind of preliminar
	// check (e.g. is TCP port open, are credentials valid,
	// are permissions sufficient) to verify the backend
	// is ready to persist our data
	IsReady(meta interface{}) (bool, error)

	// ListApplications returns all existing applications
	// (no matter if they have any slots/deployments or not)
	ListApplications(meta interface{}) ([]*schema.ApplicationData, error)

	// GetApplication returns application metadata for a given app name
	GetApplication(meta interface{}, name string) (*schema.ApplicationData, error)

	// SaveApplication saves any given application data
	SaveApplication(meta interface{}, name string, data *schema.ApplicationData) error

	// ListSlots returns all existing slots (active+inactive)
	ListSlots(meta interface{}, appName string) ([]*schema.SlotData, error)

	// DeleteSlot deletes all slot data available in the backend
	DeleteSlot(meta interface{}, appName, slotId string) error

	// SaveSlot saves any given slot data into backend
	SaveSlot(meta interface{}, appName, slotId string, slot *schema.SlotData) error

	// GetSlot returns any existing data for a given slotId
	GetSlot(meta interface{}, appName, slotId string) (*schema.SlotData, error)

	// ListSortedDeploymentsForSlotId returns n last deployments
	ListSortedDeploymentsForSlotId(meta interface{}, appName, slotId string, limitPerSlot int) ([]*schema.DeploymentData, error)

	// SaveDeployment saves any given deployment data into backend
	SaveDeployment(meta interface{}, appName, slotId, deploymentId string, data *schema.DeploymentData) error

	// GetDeployment returns deployment data
	// for a given slotId & deploymentId saved previously in the backend
	GetDeployment(meta interface{}, appName, slotId, deploymentId string) (*schema.DeploymentData, error)
}
