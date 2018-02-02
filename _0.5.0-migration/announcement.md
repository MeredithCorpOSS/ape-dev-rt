Hi all,

RT `0.5.0` was silently released.

## Good news

It contains some new cool features, highlights include commit-less deployment, ability to recreate resources via `taint`. See full changelog at https://github.com/TimeInc/ape-dev-rt/blob/master/CHANGELOG.md#050-september-12-2016

## Bad news

RT `0.5.0` stores **metadata** around each deployment & app in a separate storage (S3) for various reasons
(e.g. to allow smooth RT/Terraform upgrade per app or to speed up all `list-*` and other commands to a couple of seconds).

We want to avoid "split-brain" situations where people would use old RT that doesn't save metadata and new one which does for the same app.
Even though we have some basic protection in place, I'd like to get all applications in all accounts to `0.5.0` soon.
The migration requires a bit of cooperation as the team operating the app knows best whether the app works fine after fresh deployment.

Here are some details about the deployment state (metadata) we save alongside each app in `0.5.0`:
https://github.com/TimeInc/ape-dev-rt/blob/master/docs/deployment-state.md

### Up To Date Users

Generally speaking if you now use the last RT before `0.5.0` (`0.4.7`) which bundles Terraform `0.6.16`
it is highly unlikely you will run into any problems during/after this migration.

### Naughty Users

I'm aware of people using some ancient old versions of RT which bundle old version of Terraform
and to address this **I will be looking at your Terraform configs in https://github.com/TimeInc/ape-dev-rt-apps
in the following weeks** and possibly **send you PRs** to address any incompatibilities w/ TF `0.6.16`.

## Migration plan - Your Cooperation needed

After Terraform `0.6.16` incompatibilities are addressed I will book a short session
(length depends on the number of your apps) with each team
to do the migration to RT `0.5.0` and **lock out `< 0.5.0` users**.

Here is the migration plan divided into phases:
https://github.com/TimeInc/ape-dev-rt/blob/master/_0.5.0-migration/migration-plan.md

To make this easy I'd appreciate if you wouldn't create any new apps
that are incompatible with Terraform 0.6.16 (i.e. RT 0.4.7).

------

Sorry for any inconvenience caused by this,
I hope you'll understand the reasons and enjoy the new release of RT.
