package schema

import (
	"testing"
)

func TestApplicationDataFromJSON(t *testing.T) {
	newVersion := []byte(`{"v":99999}`)
	app := &ApplicationData{}
	err := app.FromJSON(newVersion)
	if err == nil {
		t.Fatal("Expected error on higher schema version, none received")
	}
}

func TestSlotDataFromJSON(t *testing.T) {
	newVersion := []byte(`{"v":99999}`)
	slot := &SlotData{}
	err := slot.FromJSON(newVersion)
	if err == nil {
		t.Fatal("Expected error on higher schema version, none received")
	}
}

func TestDeploymentDataFromJSON(t *testing.T) {
	newVersion := []byte(`{"v":99999}`)
	deployment := &DeploymentData{}
	err := deployment.FromJSON(newVersion)
	if err == nil {
		t.Fatal("Expected error on higher schema version, none received")
	}
}
