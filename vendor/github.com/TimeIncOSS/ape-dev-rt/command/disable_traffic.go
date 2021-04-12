package command

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/TimeIncOSS/ape-dev-rt/aws"
	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
	"github.com/TimeIncOSS/ape-dev-rt/terraform"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

func DisableTraffic(c *commons.Context) error {
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}

	regionalAWS := aws.NewAWS(c.GlobalString("aws-profile"), c.String("aws-region"))
	fmt.Printf("Operating on resources in AWS region %s\n\n", colour.boldWhite(*regionalAWS.Region))
	user, err := regionalAWS.User()
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Received AWS Account: %#v", user)

	appData, exists, err := BeginApplicationOperation(c.String("env"), c.String("app"), ds)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	if appData.UseCentralGitRepo {
		return deprecatedGitError()
	}

	slotId := c.String("slot-id")
	slotPrefix := c.String("slot-prefix")
	if slotId == "" && slotPrefix == "" {
		return fmt.Errorf("'slot-id' or 'slot-prefix' is required parameter for %q (migrated app)", c.String("app"))
	}
	if slotId != "" && slotPrefix != "" {
		return errors.New("You can specify either 'slot-id' or 'slot-prefix', not both.")
	}

	if slotPrefix != "" {
		counter, prefixExists, err := ds.GetSlotCounter(slotPrefix, appData)
		if err != nil {
			return err
		}
		if !prefixExists {
			return fmt.Errorf("Slot prefix %s does not exist", slotPrefix)
		}

		if c.Bool("previous-slot") {
			lastSlotId := fmt.Sprintf("%s%d", slotPrefix, counter)
			counter -= 1
			slotId = fmt.Sprintf("%s%d", slotPrefix, counter)
			fmt.Printf("Last slot ID is %s, disabling traffic to %s\n", lastSlotId, colour.boldRed(slotId))
		} else {
			slotId = fmt.Sprintf("%s%d", slotPrefix, counter)
			fmt.Printf("Disabling traffic to slot ID %s\n", colour.boldRed(slotId))
		}
	}

	app, err := ds.GetApplication(c.String("app"))
	if err != nil {
		return err
	}
	if app.InfraOutputs == nil {
		return fmt.Errorf("No infra outputs found for %q", c.String("app"))
	}
	outputs := app.InfraOutputs

	internalAppName, ok := outputs[terraform.AppName]
	if !ok {
		return fmt.Errorf("Output %q not found", terraform.AppName)
	}

	var scalingGroup string
	scalingGroup, err = regionalAWS.GetScalingGroupForSlotId(c.String("env"), internalAppName, slotId)
	if err != nil {
		return fmt.Errorf("Failed getting scaling group for %s slot %s", internalAppName, slotId)
	}
	if len(scalingGroup) == 0 {
		return fmt.Errorf("Slot %s has no scaling group\n", slotId)
	}

	slots, err := ds.ListSlots(c.String("app"))
	if err != nil {
		return err
	}

	hasAttachedBalancers := false
	for _, s := range slots {
		if s.IsActive && s.SlotId != slotId {
			scalingGroup, err := regionalAWS.GetScalingGroupForSlotId(c.String("env"), internalAppName, s.SlotId)
			if err != nil {
				return err
			}
			balancers, err := regionalAWS.GetBalancersFromScalingGroup(scalingGroup)
			if err != nil {
				return err
			}
			if len(balancers) > 0 {
				hasAttachedBalancers = true
			}
		}
	}
	if !hasAttachedBalancers {
		return fmt.Errorf("This is the only slot serving traffic, disabling it would cause downtime. " +
			"Do you intend to deprovision this slot/app? Use deploy-destroy instead.")
	}

	var balancers []string
	balancers, err = regionalAWS.GetBalancersForApp(internalAppName)
	if err != nil {
		return fmt.Errorf("Failed getting load balancers for %s", internalAppName)
	}

	if len(balancers) == 0 {
		return fmt.Errorf("No Load Balancer found for %s\n", internalAppName)
	}

	err = regionalAWS.DetachBalancersFromScalingGroup(balancers, scalingGroup)
	if err != nil {
		errCode := err.(awserr.Error).Code()
		switch errCode {
		case "ValidationError":
			if strings.Contains(err.Error(), "Trying to remove Load Balancers that are not part of the group") {
				return fmt.Errorf("ELBs are not attached to scaling group %s\n", scalingGroup)
			}
		}
		return fmt.Errorf("Failed detaching load balancers %s from scaling group %s\n", balancers, scalingGroup)
	}

	detachedNotice := fmt.Sprintf("Load Balancers have begun detaching from scaling group %s\n", scalingGroup)
	fmt.Printf("%s", colour.boldGreen(detachedNotice))
	return nil
}
