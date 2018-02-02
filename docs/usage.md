# Usage

```
$ ape-dev-rt
NAME:
   release tool - For amazing releases

USAGE:
   ape-dev-rt [global options] command [command options] [arguments...]

COMMANDS:
     create-app                 Initialize a new application definition
     apply-infra                Provision application infrastructure
     destroy-infra              Destroy application infrastructure.
     deploy                     Deploy an application into a given environment & slot
     deploy-destroy             Destroy an application from a given environment & slot
     disable-traffic            Detach load-balancers from the version scaling-group
     enable-traffic             Attach load-balancers to the version scaling-group
     show-traffic               Show which Scaling Groups have Load Balancers attached
     list-apps                  list-apps is deprecated as it would be ambiguous due to app/deploymentstate relationship
     list-slots                 List all slots for a given app in a given environment
     taint-infra-resource       Taint an infrastructure resource
     untaint-infra-resource     Untaint an infrastructure resource
     taint-deployed-resource    Taint a deployed resource
     untaint-deployed-resource  Untaint a deployed resource
     version                    Get version
     list-deployments           List last deployment of a given app in a given environment
     help, h                    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config							Path to the RT configuration file, defaults to ~/.rt/config [$RT_CONFIG]
   --profile						Name profile to load from RT configuration, ~/.rt/config [$RT_PROFILE]
   --verbose, -v					Prints all log messages [$RT_LOG]
   --enable-file-logging		Sends all log messages to designated location [$RT_ENABLE_FILE_LOGGING]
   --aws-profile "default"		Specified the AWS Credential Profile used to resolve AWS credentials [$RT_AWS_PROFILE]
   --help, -h						show help
   --generate-bash-completion

```

# Examples - apply & destroy

