package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/RevH/ipinfo"
	"github.com/TimeIncOSS/ape-dev-rt/aws"
	"github.com/TimeIncOSS/ape-dev-rt/command"
	"github.com/TimeIncOSS/ape-dev-rt/commons"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
	"github.com/TimeIncOSS/ape-dev-rt/hcl"
	"github.com/TimeIncOSS/ape-dev-rt/rt"
	"github.com/mitchellh/go-homedir"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
)

var boldBlue = chalk.Blue.NewStyle().WithTextStyle(chalk.Bold).Style
var boldYellow = chalk.Yellow.NewStyle().WithTextStyle(chalk.Bold).Style

var Commands = []cli.Command{
	{
		Name:   "create-app",
		Usage:  "Initialize a new application definition",
		Action: wrapCommand(command.CreateApp),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Skeleton,
			flags.AppName,
			flags.Path,
		},
	},
	{
		Name:   "apply-infra",
		Usage:  "Provision application infrastructure",
		Action: wrapCommand(command.ApplyInfra),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
			flags.Variable,
			flags.YesOverride,
			flags.Target,
			flags.Namespace,
			flags.Force,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "destroy-infra",
		Usage:  "Destroy application infrastructure.",
		Action: wrapCommand(command.DestroyInfra),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
			flags.YesOverride,
			flags.Target,
			flags.Variable,
			flags.Namespace,
			flags.Force,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "diff-infra",
		Usage:  "Show terraform plan for infrastructure level",
		Action: wrapCommand(command.DiffInfra),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
			flags.YesOverride,
			flags.Target,
			flags.Variable,
			flags.Namespace,
			flags.Force,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "deploy",
		Usage:  "Deploy an application into a given environment & slot",
		Action: wrapCommand(command.Deploy),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
			flags.SlotID,
			flags.YesOverride,
			flags.Variable,
			flags.SlotPrefix,
			flags.Target,
			flags.Namespace,
			flags.Force,
		},
		ArgsUsage: "<path-to-tf-cfgs>",
		Before:    beforeAuthedCommand,
	},
	{
		Name:   "deploy-destroy",
		Usage:  "Destroy an application from a given environment & slot",
		Action: wrapCommand(command.DeployDestroy),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
			flags.SlotID,
			flags.YesOverride,
			flags.SlotPrefix,
			flags.PreviousSlot,
			flags.Target,
			flags.Variable,
			flags.Namespace,
			flags.Force,
		},
		ArgsUsage: "<path-to-tf-cfgs>",
		Before:    beforeAuthedCommand,
	},
	{
		Name:   "diff-deploy",
		Usage:  "Show terraform plan for deployment(version) level",
		Action: wrapCommand(command.DiffDeploy),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
			flags.SlotID,
			flags.YesOverride,
			flags.Variable,
			flags.SlotPrefix,
			flags.Target,
			flags.Namespace,
			flags.Force,
		},
		ArgsUsage: "<path-to-tf-cfgs>",
		Before:    beforeAuthedCommand,
	},
	{
		Name:   "disable-traffic",
		Usage:  "Detach load-balancers from the version scaling-group",
		Action: wrapCommand(command.DisableTraffic),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.AwsRegion,
			flags.AppName,
			flags.Environment,
			flags.SlotID,
			flags.SlotPrefix,
			flags.PreviousSlot,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "enable-traffic",
		Usage:  "Attach load-balancers to the version scaling-group",
		Action: wrapCommand(command.EnableTraffic),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.AwsRegion,
			flags.AppName,
			flags.Environment,
			flags.SlotID,
			flags.SlotPrefix,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "show-traffic",
		Usage:  "Show which Scaling Groups have Load Balancers attached",
		Action: wrapCommand(command.ShowTraffic),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.AwsRegion,
			flags.AppName,
			flags.Environment,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "list-apps",
		Usage:  "list all apps for a given environment",
		Action: wrapCommand(command.ListApps),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.Refresh,
			flags.Namespace,
		},
		Before:   beforeAuthedCommand,
		Category: "app-not-required",
	},
	{
		Name:   "list-slots",
		Usage:  "List all slots for a given app in a given environment",
		Action: wrapCommand(command.ListSlots),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "list-slot-prefixes",
		Usage:  "List all slot prefixes for a given app in a given environment",
		Action: wrapCommand(command.ListSlotPrefixes),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "cleanup-slots",
		Usage:  "Cleanup inactive slots for a given app in a given environment",
		Action: wrapCommand(command.CleanupSlots),
		Flags: []cli.Flag{
			flags.AppName,
			flags.AwsProfile,
			flags.Environment,
			flags.OlderThan,
			flags.Verbose,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "add-slot-prefix",
		Usage:  "Add a new slot prefix for a given app in a given environment",
		Action: wrapCommand(command.AddSlotPrefix),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
		},
		ArgsUsage: "slot-id",
		Before:    beforeAuthedCommand,
	},
	{
		Name:   "delete-slot-prefix",
		Usage:  "Delete a slot prefix for a given app in a given environment",
		Action: wrapCommand(command.DeleteSlotPrefix),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
		},
		ArgsUsage: "slot-id",
		Before:    beforeAuthedCommand,
	},
	{
		Name:   "taint-infra-resource",
		Usage:  "Taint an infrastructure resource",
		Action: wrapCommand(command.TaintUntaintInfraResource),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.AppName,
			flags.Environment,
			flags.Module,
			flags.Namespace,
		},
		ArgsUsage: "resource-to-taint",
		Before:    beforeAuthedCommand,
	},
	{
		Name:   "untaint-infra-resource",
		Usage:  "Untaint an infrastructure resource",
		Action: wrapCommand(command.TaintUntaintInfraResource),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.AppName,
			flags.Environment,
			flags.Module,
			flags.Namespace,
		},
		ArgsUsage: "resource-to-untaint",
		Before:    beforeAuthedCommand,
	},
	{
		Name:   "taint-deployed-resource",
		Usage:  "Taint a deployed resource",
		Action: wrapCommand(command.TaintUntaintDeployedResource),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.AppName,
			flags.Environment,
			flags.SlotID,
			flags.Module,
			flags.Namespace,
		},
		ArgsUsage: "<path-to-tf-cfgs> <resource-to-untaint>",
		Before:    beforeAuthedCommand,
	},
	{
		Name:   "untaint-deployed-resource",
		Usage:  "Untaint a deployed resource",
		Action: wrapCommand(command.TaintUntaintDeployedResource),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.AppName,
			flags.Environment,
			flags.SlotID,
			flags.Module,
			flags.Namespace,
		},
		ArgsUsage: "<path-to-tf-cfgs> <resource-to-untaint>",
		Before:    beforeAuthedCommand,
	},
	{
		Name:   "version",
		Usage:  "Get version",
		Action: wrapCommand(rt.GetVersion),
		Flags: []cli.Flag{
			flags.AwsProfile,
		},
		Category: "app-not-required",
	},
	{
		Name:   "output",
		Usage:  "List output variables of a given app",
		Action: wrapCommand(command.Output),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.AppName,
			flags.Environment,
			flags.OutputName,
			flags.Namespace,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "slot-output",
		Usage:  "List output variables of a given app and slot-id",
		Action: wrapCommand(command.SlotOutput),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.AppName,
			flags.Environment,
			flags.SlotID,
			flags.OutputName,
			flags.Namespace,
		},
		Before:    beforeAuthedCommand,
		ArgsUsage: "<path-to-tf-cfgs>",
	},
	{
		Name:   "list-deployments",
		Usage:  "List last deployment of a given app in a given environment",
		Action: wrapCommand(command.ListDeployments),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
			flags.Refresh,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "validate-infra",
		Usage:  "Validates the current working directory for valid Terraform code",
		Action: wrapCommand(command.ValidateInfra),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
			flags.Namespace,
		},
		Before: beforeAuthedCommand,
	},
	{
		Name:   "validate-slots",
		Usage:  "Validates the slots directories for valid Terraform code",
		Action: wrapCommand(command.ValidateSlots),
		Flags: []cli.Flag{
			flags.AwsProfile,
			flags.Environment,
			flags.AppName,
			flags.Namespace,
		},
		Before: beforeAuthedCommand,
	},
}

