
## `make release*`
# The `make release*` targets are dependent on the following environment variables:
#	UPGRADE_VERSION - ex: "0.0.87.argocd"
#	UPGRADE_WORKDIR - ex: "upgrades/0.0.87.argocd"
#	GITHUB_TOKEN={token} - only for `release`, not `release-local`
# See .github/workflows/release.yml for how these are set up and .goreleaser.yaml on how they are used

release-local:
	@echo "Building okctl upgrade: '${UPGRADE_VERSION}' from workdir=${UPGRADE_WORKDIR} locally, not pushing"
	docker run --rm --privileged \
		-v $$PWD:/go/src/github.com/oslokommune/okctl-upgrade \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/oslokommune/okctl-upgrade/${UPGRADE_WORKDIR} \
		-e UPGRADE_VERSION \
		ghcr.io/gythialy/golang-cross:v1.17.3-2 release --rm-dist --config=/go/src/github.com/oslokommune/okctl-upgrade/.goreleaser.yaml --skip-publish

release:
	@echo "Building okctl upgrade: '${UPGRADE_VERSION}' from workdir=${UPGRADE_WORKDIR}"
	docker run --rm --privileged \
		-v $$PWD:/go/src/github.com/oslokommune/okctl-upgrade \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/oslokommune/okctl-upgrade/${UPGRADE_WORKDIR} \
		-e GITHUB_TOKEN \
		-e UPGRADE_VERSION \
		ghcr.io/gythialy/golang-cross:v1.17.3-2 release --rm-dist --config=/go/src/github.com/oslokommune/okctl-upgrade/.goreleaser.yaml