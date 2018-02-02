# Assumptions

What assumptions RT makes about the **git repository** which contains Terraform configs
(currently https://github.com/TimeInc/ape-dev-rt-apps)?

## 1. Directory inside root == unique application

Any directory that is inside root of the repository is considered as unique application
with unique name being the name of the directory.

RT expects this directory to contain valid Terraform config files
which get processed as part of any `*-app` command.

## 2. Version-specific infrastructure

All `*-version` commands expect `%ROOT%/APPLICATION_NAME/version/` to exist
and contain valid Terraform config files. **Each git commit SHA** that touched
this directory is treated as a separate deployable version.

## 3. Any `*.tpl` files get processed as Go templates

See [detailed documentation](https://golang.org/pkg/text/template/) in case you want to take advantage
of the templating system which also allows you to put conditionals into `tf` configs and separate out
differences between environments.

The following variables are available for use:

 - `/APPLICATION` (i.e. interpolated as part of any `*-app` command)
   - `{{.AwsAccountId}}` - discovered AWS Account ID (via provided AWS credentials)
   - `{{.Environment}}` - environment (`-env` flag)
   - `{{.AppName}}` - application name (`-app` flag)
 - `/APPLICATION/version/` (i.e. interpolated as part of any `*-version` command)
   - `{{.AwsAccountId}}` - discovered AWS Account ID (via provided AWS credentials)
   - `{{.Environment}}` - environment (`-env` flag)
   - `{{.AppName}}` - application name (`-app` flag)

## 4. `*-traffic` commands (ASG & ELB)

`*-traffic` commands make the following assumptions about your infrastructure

 - There's [`output`](https://www.terraform.io/docs/configuration/outputs.html)
    called `app` in one of the `/APPLICATION/*.tf` configs which outputs `appName` used below
 - There's a version-specific [**autoscaling group**](https://www.terraform.io/docs/providers/aws/r/autoscaling_group.html) in the chosen region (`us-east-1` by default)
    which follows tagging convention
   - `Name: test-appName-v000000-vinst`, i.e. `("%s-%s-v%s-vinst", environment, appName, appVersion)`
 - There's at least 1 [**ELB**](https://www.terraform.io/docs/providers/aws/r/elb.html)
    in the chosen region (`us-east-1` by default) which follows tagging convention
   - `App: appName`

If your ASGs/ELBs don't follow the conventions above the behavior is undefined.
