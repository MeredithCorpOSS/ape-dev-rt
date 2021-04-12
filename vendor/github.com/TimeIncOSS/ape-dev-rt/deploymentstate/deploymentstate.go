package deploymentstate

import (
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate/backends"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate/schema"
	"github.com/TimeIncOSS/ape-dev-rt/hcl"
	"github.com/TimeIncOSS/ape-dev-rt/rt"
	"github.com/hashicorp/go-multierror"
)

var supportedBackends = map[string]backends.Backend{
	"s3": &backends.S3{},
}

type DeploymentState struct {
	// backendList persists configured backends
	// ordereding matches ordering in HCL config
	backendList []*backends.BackendFactory
}

func New(config *hcl.DeploymentState) (*DeploymentState, error) {
	if config == nil {
		return nil, fmt.Errorf("No configuration provided")
	}
	ds := DeploymentState{}
	err := loadBackends(&ds, config.Iterator())
	if err != nil {
		return nil, err
	}

	return &ds, nil
}

func loadBackends(ds *DeploymentState, backendsCfg []map[string]interface{}) error {
	var loaded = false

	for _, backend := range backendsCfg {
		for backendName, backendValuesSlice := range backend {
			log.Printf("Loading backend %q\n", backendName)
			backend, err := ds.loadBackend(backendName)
			if err != nil {
				return err
			}
			log.Printf("Backend %q loaded.\n", backendName)

			if v, ok := backendValuesSlice.([]map[string]interface{}); ok {
				for _, backendValues := range v {
					err := backend.Initialize(backendValues)
					if err != nil {
						return fmt.Errorf("Error initializing backend: %q", err)
					}
					log.Printf("[DEBUG] Backend %s initialized w/ %q",
						backendName, backendValues)
					loaded = true
				}
			}
		}
	}

	if !loaded {
		return errors.New("No loadable backend found")
	}

	return nil
}

func (ds *DeploymentState) loadBackend(name string) (*backends.BackendFactory, error) {
	backend, ok := supportedBackends[name]
	if !ok {
		return nil, fmt.Errorf("Defined backend %s is not supported", name)
	}

	// Look for duplicates
	for _, b := range ds.backendList {
		if b.Name == name {
			return nil, fmt.Errorf("Duplicate backend defined (%s)", name)
		}
	}

	backendFactory := &backends.BackendFactory{
		Name:    name,
		Backend: backend,
	}
	ds.backendList = append(ds.backendList, backendFactory)

	return backendFactory, nil
}

func (ds *DeploymentState) AreBackendsReady() (bool, error) {
	for _, b := range ds.backendList {
		_, err := b.Backend.IsReady(b.Meta)
		if err != nil {
			return false, fmt.Errorf("There was an error getting backend %s ready: %q", b.Name, err)
		}

	}

	return true, nil
}

// For READ operations we take the first backend as single point of truth
// for simplicity (i.e. we don't deal with conflicts between backends)

func (ds *DeploymentState) ListSlots(appName string) ([]*schema.SlotData, error) {
	if len(ds.backendList) < 1 {
		return nil, fmt.Errorf("No backend found: %v", ds.backendList)
	}
	b := ds.backendList[0]

	slotData, err := b.Backend.ListSlots(b.Meta, appName)
	if err != nil {
		return nil, fmt.Errorf("Failed to list slots for %q: %s", appName, err)
	}

	return slotData, nil
}

func (ds *DeploymentState) GetSlot(appName, slotId string) (*schema.SlotData, error) {
	if len(ds.backendList) < 1 {
		return nil, fmt.Errorf("No backend found: %v", ds.backendList)
	}
	b := ds.backendList[0]

	slotData, err := b.Backend.GetSlot(b.Meta, appName, slotId)
	if err != nil {
		return nil, fmt.Errorf("Failed to get slot %s for %q: %s", slotId, appName, err)
	}

	return slotData, nil
}

