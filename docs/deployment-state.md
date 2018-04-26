# Deployment State (introduced in 0.5.0)

We persist the following things in deployment state:

 - Applications
   - name
   - infrastructure outputs
 - Slots
   - Slot ID
   - Deployer (IP, AWS ARN)
   - Details about **last** Terraform run
     - plan start/finish time
     - apply/destroy start/finish time
     - variables + outputs
     - errors + warnings
 - Deployments
   - plan start/finish time
   - apply/destroy start/finish time
   - variables + outputs
   - errors + warnings
 
For full list see the [full schema](https://github.com/TimeIncOSS/ape-dev-rt/blob/master/deploymentstate/schema/schema.go).
The initial release only contains S3 backend, but future releases may support other backends, e.g. DynamoDB or Consul.
Backend may also at some point provide locking mechanism for releases to prevent an app from being deployed
by two people at the same time.

## Example

**`deployment-state.hcl.tpl`**

```hcl
deployment_state "s3" {
  region = "us-east-1"
  bucket = "ti-deployment-state-{{.Environment}}"
  prefix = "{{.AwsAccountId}}"
}
```

## Unique Application Names

Names need to be unique within a given namespace. We use **AWS Account ID as namespace** within a given backend,
so as long as you stay unique within a given AWS account you're fine.

RT does **not** check who owns a given app and neither it cares about where is the terraform code for the given app stored.

:exclamation: :exclamation: :exclamation:

**It is your responsibility to make sure you don't deploy over someone else's app with the same name in the same namespace.**

:exclamation: :exclamation: :exclamation:

----------

## Old apps in central git repository :older_man:

To avoid extra commits & diffs in [the central repository](https://github.com/TimeIncOSS/ape-dev-rt-apps) and
also to get application off that central place to their own we enforce default backend for those.

The default backend is now S3, but may change in the future - hopefully we won't have any central repo by that time.

Until we evict all apps though all the functionality is kept in place and people are still
allowed to do "commit-based" deployments.

## New apps :tada: :shipit: :tada:

Users are discouraged from using the central git repo and put the configuration into their own repositories,
possibly alongside the project and then run RT from there.

RT expects a file called `deployment-state.hcl.tpl` in the root of the directory where you keep your infrastructure.
The only reasonable content of that file right now is:

```hcl
deployment_state "s3" {
  region = "us-east-1"
  bucket = "ti-deployment-state-{{.Environment}}"
  prefix = "{{.AwsAccountId}}"
}
```

RT will support multiple deployment state backends per app which will allow us to migrate an app gracefully
to any other backend whilst keeping data in both backends (by defining 2 `deployment_state` blocks)
for a short while during the migration period.

### Override deployments

This is a recommended directory layout:

```
.
├── Gemfile
├── infra
│   ├── deployment
│   │   ├── cloud-config.yaml
│   │   ├── main.tf
│   │   └── shared-services.tf.tpl
│   ├── deployment-state.hcl.tpl
│   ├── main.tf
│   └── shared-services.tf.tpl
└── src
```

i.e. `PROJECT_ROOT/infra` contains application-specific infrastructure and nested directories contain
terraform configs for deployment, possibly 1 per slot "type".

e.g. if you need to expose a feature branch with its own infrastructure (e.g. ElastiCache, RDS instance)
you can create `PROJECT_ROOT/infra/branch-deployment` which would contain `*.tf` files with `aws_elasticache_cluster` etc.

If you rather want a separate infrastructure for branch deployments, but share that amongst all branch-based deployments
you can put more `aws_elasticache_cluster` etc. resources into `PROJECT_ROOT/infra`, expose those via outputs
and read back in `PROJECT_ROOT/infra/branch-deployment`

#### First deployment

```sh
$ cd infra/
$ ape-dev-rt -aws-profile=ti-devops-test apply-infra -env=test -app=examplestandalone
$ ape-dev-rt -aws-profile=ti-devops-test deploy \
  -env=test \
  -app=examplestandalone \
  -slot-id=default \
  -var=nginx_tag=1.11 \
  ./deployment
```

#### Ongoing deployments

You may just stick this line into you CircleCI/TravisCI/Jenkins file to build "push-to-deploy" pipeline:

```sh
$ ape-dev-rt -aws-profile=ti-devops-test deploy \
  -env=test \
  -app=examplestandalone \
  -slot-id=default \
  -var=nginx_tag=$NEW_TAG \
  ./deployment
```

where `$NEW_TAG` may be just output from anything that uploads your artefact (ZIP file / docker image / JAR / ...).

### Blue/green deployments

The initial `0.5` release doesn't provide an easy to use pattern for the same blue/green deployment yet,
but it is possible to emulate the old behaviour used in `< 0.5.0`:

```
RELEASE_NUM=$CIRCLE_BUILD_NUM

$ ape-dev-rt -aws-profile=ti-devops-test deploy \
  -env=test \
  -app=mysuperapp \
  -slot-id=release$RELEASE_NUM \
  -var=nginx_tag=$NEW_TAG \
  ./deployment
```
Previously `$RELEASE_NUM` was a git hash in the central git repository. It is up to you to choose a reliable counter
that will increment or otherwise be unique in long term.

**If you choose to deploy from your laptop, you may need to check a global counter 1st.**
Deployment state may provide such counter [in the future](https://github.com/TimeIncOSS/ape-dev-rt/issues/207), but it doesn't yet.

```sh
$ ape-dev-rt -aws-profile=ti-devops-test list-slots

release11
 - last deployed 2016-09-09 16:20:00.616220926 +0000 UTC by arn:aws:iam::187636751137:user/rsimko1016
 - last variables: map["app_name":"mysuperapp" "app_version":"release12" "environment":"test"]
 - last outputs: map["app":"mysuperapp" "aws_region":"us-east-1" "environment":"test" "version":"release12"]

release12
 - last deployed 2016-09-09 17:21:00.626220926 +0000 UTC by arn:aws:iam::187636751137:user/rsimko1016
 - last variables: map["app_name":"mysuperapp" "app_version":"release12" "environment":"test"]
 - last outputs: map["app":"mysuperapp" "aws_region":"us-east-1" "environment":"test" "version":"release12"]

$ ape-dev-rt -aws-profile=ti-devops-test enable-traffic \
  -env=test \
  -app=mysuperapp \
  -slot-id=release12

# Check that new version works fine

$ ape-dev-rt -aws-profile=ti-devops-test disable-traffic \
  -env=test \
  -app=mysuperapp \
  -slot-id=release11

# Check that app is still fine after detaching old deployment

$ ape-dev-rt -aws-profile=ti-devops-test deploy-destroy \
  -env=test \
  -app=mysuperapp \
  -slot-id=release11 \
  ./deployment
```
