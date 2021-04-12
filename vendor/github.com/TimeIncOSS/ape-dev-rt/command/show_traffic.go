package command

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/TimeIncOSS/ape-dev-rt/aws"
	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
	"github.com/TimeIncOSS/ape-dev-rt/terraform"
)

func ShowTraffic(c *commons.Context) error {
	ds, ok := c.CliContext.App.Metadata["ds"].(*deploymentstate.DeploymentState)
	if !ok {
		return fmt.Errorf("Unable to find Deployment State in metadata")
	}

	defaultAWS := aws.NewAWS(c.GlobalString("aws-profile"), "us-east-1")
	regionalAWS := aws.NewAWS(c.GlobalString("aws-profile"), c.String("aws-region"))
	user, err := defaultAWS.User()
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

	var slots []*slotData
	if appData.UseCentralGitRepo {
		return deprecatedGitError()
	}

	app, err := ds.GetApplication(c.String("app"))
	if err != nil {
		return err
	}
	if app.InfraOutputs == nil {
		return fmt.Errorf("No infra outputs found for %q", c.String("app"))
	}

	internalAppName, ok := app.InfraOutputs[terraform.AppName]
	if !ok {
		return fmt.Errorf("Output %q not found", terraform.AppName)
	}

	_slots, err := ds.ListSlots(c.String("app"))
	if err != nil {
		return err
	}
	for _, s := range _slots {
		if !s.IsActive {
			continue
		}
		if s.LastTerraformRun == nil {
			continue
		}

		gc := &slotData{
			SlotId:     s.SlotId,
			Variables:  sortedVars(s.LastTerraformRun.Variables),
			FinishTime: s.LastTerraformRun.FinishTime,
		}

		slots = append(slots, gc)
	}

	return decorateAndPrintSortedVersions(internalAppName, c.String("env"), slots, regionalAWS,
		os.Stdout, colour)
}

func sortedVars(vars map[string]string) string {
	s := "vars: map["

	// Sort the keys
	keys := make([]string, len(vars), len(vars))
	i := 0
	for k, _ := range vars {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	i = 0
	for _, k := range keys {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%q:%q", k, vars[k])
		i++
	}
	s += "]"
	return s
}

func decorateAndPrintSortedVersions(internalAppName, env string, slots []*slotData, a *aws.AWS,
	w io.Writer, colour *colours) error {

	var dateLayout = "Mon Jan 02 15:04:05 -0700 2006"
	fmt.Fprintf(w, "Discovering resources in AWS region %s\n\n", colour.boldWhite(*a.Region))
	for _, version := range slots {
		scalingGroup, err := a.GetScalingGroupForSlotId(env, internalAppName, version.SlotId)
		if err != nil {
			return err
		}

		var versionSlug = fmt.Sprintf("%s (%s) %s", version.SlotId,
			version.FinishTime.Format(dateLayout),
			version.Variables)

		if len(scalingGroup) > 0 {
			var balancers []*aws.Balancer
			balancers, err = a.GetBalancersFromScalingGroup(scalingGroup)
			if err != nil {
				return err
			}

			resourceCount := ""
			if len(balancers) > 1 {
				resourceCount = fmt.Sprintf("%v ELBs", len(balancers))
			} else {
				resourceCount = fmt.Sprintf("%v ELB", len(balancers))
			}
			if len(balancers) == 0 {
				resourceCount = colour.boldWhite(resourceCount)
			} else {
				resourceCount = colour.boldGreen(resourceCount)
			}
			fmt.Fprintf(w, "%s - %s", versionSlug, resourceCount)

			instanceIds, err := a.GetInstanceIdsFromScalingGroup(scalingGroup)
			if err != nil {
				fmt.Println("")
				return err
			}

			var instanceToIpMap = make(map[string]string, 0)
			instanceToIpMap, err = a.GetPrivateIpsForInstanceIds(instanceIds)
			if err != nil {
				log.Printf("[ERROR] Unable to get private IPs of instance IDs (%v): %s", instanceIds, err)
			}

			for _, b := range balancers {
				decorateAndPrintBalancer(b, scalingGroup, w, colour)
				instances, err := a.DescribeBalancedInstanceHealth(b.Name)
				if err != nil {
					return err
				}
				for _, i := range instances {
					decorateAndPrintInstanceHealth(i, instanceIds, instanceToIpMap, w, colour)
				}
				fmt.Print("\n")
			}
		} else {
			fmt.Fprintf(w, "%s - %s", versionSlug, colour.boldRed(fmt.Sprintf("no ASG")))
		}
		fmt.Fprint(w, "\n")
	}
	return nil
}

func decorateAndPrintBalancer(b *aws.Balancer, scalingGroup string, w io.Writer, colour *colours) {
	balancerState := ""

	switch b.State {
	case "Added":
		balancerState = fmt.Sprintf("%s to", colour.boldGreen(b.State))
	case "InService":
		balancerState = fmt.Sprintf("%s for", colour.boldGreen(b.State))
	case "Adding":
		balancerState = fmt.Sprintf("%s to", colour.boldYellow(b.State))
	case "Removing":
		balancerState = fmt.Sprintf("%s from", colour.boldRed(b.State))
	}
	fmt.Fprintf(w, "\n  ELB %s %s ASG %s", b.Name, balancerState, scalingGroup)
	return
}

func decorateAndPrintInstanceHealth(i *aws.InstanceHealth, instanceIds []*string, idToIp map[string]string, w io.Writer, colour *colours) {
	state := "Unknown"

	if i.State == "InService" {
		state = colour.boldGreen(fmt.Sprintf("%s", i.State))
	} else if i.State == "OutOfService" {
		state = colour.boldRed(fmt.Sprintf("%s", i.State))
	}
	suffix := ""
	for _, id := range instanceIds {
		if *id == i.InstanceID {
			suffix = fmt.Sprintf(" (this version)")
			if ip, ok := idToIp[i.InstanceID]; ok {
				suffix += fmt.Sprintf(" - %s", ip)
			}
		}
	}
	fmt.Fprintf(w, "\n    EC2 %s %s%s", i.InstanceID, state, suffix)
	return
}

type slotData struct {
	SlotId     string
	Variables  string
	FinishTime time.Time
}