func (ds *DeploymentState) ListLastDeployments(appName, slotId string, limit int) ([]*schema.DeploymentData, error) {
	if len(ds.backendList) < 1 {
		return nil, fmt.Errorf("No backend found: %v", ds.backendList)
	}
	b := ds.backendList[0]

	deployments, err := b.Backend.ListSortedDeploymentsForSlotId(
		b.Meta, appName, slotId, limit)
	if err != nil {
		return nil, fmt.Errorf("Failed to list last %d deployments of %q/%q: %s", limit, appName, slotId, err)
	}

	return deployments, nil
}

func (ds *DeploymentState) ListApplications() ([]*schema.ApplicationData, error) {
	if len(ds.backendList) < 1 {
		return nil, fmt.Errorf("No backend found: %v", ds.backendList)
	}
	b := ds.backendList[0]

	apps, err := b.Backend.ListApplications(b.Meta)
	if err != nil {
		return nil, fmt.Errorf("Failed listing applications: %s", err)
	}
	return apps, err
}

func (ds *DeploymentState) GetApplication(name string) (*schema.ApplicationData, error) {
	if len(ds.backendList) < 1 {
		return nil, fmt.Errorf("No backend found: %v", ds.backendList)
	}
	b := ds.backendList[0]

	app, err := b.Backend.GetApplication(b.Meta, name)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (ds *DeploymentState) SaveApplication(name string, data *schema.ApplicationData) error {
	for _, b := range ds.backendList {
		err := b.Backend.SaveApplication(b.Meta, name, data)
		if err != nil {
			return fmt.Errorf("Failed to save application data to backend %s: %q", b.Name, err)
		}
	}

	return nil
}

func (ds *DeploymentState) SupportsWriteLock() (bool, error) {
	if len(ds.backendList) < 1 {
		return false, fmt.Errorf("No backend found: %v", ds.backendList)
	}
	b := ds.backendList[0]

	return b.Backend.SupportsWriteLock(), nil
}

func (ds *DeploymentState) GetDeployment(appName, slotId, deploymentId string) (*schema.DeploymentData, error) {
	if len(ds.backendList) < 1 {
		return nil, fmt.Errorf("No backend found: %v", ds.backendList)
	}
	b := ds.backendList[0]

	deployment, err := b.Backend.GetDeployment(
		b.Meta, appName, slotId, deploymentId)

	if err != nil {
		return nil, fmt.Errorf("Failed getting deployment %s of %q for slot %s: %s",
			deploymentId, appName, slotId, err)
	}

	return deployment, nil
}

func (ds *DeploymentState) BeginDeployment(appName, slotId string, isDestroy bool, pilot *schema.DeployPilot, startTime time.Time,
	vars map[string]string) (*schema.DeploymentData, error) {
	deploymentId := generateUniqueDeploymentId()

	tf := schema.TerraformRun{
		IsDestroy:        isDestroy,
		Variables:        vars,
		TerraformVersion: rt.TerraformVersion,
	}

	data := &schema.DeploymentData{
		DeploymentId: deploymentId,
		DeployPilot:  pilot,
		Terraform:    &tf,
		RTVersion:    rt.Version,
		StartTime:    startTime,
	}

	for _, b := range ds.backendList {
		slotData, err := b.Backend.GetSlot(b.Meta, appName, slotId)
		if err != nil {
			_, ok := err.(*backends.SlotNotFound)
			if !ok {
				return nil, fmt.Errorf("Unable to get slot data for %s / %s: %s", appName, slotId, err)
			}
			// Creating new slot here
			slotData = &schema.SlotData{
				IsActive: true,
			}
		}
		slotData.LastDeployPilot = pilot
		slotData.LastDeploymentStartTime = startTime

		err = b.Backend.SaveSlot(b.Meta, appName, slotId, slotData)
		if err != nil {
			return nil, fmt.Errorf("Unable to save slot data for %s / %s: %s", appName, slotId, err)
		}

		err = b.Backend.SaveDeployment(b.Meta, appName, slotId, deploymentId, data)
		if err != nil {
			return nil, fmt.Errorf("There was an error beginning the deployment with backend %s: %q", b.Name, err)
		}
	}

	return data, nil
}

func (ds *DeploymentState) FinishDeployment(appName, slotId, deploymentId string, isActive bool,
	data *schema.DeploymentData, tfRun *schema.FinishedTerraformRun) error {

	data.Terraform.StartTime = tfRun.StartTime
	data.Terraform.FinishTime = tfRun.FinishTime
	data.Terraform.PlanStartTime = tfRun.PlanStartTime
	data.Terraform.PlanFinishTime = tfRun.PlanFinishTime
	data.Terraform.ResourceDiff = tfRun.ResourceDiff
	data.Terraform.Outputs = tfRun.Outputs
	data.Terraform.ExitCode = tfRun.ExitCode
	data.Terraform.Warnings = tfRun.Warnings
	data.Terraform.Stderr = tfRun.Stderr

	for _, b := range ds.backendList {
		err := b.Backend.SaveDeployment(b.Meta, appName, slotId, deploymentId, data)
		if err != nil {
			return fmt.Errorf("There was an error finishing the deployment with backend %s: %q", b.Name, err)
		}

		slotData, err := b.Backend.GetSlot(b.Meta, appName, slotId)
		if err != nil {
			return fmt.Errorf("Unable to get slot data for %s / %s: %s", appName, slotId, err)
		}
		slotData.IsActive = isActive
		if slotData.LastTerraformRun == nil {
			slotData.LastTerraformRun = &schema.TerraformRun{}
		}
		slotData.LastTerraformRun.PlanStartTime = data.Terraform.PlanStartTime
		slotData.LastTerraformRun.PlanFinishTime = data.Terraform.PlanFinishTime
		slotData.LastDeployPilot = data.DeployPilot
		slotData.LastTerraformRun = data.Terraform
		err = b.Backend.SaveSlot(b.Meta, appName, slotId, slotData)
		if err != nil {
			return fmt.Errorf("Unable to save slot data for %s / %s: %s", appName, slotId, err)
		}
	}

	return nil
}

func (ds *DeploymentState) GetSlotCounter(prefix string, appData *schema.ApplicationData) (int64, bool, error) {
	counter, ok := appData.SlotCounters[prefix]
	if !ok {
		return 0, false, nil
	}

	return counter, true, nil
}

func (ds *DeploymentState) AddSlotCounter(prefix string, appData *schema.ApplicationData) (*schema.ApplicationData, error) {
	if c, ok := appData.SlotCounters[prefix]; ok {
		return appData, fmt.Errorf("Slot counter %s already exists (current value: %d)", prefix, c)
	}

	if len(appData.SlotCounters) == 0 {
		appData.SlotCounters = make(map[string]int64, 0)
	}

	appData.SlotCounters[prefix] = 0

	return appData, nil
}

func (ds *DeploymentState) IncrementSlotCounter(prefix string, appData *schema.ApplicationData) (int64, *schema.ApplicationData, error) {
	appData.SlotCounters[prefix] += 1
	return appData.SlotCounters[prefix], appData, nil
}

func (ds *DeploymentState) DeleteSlot(appName, slotId string) error {
	var _errors error
	for _, b := range ds.backendList {
		err := b.Backend.DeleteSlot(b.Meta, appName, slotId)
		if err != nil {
			_errors = multierror.Append(err)
		}
	}
	return _errors
}

// TODO: Create ListSlotNames function if ListSlots becomes inefficient
// As DeleteSlotCounter does not need to make a connection and only needs SlotIds
func (ds *DeploymentState) DeleteSlotCounter(prefix string, appData *schema.ApplicationData) (*schema.ApplicationData, error) {
	if _, ok := appData.SlotCounters[prefix]; !ok {
		return appData, fmt.Errorf("Slot counter with prefix %s does not exist", prefix)
	}
	delete(appData.SlotCounters, prefix)
	return appData, nil
}

// DIRTY HACK! 🐉
// File-based backends (like S3) list objects in lexicographical order
// and offer no easy ways to efficiently sort objects/files.
// Reversed timestamps sorted lexicographically are sorted from newest to oldest.
// This reduces complexity and data usage as we don't have to re-sort
// the list of deployments nor paginate if we only need latest deployment
func generateUniqueDeploymentId() string {
	// TODO: If we can avoid file-based backends, we can generate IDs any way we want
	maxInt := math.MaxInt64
	reversedTimestamp := maxInt - int(time.Now().UTC().Unix())
	return fmt.Sprintf("%020d", reversedTimestamp)
}
