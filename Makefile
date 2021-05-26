SHELL         = bash
.PHONY: release

help: ## Print this menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

release: ## Make a release. This target assumes that running `git tag` returns the version to release.
	$(eval UPGRADE_VERSION := $(shell git describe --tags HEAD)) # goreleaser does this too

	cd ${UPGRADE_VERSION} && \
	UPGRADE_VERSION=${UPGRADE_VERSION} \
		goreleaser release \
			--config ../.goreleaser.yaml \
			--rm-dist

release-test: ## Test making a release. Example usage: make release-test UPGRADE_VERSION=0.0.50
	@if [[ -z "${UPGRADE_VERSION}" ]]; then\
		echo "You must specify UPGRADE_VERSION=<upgrade version>";\
		echo "";\
		exit 1;\
	fi;

	# Because goreleaser expects this tag, we'll read it here as well.
	cd ${UPGRADE_VERSION} && \
		goreleaser release \
			--config ../.goreleaser.yaml \
			--rm-dist --skip-publish --snapshot
