# Developing
Please help us by contributing to the project, feel free to ask what needs to be done next or submit a PR!

## Get Go-ing

For starters you will need a working golang environment. [Here's a Gist](https://gist.github.com/vsouza/77e6b20520d07652ed7d) to get up and running via Homebrew on OS X.

### Go version

We expect and support Go `1.6+`. Lower versions handle vendoring differently and we rely on vendoring to work like it does in `1.6`.

## Install, Compile, Test, and Run

Assuming you've got a working golang environment.

**Install**
```
go get github.com/TimeIncOSS/ape-dev-rt
```

**Compile**
```
cd $GOPATH/src/github.com/TimeIncOSS/ape-dev-rt
make installdeps
make dev
```

**Test**
```
make test
```
There's a bunch of unit tests, keep adding those & keep existing ones up to date, please!

`make test` also runs `go vet` to keep the code readable and clean. Do this before submitting a PR, otherwise CI bot may embarrass you in front of everyone :smirk:

**Run**
```
cd $GOPATH/src/github.com/TimeIncOSS/ape-dev-rt
./bin/ape-dev-rt
```

**Dependency management**
- We currently use a custom patch of [glide](https://github.com/Masterminds/glide/pull/710) to manage dependencies as the official Glide manager currently has difficulties parsing govendor dependencies which Hashicorp uses. We also ignore the transitive dependency of [CoreOS Ignition](github.com/coreos/ignition) as it uses C source files and is not vital to the current use cases of RT.
- Versions are specified in `glide.yaml`
- When committing dependencies, split your PR into multiple commits; one with the vendored dependency and another commit with the actual code
- Most use cases are covered on the [home page of Glide](https://glide.sh)
- Common use cases:
  - To pin something to a specific version:
      - Update/add e.g. `version: v0.8.0`
      - `glide update`
  - To update all dependencies (within defined version ranges):
      - `glide update`
  - To add a new dependency:
      - `glide get github.com/user/repo/package`
  - To clear the Glide cache:
      - `glide cc`

# Debugging

If you want to see what's happening under the hood, set environment variable `RT_LOG=1` or pass `--verbose`, all `log` messages will be written out to dated files in: `~/.rt/logs/`.

# Releasing

Terraform now uses godeps, so dependencies are updated separately:
```sh
cd $GOPATH/src/github.com/TimeIncOSS/ape-dev-rt
make installdeps
```

1. Make sure you have the ti-devops-test AWS profile set as default as you'll be copying artefacts to S3 bucket in that account
2. Update the [CHANGELOG](https://github.com/TimeIncOSS/ape-dev-rt/blob/master/CHANGELOG.md)
3. Bump version in `/rt/version.go`
4. Commit the changes (commit message is fine as the new version number), e.g. `git commit -m "v0.0.6"`
5. Create tag, eg. `v0.0.6`: `git tag -a v0.0.6`
6. Generate & upload artifacts: `make release`
7. Push new tag: `git push origin master --tags`
8. Update cask in [the brewcask tap repository](https://github.com/TimeIncOSS/homebrew-cask-tap/blob/master/Casks/ape-dev-rt.rb) - specify the version number, the build script will give you the sha of the zipped file
9. Install the new version locally from the Casks repo cloned: E.g. `brew cask install ./Casks/ape-dev-rt.rb`
10. Commit and push the changes made to the Casks repo
11. Announce the release via Slack in #platform-engineering, example below

> RT `0.3.13` is available for download - update via `brew update && brew cask install ape-dev-rt`, changelog here: https://github.com/TimeIncOSS/ape-dev-rt/blob/master/CHANGELOG.md#0313-march-2-2016
