#!/usr/bin/env bash

SAMPLE_TAG="0.0.81+argocd"
TEST_REPO_NAME="okctl-upgrade-test"
TEST_REPO_FULL_NAME="oslokommune/$TEST_REPO_NAME"

set -e
if [[ $* == "-h" ]]; then
    ME=$(basename $0)
    echo "This will create a release in the $TEST_REPO_NAME repository from the current branch."
    echo "Requirements:"
    echo "* You must be checked out on the branch you want to release."
    echo "* You must have pushed that branch."
    echo "* You must have set a GITHUB_TOKEN"
    echo "* You must have the 'gh' utility installed"
    echo
    echo "USAGE:"
    echo "$ME <TAG>"
    echo
    echo "TAG must be a non-existent tag locally and remotely"
    echo
    echo "EXAMPLE:"
    echo "$ME 0.0.81+argocd"
    echo
    exit 0
fi

function panic() {
  echo Error: $1
  exit 1
}

TAG=$1

if [[ -z $TAG ]]; then
  echo Missing: TAG
  echo See -h for usage
  exit 1
fi

if [[ ! "$TAG" == *"+"* ]]; then
  echo "TAG does not contain '+', are you sure your TAG is correct?"
  echo "Sample TAG: $SAMPLE_TAG"
  exit 1
fi

if [[ ! $(git status -s | wc -l) -eq 0 ]]; then
    echo "Git status dirty, commit before continuing."
    exit 1
fi

FEATURE_BRANCH=$(git branch --show-current)

git fetch
if [[ ! $(git diff "origin/$FEATURE_BRANCH" --numstat | wc -l) -eq 0 ]]; then
  echo "You have unpushed commits. Push before continuing."
  exit 1
fi

export UPGRADE_VERSION=${TAG/\+/.}
TEST_REPO_DIR="/tmp/$TEST_REPO_NAME"

echo "UPGRADE_VERSION: $UPGRADE_VERSION"

if [[ -d "$TEST_REPO_DIR" ]]; then
  rm -rf $TEST_REPO_DIR
fi

echo "----------------------------------------------------------------------------"
echo
echo "Preparing $TEST_REPO_FULL_NAME repository..."
echo

git clone "git@github.com:$TEST_REPO_FULL_NAME.git" $TEST_REPO_DIR
cd "$TEST_REPO_DIR"

git remote add okctl-upgrade git@github.com:oslokommune/okctl-upgrade.git
git fetch okctl-upgrade
git checkout -b "$FEATURE_BRANCH" --track "okctl-upgrade/$FEATURE_BRANCH"
git remote rm okctl-upgrade # to make sure we do changes in test repository, not original one

echo
echo Delete local and remote tag if exists...
echo
(git tag -d "$TAG" 2>/dev/null) || true
(git push --delete origin "$TAG" 2>/dev/null) || true

# shellcheck disable=SC2164
cd "$TEST_REPO_DIR/$UPGRADE_VERSION"

echo "----------------------------------------------------------------------------"
echo
echo "Deleting previous release(s) if any..."
echo

while [ 1 ]
do
  echo "Finding colliding releases..."
  RELEASE_TAGS=$(gh release list -R "$TEST_REPO_FULL_NAME" | cut -f 3 | (grep "$TAG" || true))

  if [[ ${#RELEASE_TAGS} -gt 0 ]]; then
    echo "Found colliding release with tags:"
    echo "$RELEASE_TAGS"
    echo "Deleting release with tag: $TAG"
    echo Running: gh release delete -R $TEST_REPO_FULL_NAME "$TAG" -y
    gh release delete -R $TEST_REPO_FULL_NAME "$TAG" -y
  else
    echo "No more colliding release tags found."
    break
  fi
done

echo "----------------------------------------------------------------------------"

echo
echo Disabling release workflow for this branch...
echo
RELEASE_FILE="$TEST_REPO_DIR/.github/workflows/release.yml"

git checkout "$FEATURE_BRANCH"
rm "$RELEASE_FILE"
git add "$RELEASE_FILE"
git commit -m "Don't release on push, we're doing it locally"

git tag -s "$TAG" -m "Upgrade $TAG"

echo
echo Pushing tag...
git push --atomic origin "$TAG"


echo "----------------------------------------------------------------------------"
echo
echo "Creating release... Running from dir: $(pwd)"
echo
goreleaser release \
  --config ../.goreleaser.yaml \
  --rm-dist

rm -rf "$TEST_REPO_DIR"

echo "----------------------------------------------------------------------------"
echo
echo "Done. Now open https://github.com/$TEST_REPO_FULL_NAME/releases and enjoy your new release."
echo
