package main

import (
	"github.com/TimeInc/ape-dev-rt/commons"
	"github.com/TimeInc/ape-dev-rt/git"
	"github.com/TimeInc/ape-dev-rt/validators"
	"github.com/urfave/cli"
)

type FlagDefinitions struct {
	Config            cli.StringFlag
	Profile           cli.StringFlag
	Module            cli.StringFlag
	Path              commons.StringFlag
	Skeleton          commons.StringFlag
	AwsProfile        commons.StringFlag
	AwsRegion         commons.StringFlag
	Environment       commons.StringFlag
	AppName           commons.StringFlag
	SlotID            commons.StringFlag
	OlderThan         commons.StringFlag
	Namespace         commons.StringFlag
	OutputName        cli.StringFlag
	Target            cli.StringFlag
	Refresh           cli.BoolFlag
	Verbose           cli.BoolFlag
	EnableFileLogging cli.BoolTFlag
	YesOverride       cli.BoolFlag
	Force             cli.BoolFlag
	Variable          cli.StringSliceFlag
	SlotPrefix        cli.StringFlag
	PreviousSlot      cli.BoolFlag
	Xlegacy           cli.BoolFlag
}

var flags = FlagDefinitions{
	Config: cli.StringFlag{
		Name:   "config",
		Usage:  "Path to the RT configuration file, defaults to ~/.rt/config",
		EnvVar: "RT_CONFIG",
	},

	Profile: cli.StringFlag{
		Name:   "profile",
		Usage:  "Name profile to load from RT configuration, ~/.rt/config",
		EnvVar: "RT_PROFILE",
	},

	Module: cli.StringFlag{
		Name:  "module",
		Usage: "Name of the module the resource belongs to",
	},

	Path: commons.StringFlag{
		StringFlag: cli.StringFlag{
			Name:   "path",
			Usage:  "Local path to clone of the apps GIT repository (" + git.AppConfigRepoUrl + ")",
			EnvVar: "RT_PATH",
		},
		Validator: validators.StringIsValidPath,
	},

	Skeleton: commons.StringFlag{
		StringFlag: cli.StringFlag{
			Name:  "skeleton",
			Usage: "",
			Value: "default",
		},
		Validator: validators.NonEmptyString,
	},

	AwsProfile: commons.StringFlag{
		StringFlag: cli.StringFlag{
			Name:   "aws-profile",
			Usage:  "Specify the AWS Credential Profile used to resolve AWS credentials",
			Value:  "default",
			EnvVar: "RT_AWS_PROFILE",
		},
		Validator: validators.NonEmptyString,
	},

	AwsRegion: commons.StringFlag{
		StringFlag: cli.StringFlag{
			Name:   "aws-region",
			Usage:  "Specify the AWS Region where application resources reside",
			Value:  "us-east-1",
			EnvVar: "RT_AWS_REGION",
		},
		Validator: validators.NonEmptyString,
	},

	Environment: commons.StringFlag{
		StringFlag: cli.StringFlag{
			Name:   "env",
			Usage:  "The AWS environment name, used to lookup the application state in S3 ([a-zA-Z0-9_]+)",
			EnvVar: "RT_ENV",
		},
		Validator: validators.IsEnvironmentNameValid,
	},

	AppName: commons.StringFlag{
		StringFlag: cli.StringFlag{
			Name:  "app",
			Usage: "Name of the AWS application ([a-z0-9-_]+)",
		},
		Validator: validators.IsApplicationNameValid,
	},

	SlotID: commons.StringFlag{
		StringFlag: cli.StringFlag{
			Name:  "slot-id",
			Usage: "Name of the slot",
		},
		Validator: validators.IsSlotIDValid,
	},

	Refresh: cli.BoolFlag{
		Name:  "refresh",
		Usage: "Will update the cached realease tool state by interrogating the AWS API",
	},

	Verbose: cli.BoolFlag{
		Name:   "verbose, v",
		Usage:  "Prints all log messages",
		EnvVar: "RT_LOG",
	},

	EnableFileLogging: cli.BoolTFlag{
		Name:   "enable-file-logging",
		Usage:  "Sends all log messages to designated location",
		EnvVar: "RT_ENABLE_FILE_LOGGING",
	},

	YesOverride: cli.BoolFlag{
		Name:  "y, yes, yas", //yas is a term used by young people to mean yes
		Usage: "Makes operations non-interactive by confirming all prompts and defaults",
	},

	Force: cli.BoolFlag{
		Name:  "f, force",
		Usage: "Forces an Apply or Destroy command without a need for a creation or deletion of resources",
	},

	Xlegacy: cli.BoolFlag{
		Name:  "x, xlegacy",
		Usage: "Terraform core graphing is done using pre-Terraform v0.8.0 paths",
	},

	Variable: cli.StringSliceFlag{
		Name:  "var",
		Usage: "Variable to pass to Terraform. e.g. -var=key=val",
	},

	SlotPrefix: cli.StringFlag{
		Name:  "slot-prefix",
		Usage: "Slot prefix",
	},

	PreviousSlot: cli.BoolFlag{
		Name:  "previous-slot",
		Usage: "Whether to use previous slot with given slot prefix",
	},

	OlderThan: commons.StringFlag{
		StringFlag: cli.StringFlag{
			Name:  "older-than",
			Usage: "Time period in which to cleanup slots (e.g. 7day for 7 days, see github.com/ninibe/bigduration)",
		},
		Validator: validators.IsBigDurationValid,
	},

	Namespace: commons.StringFlag{
		StringFlag: cli.StringFlag{
			Name:  "namespace",
			Usage: "Namespace for the tfstate files. Value defaults to AWS Account ID",
			Value: "default",
		},
		Validator: validators.IsNamespaceValid,
	},

	OutputName: cli.StringFlag{
		Name:  "name",
		Usage: "Name of the output value",
	},

	Target: cli.StringFlag{
		Name:  "target",
		Usage: "A resource address to target. [module path][resource spec]",
	},
}
