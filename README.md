This repository contains functionality for upgrading [okctl](https://github.com/oslokommune/okctl).

# Developing principles

## Use required flags

Upgrades [MUST](https://www.ietf.org/rfc/rfc2119.txt) support the flags described in [this code](template/main.go). The behavior
of the flags MUST be as described in the code's comments.

## Return correct exit code

Upgrades MUST return a non-zero exit code if it doesn't complete successfully for any reason. For instance if the user is prompted
whether to continue and answers no, the upgrade must return a non-zero exit code.

## Support idempotency

Upgrades MUST be idempotent. How to implement is this is up to the update itself, but the straight forward way is using the same
logic as okctl uses to check if an upgrade has been run, which is checking the cluster's state storage.

## Avoid re-runs

If you want to make sure okctl upgrade doesn't re-run an upgrade that has been run manually, that is, outside of
`okctl upgrade`, then you MUST ensure that the upgrade updates okctl's state, marking the upgrade as run.

## Avoid cross-upgrade imports

Any code in an upgrade MUST NOT import code from another upgrade.

Upgrading is a complex domain, and to ensure that upgrades stay simple and isolated, they must avoid such imports. If downstream
migrations depend on an upstream migrations, we might break any of the downstream migrations, which we want to avoid.

If you want to reuse logic, either duplicate it or import it from somewhere common outside the migrations. However, make sure
changes to reuse logic doesn't break any of the migrations using the common logic (by having tests, keeping the API stable, etc.).

# How to create an upgrade

This section describes a workflow for developing, testing and releasing an upgrade. 

## Create the upgrade

* `cp -r template <okctl target version>`, where `<okctl target version>` is explained under [Upgrade binaries](#upgrade-binaries).
    * Example: `cp -r template 0.0.60.some-component`
* Edit the upgrade to your needs (:information_source: Tip: start with `upgrade.go`)

## Test the upgrade

### Test continuously while developing

You need two things for running the upgrade directly (i.e. not using `okctl upgrade`).

* Environment variables

Every upgrade must run with the same environment as `okctl show credentials` (which equals `okctl venv`).

So if running through an IDE such as IntelliJ, add environmental variables to be at least those from `okctl show credentials`.

* Okctl state

```shell
okctl venv -c cluster.yaml
okctl maintenance state-acquire-lock # Optional

# Replace CLUSTER_NAME with cluster name (found in cluster.yaml)
okctl maintenance state-download -p ~/.okctl/localState/CLUSTER_NAME/state.db
```

Now you can run the upgrade.

Then upload the state:

```shell
# Replace CLUSTER_NAME with cluster name (found in cluster.yaml)
okctl maintenance state-download -p ~/.okctl/localState/CLUSTER_NAME/state.db
okctl maintenance state-release-lock
```



### End-to-end test using `okctl upgrade`

A more thorough and better test, is using `okctl upgrade`. 

`okctl upgrade` fetches upgrades from releases in this repository. To test that an upgrade works before releasing, we can release
the upgrade in a mirror repository used for testing: [okctl-upgrade-test](https://github.com/oslokommune/okctl-upgrade-test).
We then okctl to use that one.

To do so, follow the following steps.

### Create a release in the test repository

In this repository, run

```shell
./test-release.sh TAG 
```

where TAG is the tag you want to release with. See [release the upgrade](#release-the-upgrade) for details.

Example

```shell
./test-release.sh 0.0.80+some-component 
```

### Run the test upgrade

In okctl repository [pkg/upgrade/upgrade.go](https://github.com/oslokommune/okctl/blob/master/pkg/upgrade/upgrade.go#L30), set the constant `OkctlUpgradeRepo` so it becomes

```
OkctlUpgradeRepo = "okctl-upgrade-test"
```

Then build okctl and run

```shell
# Remember to use the okctl binary you just built
/path/to/local-built/okctl -c cluster.yaml upgrade 
```

and see that it works as expected.

#### Enable re-run

`okctl upgrade` will only run the upgrade once. So to be able to run it again this way, you need to reset the Okctl state. You
can do it by running:

```shell
cd your-iac-repository
okctl maintenance state-acquire-lock # Optional
okctl maintenance state-download
boltbrowser state.db
```

In Boltbrowser, go to upgrade -> Upgrade

and delete the entry for your upgrade. (Use `D` to delete it.)

Then upload the state

```shell
okctl maintenance state-upload state.db
okctl maintenance state-release-lock
```

## Release the upgrade 

To make the actual release, first push the upgrade to the main branch (through a PR, preferrably). Then run the following:

```shell
TAG="0.0.5+some-component" # GitHub actions will then look for the dir 0.0.5.some-component
git checkout main && git pull
git tag -s "$TAG" -m "Upgrade $TAG"
git push --atomic origin main $TAG
```

GitHub actions takes care of the rest.

## There is an error with my upgrade, what do I do

If you after releasing an upgrade discover problems with the upgrade, you can:

* delete release and the tag corresponding tag in GitHub
* update the upgrade code, and create a new release as described by the steps above.

Note: Update existing released upgrades with care, as some users may have already executed them.

# Implementation details

This section describes inner workings of how Okctl upgrades in the context of this repository work. 

The GitHub release action will produce the outputs described below.

## Upgrade binaries

This project creates binaries whose file name will follow this naming convention:

```
okctl-upgrade_<okctl target version>_<os>_<arch>
```

* `okctl target version` is the version of okctl the upgrade should upgrade to.

For instance, after running upgrades with target version `0.0.5` and `0.0.6`, the okctl infrastructure should be as if creating
the infrastructure with okctl version `0.0.6`.

The version format can be a semantic version (e.g. `0.0.10`), or a semantic version with an identifier (`0.0.10.some-identifier`).
If using an identifier, it MUST be on the format `<semver>.<identifier>` where `<identifier>` MUST NOT contain dots or underscores.

The intention of `identifier` is to be a human readable string that communicates what is being upgraded, for instance `loki`
or `argocd`. So an identifier SHOULD be present.

TL;DR, some examples:

```shell
0.0.64.argocd
0.0.65.loki-pvc
0.0.66 # valid, but not preferable as it's missing an identifier 
```

* `os` is `Linux` or `Darwin`
* `arch` is for instance `amd64`

Examples:

* `okctl-upgrade_0.0.63_Linux_amd64`
* `okctl-upgrade_0.0.64_Linux_amd64`
* `okctl-upgrade_0.0.64_Darwin_amd64`
* `okctl-upgrade_0.0.65_Linux_amd64`
* `okctl-upgrade_0.0.65.loki-persistence_Linux_amd64`

## Releases

Every binary will be put in its own release, and tagged with the `<okctl target version>`.

See https://github.com/oslokommune/okctl-upgrade/releases.
