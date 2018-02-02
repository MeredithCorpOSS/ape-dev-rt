## 0.8.0 (Unreleased)

BREAKING CHANGES:

IMPROVEMENTS:

## 0.7.0 (December 16, 2016)

BREAKING CHANGES:
 - Previously deprecated flags (`-version` & `-legacy-behaviour`) and deprecated commands (`*-app` & `*-version`) removed [GH-271]
 - Deprecated logic that previously allowed you to work with apps in the central repository is now gone.
   This version does NOT perform any git operations and **will not work with apps that are in the central repo**. :wave:
 - `remote_state` block is now required, see https://github.com/TimeInc/ape-dev-rt/blob/master/docs/remote_state.md for more

IMPROVEMENTS:
 - Added check for active slots to destroy-infra [GH-270]
 - Increased Clippy's bubble size restriction [GH-273]
 - Added target flag to apply and destroy commands [GH-274]
 - Added variables flag for deploy-destroy and destroy-infra [GH-276]
 - Added namespace flag that allows users to chose a custom namespace or default to AWS Account Id [GH-279]

BUG FIXES:
- Apply and destrpy commands can now be carried out regardless of a resource change using force flag [GH-275]

## 0.6.3 (December 9, 2016)

IMPROVEMENTS:
 - more precise error message for S3 Access Denied [GH-263]
 - aws: More specific error message when wrong credentials provided [GH-264]
 - Disallow disable-traffic if it's last slot w/ traffic [GH-266]
 - Added delete slot prefixes feature [GH-267]
 - Added output and slot-output commands [GH-269]

BUG FIXES:
 - Error out gracefully if no instances found [GH-262]

## 0.6.2 (November 3, 2016)

IMPROVEMENTS:
 - Expose application name (`-app` flag) as `AppName` variable in `*.tf.tpl` in infrastructure level [GH-254]
 - Allow hyphens in app name [GH-255]

BUG FIXES:


## 0.6.1 (October 24, 2016)

IMPROVEMENTS:
 - Ability to use an ec2 instance profile for authentication [GH-253]


## 0.6.0 (October 18, 2016)