This expects the shared stack to be [already deployed](https://github.com/TimeInc/ape-dev-terraform/tree/master/401279337454-timeinc-damproject-test) outside of RT via Terraform and have the TF state stored in [a specific S3 bucket & path](https://github.com/TimeInc/ape-dev-terraform/tree/master/401279337454-timeinc-damproject-test#how-to-setup-key) by convention and the app `example` to de defined like:  [ape-dev-rt-apps](https://github.com/TimeInc/ape-dev-rt-apps/tree/master/example)

Expected valid environment names are currently `test` and `prod`. These are then injected [as a template variable](https://github.com/TimeInc/ape-dev-rt-apps/blob/master/shared-services.tf#L5) so that RT & TF know where to look for the `*.tfstate` files which are stored separately for each environment.

```
# Test environment
ape-dev-rt --aws-profile=ti-dam-test list-apps --env=test
ape-dev-rt --aws-profile=ti-dam-test list-slots --env=test --app=example

ape-dev-rt --aws-profile=ti-dam-test apply-infra --env=test --app=example
ape-dev-rt --aws-profile=ti-dam-test deploy --env=test --app=example --version="master"

ape-dev-rt --aws-profile=ti-dam-test deploy-destroy --env=test --app=example --version="master"
ape-dev-rt --aws-profile=ti-dam-test destroy-infra --env=test --app=example
```
```
# Production environment
ape-dev-rt --aws-profile=ti-dam-prod list-apps --env=prod
ape-dev-rt --aws-profile=ti-dam-prod list-slots --env=prod --app=example

ape-dev-rt --aws-profile=ti-dam-prod apply-infra --env=prod --app=example
ape-dev-rt --aws-profile=ti-dam-prod deploy --env=prod --app=example --version="master"

ape-dev-rt --aws-profile=ti-dam-prod deploy-destroy --env=prod --app=example --version="master"
ape-dev-rt --aws-profile=ti-dam-prod destroy-infra --env=prod --app=example
```

# Traffic Management

Release Tool [v0.4.0](https://github.com/TimeInc/ape-dev-rt/blob/master/CHANGELOG.md#040-march-10th-2016) introduces __Traffic Management__ to control the relationship between Auto Scaling Groups and Elastic Load Balancers.

A typical RT application will receive traffic on a number of Elastic Load Balancers. These ELBs become attached to the Auto Scaling Group for a single version of the application.

## Enable & Disable Traffic

- `enable-traffic` takes the same arguments as `deploy` (`env`,`app`,`slot-id`) and attaches ELBs to the ASG for that slot ID.

- `disable-traffic` takes the same arguments as `deploy` (`env`,`app`,`slot-id`) and detaches ELBs from the ASG for that slot ID.

## Show Traffic

- `show-traffic` takes the same arguments as `list-versions` (`env`,`app`). It describes active versions of the application, examines ASGs for those versions to determine what ELBs are attached, and displays the health-status of EC2 Instances attached to those ELBs.

Instances which correspond to the given ASG are noted as `(this version)`. This is helpful when ELBs are attached to multiple ASGs, as might happen when deploying a new release of the app.

## Tainting and Untainting a resource

Tainting a resource forces it to be destroyed and recreated on the next apply.
This does not modify infrastructure and only modifies the state file.
Resources can be manually marked as tainted or as a result of a provisioner failing on a resource.
Untainting a resource returns its state to normal.

See Terraform docs for more on [taint](https://www.terraform.io/docs/commands/taint.html)
and [untaint](https://www.terraform.io/docs/commands/untaint.html)

- `taint-app-resource` and `untaint-app-resource` take the arguments (`env`,`app`) as mandatory. You must also specify the resource you wish to taint.
	- E.g. `ape-dev-rt taint-app-resource -env=test -app=example aws_cloudwatch_log_group.application`
- You can optionally specify a module if the resource belongs to a module. To do so, specify the module flag.
	- E.g. `ape-dev-rt taint-app-resource -env=test -module=some-module -app=example aws_cloudwatch_log_group.application`

## Demo with the `example` app

Here's a simple deployment process, it took rougly 90 seconds.
![gif](http://brohenry-public-bucket.s3.amazonaws.com/rt_traffic_20160308.gif)

- **show-traffic** reveals an old version (`555f259`) is already "InService".
- **enable-traffic** puts the new version (`76feaa5`) into operation.
- We wait for **show-traffic** to have both versions "InService".
- **disable-traffic** takes the old version out of operation.

(You can open the above GIF image with `Preview.app` to see it frame-by-frame.)

```sh
# The demo uses a simple script to refresh output from show-traffic
watch-traffic() {
	START=`date +%s`;
	while true; do
		ape-dev-rt show-traffic -env=$1 -app=$2 > /tmp/wt;
		clear; echo -e "$(expr $(date +%s) - $START) second[s] elapsed\n";
		cat /tmp/wt; sleep 7;
	done; rm /tmp/wt
}
watch-traffic test example
```

# AWS Credentials

**This is probably an area of RT that could be improved.**

In the example above, RT will try and get credentials from all default providers [defined in the code](https://github.com/TimeInc/ape-dev-rt/blob/master/aws/aws.go#L162-L168):

 - `~/.aws/credentials`
 - Environment variables `AWS_*`

:warning: Please keep in mind that Terraform may use different providers and/or ordering and may change in time as [the upstream codebase evolves](https://github.com/hashicorp/terraform/blob/master/builtin/providers/aws/config.go#L339-L348) since the project is still young and we have seen changes in that part of codebase lately. :warning:

We typically don't hardcode credentials in tf templates, so `StaticProvider` gets skipped, so at the moment (`6a3ed429ade7bd55894b2074738fbb3345f55923`) the ordering is exactly opposite to RT:

 - Environment variables `AWS_*`
 - `~/.aws/credentials`

**It is recommended to always use `~/.aws/credentials` w/out `default` profile specified and unset any `AWS_` variables when using RT to avoid strange edge cases.**

## AWS config profiles

If you need managing more than 1 account (and you likely do) in `~/.aws/credentials`, you'll use profiles.

How can you tell both RT & TF to use specific profiles? Either via a CLI flag, e.g.

```
ape-dev-rt --aws-profile=ti-dam-test list-apps --env=test
```

or an environment variable

```
AWS_PROFILE=ti-dam-test ape-dev-rt list-apps --env=test
```