func beforeAuthedCommand(c *cli.Context) error {
	user, err := authenticateWithAWS(c)
	if err != nil {
		return err
	}
	err = getCurrentIpAddress(c)
	if err != nil {
		log.Printf("[WARN] Unable to get IP address: %s", err)
	}

	cfg, cfgPath, err := loadConfig(c.Command, c.String("env"), user.AccountID)
	if err != nil {
		return err
	}
	if cfg.RemoteState == nil {
		url := "https://github.com/TimeIncOSS/ape-dev-rt/blob/master/docs/remote_state.md"
		return fmt.Errorf("No 'remote_state' block found in %q. See %s for more details.",
			cfgPath, url)
	}
	c.App.Metadata["remote_state"] = cfg.RemoteState

	if c.String("env") == "" {
		return errors.New("No environment defined. Please use -env flag")
	}

	ds, err := loadDeploymentState(c.String("env"), c.String("app"), cfg.DeploymentState)
	if err != nil {
		return err
	}
	c.App.Metadata["ds"] = ds

	return nil
}

func authenticateWithAWS(c *cli.Context) (*aws.User, error) {
	a := aws.NewAWS(c.GlobalString("aws-profile"), "us-east-1")
	log.Println("[INFO] Verifying AWS credentials")
	user, err := a.User()
	if err != nil {
		return nil, err
	}

	c.App.Metadata["session"] = a
	c.App.Metadata["user"] = user

	log.Printf("[DEBUG] Received AWS Account: %#v", user)
	fmt.Printf("Authenticated as %s (%s) @ %s \n", boldBlue(user.Name), boldBlue(user.UserID), boldBlue(user.AccountID))

	return user, nil
}