BREAKING CHANGES:

 - Terraform upgraded to `0.7.6` (from `0.6.16`).
   Please read the official [Terraform Upgrade Guide](https://www.terraform.io/upgrade-guides/0-7.html) and feel free to use [upgrade shell script](https://github.com/TimeInc/ape-dev-rt/blob/master/_0.6.0/tf-upgrade.sh) before upgrading RT to this version. It is very likely you will need to change some of your terraform configs (`.tf`).


## 0.5.3 (October 13, 2016)

IMPROVEMENTS:

 - deploymenstate: Support slot counter for blue/green deployments [GH-227]
 - cmd/show-traffic: Show private IPs of active instances [GH-249]
 - commands: Add `cleanup-slots` for cleaning up old unused slots [GH-248]
 - Allow dots and hyphens in slot-id [GH-247]

BUG FIXES:

 - Report exit code 1 in case of error [GH-240]
 - Allow saving app details even if TF fails [GH-237]
 - Improve error message about missing config [GH-241]
 - Show slot deployments in progress correctly [GH-235]
 - Make tainting work for out-of-bound deployments [GH-242]

## 0.5.2 (September 20, 2016)

IMPROVEMENTS:

 * cmd/`apply-infra`: Allow passing Terraform vars via flag [GH-229]
 * command: Clean up temp files after successful %operation% [GH-230]
 * commons: Enforce assumptions about infra directory [GH-222]
 * clippy: Expand bubble to 4 lines [GH-232]

BUG FIXES:

 * cmd/`apply-infra`: Display stderr for exit code != 0 [GH-217]

## 0.5.1 (September 13, 2016)

BUG FIXES:

 * cmd/`list-deployments`: Ignores terraform outputs [GH-211]
 * cmd/`deploy`: Discard stderr stream [GH-212]
 * cmd: Discard stderr in remaining commands [GH-213]

## 0.5.0 (September 12, 2016)

DEPRECATIONS / BREAKING CHANGES:

  * As we need all apps to report deployment metadata we'll enforce all apps to upgrade per the migration plan
    which may break apps due to the fact that we bundle 0.6.16 and some people may have been using older RT versions
    which bundled older versions of Terraform.
  * New/deprecated commands:
    * `apply-app` => `apply-infra`
    * `destroy-app` => `destroy-infra`
    * `apply-version` => `deploy`
    * `destroy-version` => `deploy-destroy`
    * `list-versions` => `list-slots` + `list-deployments`
  * `list-apps` does not pull data from central git repository anymore. It instead pulls metadata from S3 (default `deploymentstate` backend - see below). [GH-202]

FEATURES:

  * **Deployment Data** are now persisted in predefined backend [GH-202]
  * **Override deployments** (aka non-blue/green) are allowed as long as underlying TF configs allow it [GH-202]
  * New `-var` flag introduced for `deploy` command to allow users passing arbitrary variables per deployment and persist those per slot. [GH-202]

IMPROVEMENTS:

  * Ability to taint and untaint resources created by an application through the `taint-infra-resource` and `untaint-infra-resource` commands [GH-125]
  * Ability to taint and untaint deployed resources via `taint-deployed-resource` & `untaint-deployed-resource` [GH-193]
  * All `terraform apply` operations now use generated plan which reduces a chance of conflict between 2+ team members deploying at the same time. `apply-app`/`apply-version` should both error out if something has changed between plan generation & confirmation. [GH-202]
  * As RT now saves version used to do a deployment of a given app, it will prevent you from using an older version of RT
    if newer RT version has touched a given app. RT will error out and you'll be forced to upgrade. [GH-202]

BUG FIXES:

  * Clippy prompt could in some cases hide whole line of text in the prompt bubble [GH-201]
  * `shared-services` is now treated as an invalid application name [GH-203]

## 0.4.7 (June 20, 2016)

IMPROVEMENTS:

  * Display details about authenticated user/role for all `apply-*` and `destroy-*` commands [GH-117]

BUG FIXES:

  * `*-traffic` commands now perform more precise discovery of ASGs (i.e. RT now works corrently if 1 git hash is changing 2 or more applications) [GH-114]

## 0.4.6 (May 10, 2016)

IMPROVEMENTS:

  * Terraform bumped to `0.6.16` - see [upstream Changelog](https://github.com/hashicorp/terraform/blob/master/CHANGELOG.md#0616-may-9-2016) to see bugfixes & features which are now also part of RT

## 0.4.5 (April 13, 2016)

BUG FIXES:

  * Version state file is now being cleaned up in all possible cases on `destroy-version` which should speed up `list-versions` in long term [GH-111]

## 0.4.4 (April 5th, 2016)

IMPROVEMENTS:

  * `*-traffic` commands accept an `--aws-region` argument to specify where the app is deployed
  * Add support for `null` provider + `local-exec` provisioner [GH-108]
  * New error when `apply-version` is run before `apply-app`

## 0.4.3 (March 28th, 2016)

IMPROVEMENTS:

  * `-y` flag confirms interactive prompts automatically [GH-107]

## 0.4.2 (March 22nd, 2016)

BUG FIXES:

  * Elastic Load Balancers are now discovered based upon the internal app name [GH-105]

BREAKING CHANGES:

  * `--version` flag was removed to preserve a single way to get version (`version` subcommand) [GH-103]
  * `-v` flag (previously used for printing version) has been repurposed as an alias to `--verbose` [GH-103]
  * Alias subcommands `apply`, `destroy`, `init`, `list` have been removed [GH-103]

## 0.4.1 (March 11th, 2016)

BUG FIXES:

  * `tfstate` file is only removed from S3 on `destroy-version` if user actually confirms dialog [GH-100]
  * ELBs discovery by tags matches both underscore and hyphen delimited app names [GH-104]

## 0.4.0 (March 10th, 2016)

IMPROVEMENTS:

  * `show-traffic` to view load balancers connected to version ASGs [GH-95]
  * `enable-traffic` to attach app load balancers to version ASGs [GH-95]
  * `disable-traffic` to detach app load balancers from version ASGs [GH-95]

## 0.3.14 (March 9th, 2016)

IMPROVEMENTS:

  * Validation errors (e.g. when `env` is missing or has wrong format) made more human-compatible [GH-97]

BUG FIXES:
  * `create-app` would fail saying `[ERROR] aws-profile cannot be blank`
  * `destroy-version` will now throw an error if you try passing `-version` which doesn't belong to `-app` [GH-96]

INTERNAL CHANGES:

  * Logging was added to help troubleshooting any troubles around terraform integration and Clippy prompt [GH-99]
  * RT is built and tested against Go 1.6 (previously 1.5) [GH-98]

## 0.3.13 (March 2, 2016)

BUG FIXES:

  * `apply-version` without explicit `-version` won't throw an error (`master` is now acceptable option) [GH-94]

## 0.3.12 (February 24, 2016)

BUG FIXES:

  * `apply-version` will now throw an error if you try passing `-version` which doesn't belong to `-app` [GH-93]
  * core: When running `git log` in order to find the last commit in `list-apps`
    we were historically asking for all commits and throwing away all except the last one.
    We now ask for the last one only which should make `list-apps` slightly faster. [GH-93]

## 0.3.11 (January 26, 2016)

BUG FIXES:

  * `tfstate` files are now deleted from S3 after you run `destroy-version` which should speed up `list-versions` [GH-91]
