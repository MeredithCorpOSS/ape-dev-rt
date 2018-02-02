# Internals

How RT works under the hood and what was the original motivation behind it?

RT consists of the following parts:

 - [Git](#git)
 - [Go Templates](#go-templates)
 - [Terraform](#terraform)
   - [Remote State Storage](#remote-state-storage)
 - [Custom AWS functionality (`traffic` commands)](#custom-functionality-traffic-commands)

each of which is described below.

## Git

### Why?

To be able to bring up two or more different versions of an app into production
we need to be able to identify these versions somehow.

Decision was made to leverage git for this purpose.

### History

In the very early stage we experimented with [`libgit2`](https://github.com/libgit2/git2go)
but it turns out it has some [bugs](https://github.com/libgit2/git2go/issues/206)
that would significantly affect us and also libgit offers a very low-level API
which isn't very easy to work with for basic git commands.

We wanted this part of RT to be rock-solid which is why we instead decided to use
the `git` binary installed in the system.


### Risks

 - everyone having write access to the `ape-dev-rt-apps` repo could potentially mess things up for all users
 - releasing new version implies new commit which just bumps revision/version ID - that is bothersome
 - it is hard to build automation around it

### Current State

Deprecated for applications migrated out of the central repository as part of rt functionality. All Git functionality will be eventually removed.


## Go Templates

RT will process any files named `*.tf.tpl` and `deployment-state.hcl.tpl` as Go templates prior to handing such files over to Terraform.

### Why?

The main reason we use this extra templating layer is because [`deploymentstate`](https://github.com/TimeInc/ape-dev-rt/blob/master/docs/deployment-state.md) exists outside of Terraform and as a result cannot use any variable logic within Terraform.

### Outcome

 - This makes running Terraform commands against raw directories trickier because `*.tf.tpl` files
   are ignored by Terraform and otherwise do not contain HCL-valid syntax prior to template processing.
   - Both `terraform validate` & `terraform plan` would be helpful to run ahead of `git push`

### Future

`*.tf.tpl` will eventually be deprecated.


## Terraform

### Why?

We leverage [Terraform](https://github.com/hashicorp/terraform) a lot as it solves
most of the implementation details we'd otherwise have to deal with, namely

 - AWS API/SDK handling and how different resources work together generally
 - DSL and/or config parsing and how to effectively translate that into API calls

Even though our stack is and in the nearest future will be mainly AWS, we did not
want to lock the solution unnecessarily into AWS.

More importantly there will always be services that AWS doesn't provide and
we will need to automate/orchestrate those too. e.g. PagerDuty, DynDNS, Pingdom.

Even if Terraform doesn't support a specific provider, it is OSS
and it has a powerful [plugin architecture](https://www.terraform.io/docs/plugins/index.html).

### Outcomes

 - Each application and its slot needs to be described in a form of HCL template
 - Resources _can_ be also described via CloudFormation since [Terraform supports it](https://www.terraform.io/docs/providers/aws/r/cloudformation_stack.html)
 - One of the main differences from other tools that do infra provisioning (Puppet, CloudFormation, Ansible, ...)
   is that Terraform does not do resource discovery based on tags and does not expect provider to persist the state.
   This however means that it does need to manage state of each resource in a form of `tfstate` file (JSON).
   This in turn means that a team needs to share the `tfstate` file somehow. See the _"Remote State Storage"_ section below.

### Future

As opposed to many other solutions, Terraform 0.7+ will allow building applications/infrastructure
on top of existing one (e.g. VPCs that have been created completely out of Terraform and/or even before Terraform existed).

See https://github.com/hashicorp/terraform/pull/4961

### Remote State Storage

Terraform as of today (April 2016) supports 7 backends (Artifactory, Atlas, Consul, etcd, HTTP, S3, Swift).

At the time of releasing the initial version of RT (May 2015) Terraform only supported 4 (Atlas, Consul, HTTP and S3).

#### S3

A decision was made back in May 2015 to use S3 mainly for these reasons:

 - finegrained permissions management via S3 bucket policies which allows sharing tfstate between teams
 - availability of S3 (`99.99%` backed by SLA)
 - backup strategy
   - region replication
   - versioning allowing us to revert files back to past versions if an accident happens
   - number of working and supported AWS SDKs allowing creation of custom backup scripts easily
 - it does not involve extra operational overhead (i.e. the service availability is achieved by AWS engineers)
 - logging which allows auditing (i.e. who changed what and when)

#### Why (not) etcd or consul

Both solutions would introduce extra operational overhead and both would require effort to be invested into making a good backup strategy and HA deployments. This would significantly increase the scope of this project.

#### Why (not) Atlas

Atlas is a commercial offering which would introduce some periodic fees and wouldn't bring much value in this specific case (app deployments) since Atlas (as of April 2016) does not have a notion of application/slots/deployments. It is mostly aimed at ops doing purely infrastructure deployments.

When revisiting this option we should compare:
 - long-term costs
 - backup strategy
 - permissions management
 - availability & SLA

#### Why (not) Artifactory

There was no Artifactory backend supported in May 2015.

When revisiting this option we should compare:
 - long-term costs
 - backup strategy
 - permissions management
 - availability & SLA


## Custom functionality (`traffic` commands)

RT provides an extra functionality beyond Terraform. This currently includes commands for attaching/detaching
ELBs to/from Autoscaling Groups.

### Why

We distinguish between two contexts/layers (application and slot). These are typically enough to build 
the whole infrastructure for each application and its slots.

Sometimes changes need to be done out of these contexts though - i.e. the state of a slot may change
since slot may or may not be accepting traffic (i.e. ELB may or may not be attached to the ASG).

### Outcomes

 - RT allows users to perform certain actions on top of the infrastructure that was built via Terraform
 - Such actions typically don't conflict with Terraform's state thanks to `lifecycle`'s `ignore_updates` flag.
 - If a user wants to be able to perform changes beyond what Terraform has built, the HCL templates need to follow
   certain conventions and may need to define `lifecycle.ignore_updates` for certain resources.

### Risks

 - Since most of the AWS functionality is handled by Terraform and treated as implementation details
   these custom functions re-introduce some problems:
   - AWS region handling
   - AWS API credentials handling
   - identification/discovery of resources built by Terraform
   - AWS API throttling


------------

> to be moved under "RT expectations" per https://trello.com/c/KCM4pgza/85-allow-infra-code-to-be-in-any-repository

## Versioning of apps/configs

Applications in the Release Tool are bicameral as we distinguish between the "infra" and "slot" configuration details.

### Concepts

**Infra configuration** describes resources whose lifecycle match that of the whole application. There is a single "infra configuration" for each app and changes are **incremental** - they can be modified and updated like a normal Terraform configuration.

**Slot configuration** describes resources whose lifecycle match that of a single application slot. There will be multiple slot configurations for a single app. Each slot represent a repository snapshot, so slot changes are **immutable** - slots may only be created or destroyed.