func getCurrentIpAddress(c *cli.Context) error {
	// TODO: Just use request headers instead when RT becomes a server w/ API
	ip, err := ipinfo.MyIP()
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Detected IP: %s", ip.IP)

	fmt.Printf("Current IP Address: %s\n", boldBlue(ip.IP))
	c.App.Metadata["current_ip"] = ip.IP
	return nil
}

func loadConfig(cmd cli.Command, env, awsAccId string) (*hcl.HclConfig, string, error) {
	var cfgPath string
	var err error
	if cmd.HasName("list-apps") {
		cfgPath, err = homedir.Expand("~/.rt/")
		if err != nil {
			return nil, cfgPath, err
		}
	} else {
		cfgPath, err = os.Getwd()
		if err != nil {
			return nil, cfgPath, err
		}
	}

	return hcl.LoadConfigFromPath(env, awsAccId, cfgPath)
}

func loadDeploymentState(env, appName string, cfg *hcl.DeploymentState) (*deploymentstate.DeploymentState, error) {
	ds, err := deploymentstate.New(cfg)
	if err != nil {
		return ds, fmt.Errorf("Failed to load deployment state backends for %s: %s", env, err)
	}

	ready, err := ds.AreBackendsReady()
	if err != nil {
		return ds, fmt.Errorf("Deployment state backend(s) not ready for %s (environment): %s", env, err)
	}
	if !ready {
		return ds, fmt.Errorf("Deployment state backend(s) not ready for %s (environment).", env)
	}

	supportsLock, err := ds.SupportsWriteLock()
	if err != nil {
		return nil, fmt.Errorf("Unable to verify lock support for deployment state backend(s): %s", err)
	}
	if !supportsLock {
		fmt.Printf("%s Locking is not supported. Parallel releases of %s may cause issues, check with your team.\n\n",
			"Note:", appName)
	}

	return ds, nil
}

// produces a cli.Command.Action by wrapping our custom Action contract and passing
// a configuration backed Context instance to the wrapped Action
func wrapCommand(cmd func(c *commons.Context) error) func(*cli.Context) error {
	return func(cliContext *cli.Context) error {
		log.Printf("[DEBUG] Executing command: %q", os.Args)
		c := commons.NewContext(cliContext)

		if cliContext.Command.Category == "DEPRECATED" {
			fmt.Printf("%s %s\n", boldYellow("WARNING:"), boldYellow(cliContext.Command.Usage))
		}

		if !(cliContext.Command.Category == "app-not-required") {
			if cliContext.String("app") == "" {
				return fmt.Errorf("No application name defined for environment '%s'. Please use -app flag", cliContext.String("env"))
			}
		}

		flags := append(cliContext.App.Flags, cliContext.Command.Flags...)
		for _, f := range flags {
			if vf, ok := f.(commons.ValidatedFlag); ok {
				err := vf.Validate(c)
				if err != nil {
					return err
				}
			}
		}

		err := cmd(c)
		if err != nil {
			return err
		}

		return nil

	}
}
