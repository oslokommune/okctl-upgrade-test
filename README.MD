This repository contains functionality for upgrading [okctl](https://github.com/oslokommune/okctl).

# Developing principles

* Upgrades MUST support `-f` and `--force`. That is, when executing an upgrade binary without these arguments, it MUST
  make zero changes. Only when run with `-f` or `--force` must the upgrade make any changes.
* Upgrades MUST be idempotent. How to implement is this is up to the update itself, but the straight forward way is
  using the same logic as okctl uses to check if an upgrade has been run, which is checking the cluster's state storage. 
* Any code in a migration [MUST NOT](https://www.ietf.org/rfc/rfc2119.txt) import code from another migration.
  Upgrading is a complex domain, and this is to ensure that upgrades stays simple and isolated. If downstream migrations
  depend on an upstream migrations, we might break any of the downstream migrations, which we want to avoid. If you want
  to reuse logic, either duplicate it or import it from somewhere common outside the migrations. However, make sure
  changes to reuse logic doesn't break any of the migrations using the common logic (by having tests, keeping the API
  stable, etc).
* If you want to make sure okctl upgrade doesn't re-run an upgrade that has been run manually, that is, outside of
  `okctl upgrade`, then you MUST ensure that the upgrade updates okctl's state, marking the upgrade as run.

# How to create an upgrade

* `cp -r template <okctl target version>`, where `<okctl target version>` is explained under [Binaries](#binaries).
   * Example: `cp -r template 0.0.60`
* Optional: To have a look at the release before publishing, run `make release-test UPGRADE_VERSION=0.0.65`, which will
create a `dist` directory.
* To make the actual release, do:

```shell
TAG="0.0.5" # Or 0.0.6.some-hotfix
git add $TAG
git commit -m "✔ Add upgrade $TAG"
git tag -s "$TAG" -m "Upgrade $TAG"
git push --atomic origin main $TAG
```

Github actions takes care of the rest.

# Outputs

The github release action will produce the outputs described below.

## Binaries

This project creates binaries whose file name will follow this naming convention:

```
okctl-upgrade_<okctl target version>_<os>_<arch>
```

* `okctl target version` is the version of okctl the upgrade should upgrade to.

For instance, after running upgrades with target version `0.0.5` and `0.0.6`, the okctl infrastructure should be as if
creating the infrastructure with okctl version `0.0.6`.
 
The version format can be a semantic version (e.g. `0.0.10`), or a semantic version with a hotfix (`0.0.10.some-hotfix`). If
using a hotfix, it MUST be on the format `<semver>.<hotfix identifier>` where `<hotfix identifier>` MUST NOT contain
dots or underscores.

TL;DR, some examples:

```shell
0.0.63
0.0.64
0.0.65.some-hotfix
```

  * `os` is `Linux` or `Darwin`
  * `arch` is for instance `amd64`

Examples:
* `okctl-upgrade_0.0.63_Linux_amd64`
* `okctl-upgrade_0.0.64_Linux_amd64`
* `okctl-upgrade_0.0.64_Darwin_amd64`
* `okctl-upgrade_0.0.65_Linux_amd64`
* `okctl-upgrade_0.0.65.some-hotfix_Linux_amd64`

## Releases

Every binary will be put in its own release, and tagged with the `<okctl target version>`.

See https://github.com/oslokommune/okctl-upgrade/releases.
