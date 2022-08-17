#!/usr/bin/env bash
EKSCTL_VERSION="v0.104.0"

#
# Functions
#
function run_with_output() {
  run_cmd true "$@"
}

function run_no_output() {
  run_cmd false "$@"
}

function run_cmd() {
  local GET_OUTPUT=$1
  # shellcheck disable=SC2124
  local CMD="${@:2}"
  local TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")

  echo "" >&2
  echo "~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~" >&2
  echo -e "Running command [$TIMESTAMP]: \e[96m${CMD}\e[0m" >&2
  echo "~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~" >&2

  local RESULT=""
  if [[ GET_OUTPUT == "true" ]]; then
    RESULT=$($CMD)
  else
    $CMD
  fi

  local ERROR_CODE=$?

  if [[ ! $ERROR_CODE == 0 ]]; then
    echo Command failed with error code $ERROR_CODE: "$CMD" >&2
    echo Aborting. >&2

    exit $ERROR_CODE
  fi

  if [[ $GET_OUTPUT == "true" ]]; then
    echo "$RESULT" >&2
    echo "~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~" >&2
  fi

  echo "$RESULT"
}

function require_installed_cmd() {
  local CMD=$1
  if ! command -v $CMD &> /dev/null
  then
      echo "Executable '$CMD' could not be found. Install before retrying."
      exit 1
  fi
}

function require_version_match() {
  local EXPECTED_VERSION=$1
  local ACTUAL_VERSION=$2
  local BINARY=$3

  if [[ "$EXPECTED_VERSION" != "$ACTUAL_VERSION" ]]; then
    echo "$BINARY version does not return expected version."
    echo "Expected: $EXPECTED_VERSION"
    echo "Actual:   $ACTUAL_VERSION"
    exit 1
  fi
}

# https://kubernetes.io/releases/
# https://github.com/kubernetes/kubernetes/releases
function get_kubectl_version() {
  local EKS_TARGET_VERSION=$1
  if [[ $EKS_TARGET_VERSION == "1.20" ]]; then
    echo "v1.20.15"
  elif [[ $EKS_TARGET_VERSION == "1.21" ]]; then
    echo "v1.21.14"
  elif [[ $EKS_TARGET_VERSION == "1.22" ]]; then
    echo "v1.22.11"
  elif [[ $EKS_TARGET_VERSION == "1.23" ]]; then
    echo "v1.23.8"
  elif [[ $EKS_TARGET_VERSION == "1.24" ]]; then
    echo "v1.24.2"
  else
    echo "I don't know about any matching kubectl version for EKS version '$EKS_TARGET_VERSION'. Update this script."
    exit 1
  fi
}

#
# Args check
#
if [[ $* == "-h" || -z "$1" || -z "$2" || -z "$3" || "$#" -lt 3 ]]
then
    ME=$(basename $0)
    echo -e "\e[1mUSAGE:\e[0m"
    echo "$ME <cluster-manifest file> <aws-region> <EKS target version> [dry-run={false|true}]"
    echo
    echo "cluster-manifest file      The Okctl cluster manifest"
    echo "aws-region                 AWS region"
    echo "EKS target version         Example: 1.21"
    echo "dry-run                     Default true. Set to false to actually run upgrade."
    echo
    echo -e "\e[1mEXAMPLES:\e[0m"
    echo "# Run with dry-run, i.e. do no changes, i.e. it's safe to run:"
    echo "$ME cluster-dev.yaml eu-west-1 1.21"
    echo
    echo "# Run with dry-run false, i.e. actually run the upgrade:"
    echo "$ME cluster-dev.yaml eu-west-1 1.21 dry-run=false"
    echo
    exit 0
fi

#
# Get and validate all input variables
#
CLUSTER_MANIFEST="$1"
export AWS_REGION="$2"
DRY_RUN=true

if [[ ! -f "$CLUSTER_MANIFEST" ]]; then
  echo "File does not exist: $CLUSTER_MANIFEST"
  exit 1
fi

NODEGROUP_FILE="/tmp/nodegroup_config.yaml"
if [[ -f $NODEGROUP_FILE ]]; then
  rm /tmp/nodegroup_config.yaml
fi

if [[ $4 == "dry-run=false" ]]; then
  DRY_RUN=false
fi

EKS_TARGET_VERSION=$3
if [[ ! "$EKS_TARGET_VERSION" =~ ^1\.[0-9]{2}$ ]]; then
  echo "Target EKS verison must match regex '^1\.[0-9]{2}$'"
  echo "You used: $EKS_TARGET_VERSION"
  exit 1
fi

EKS_VERSION_WITH_DASH=${EKS_TARGET_VERSION//\./-} # Convert 1.22 to 1-22
TARGET_BINARY_DIR=/tmp/eks-upgrade/$EKS_VERSION_WITH_DASH
KUBECTL_VERSION=$(get_kubectl_version "$EKS_TARGET_VERSION")

CLUSTER_NAME=$(yq e '.metadata.name' "$CLUSTER_MANIFEST")
AWS_ACCOUNT=$(yq e '.metadata.accountID' "$CLUSTER_MANIFEST")

#
# Test dependencies
#
require_installed_cmd yq
require_installed_cmd jq
require_installed_cmd aws

echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Verify AWS account"
echo "------------------------------------------------------------------------------------------------------------------------"
LOGGED_IN_AWS_ACCOUNT=$(run_with_output aws sts get-caller-identity | jq -r '.Account')
if [[ -z $LOGGED_IN_AWS_ACCOUNT || "$LOGGED_IN_AWS_ACCOUNT" == "" ]]; then
    echo
    echo "Error: You are not logged in."
    echo
    echo "Solution:"
    echo "- You need to log in. Details: https://www.okctl.io/authenticating-to-aws/#aws-single-sign-on-sso"
    exit 1
fi

if [[ "$LOGGED_IN_AWS_ACCOUNT" != "$AWS_ACCOUNT" ]]; then
  echo
  echo "Error: Logged in AWS account '$LOGGED_IN_AWS_ACCOUNT' does not match AWS account in $CLUSTER_MANIFEST '$AWS_ACCOUNT'."
  echo
  echo "Cause:"
  echo "- You must be logged in to the same AWS account as specifyed in $CLUSTER_MANIFEST."
  echo
  echo "Solution:"
  echo "- Run 'aws sso login' and set correct AWS_PROFILE."
  echo "  For details, see: https://www.okctl.io/authenticating-to-aws/#aws-single-sign-on-sso"
  exit 1
fi

echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Download dependencies to $TARGET_BINARY_DIR"
echo "------------------------------------------------------------------------------------------------------------------------"
mkdir -p "$TARGET_BINARY_DIR"

case "$(uname -s)" in
   Darwin)
     EKSCTL_URL="https://github.com/weaveworks/eksctl/releases/download/${EKSCTL_VERSION}/eksctl_Darwin_amd64.tar.gz"
     KUBECTL_URL="https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/darwin/amd64/kubectl"
     ;;

   Linux)
     EKSCTL_URL="https://github.com/weaveworks/eksctl/releases/download/${EKSCTL_VERSION}/eksctl_Linux_amd64.tar.gz"
     KUBECTL_URL="https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
     ;;

   *)
     echo 'Not using macOS or Linux, aborting'
     exit 1
     ;;
esac

# run_with_output doesn't handle big outputs so well, so running command directly instead
echo -e "Running: \e[96mcurl --location  $EKSCTL_URL | tar xz -C  $TARGET_BINARY_DIR\e[0m"
                        curl --location "$EKSCTL_URL"  | tar xz -C "$TARGET_BINARY_DIR"
ERROR_CODE=$?
if [[ ! $ERROR_CODE == 0 ]]; then
  echo eksctl curl failed
  exit $ERROR_CODE
fi


echo -e "Running: \e[96mcurl --location  $KUBECTL_URL  -o  $TARGET_BINARY_DIR/kubectl\e[0m"
                        curl --location "$KUBECTL_URL" -o "$TARGET_BINARY_DIR/kubectl"
ERROR_CODE=$?
if [[ ! $ERROR_CODE == 0 ]]; then
  echo kubectl curl failed
  exit $ERROR_CODE
fi

run_with_output chmod +x "$TARGET_BINARY_DIR/eksctl"
run_with_output chmod +x "$TARGET_BINARY_DIR/kubectl"

EKSCTL=$TARGET_BINARY_DIR/eksctl
KUBECTL=$TARGET_BINARY_DIR/kubectl

# Verify that binaries works and return correct version
EKSCTL_VERSION_ACTUAL=$(run_with_output "$EKSCTL" version -o json | jq -r '.Version')
require_version_match "$EKSCTL_VERSION" "v${EKSCTL_VERSION_ACTUAL}"

KUBECTL_VERSION_ACTUAL=$(run_with_output "$KUBECTL" version --client=true --output='yaml' | yq e '.clientVersion.gitVersion')
require_version_match "$KUBECTL_VERSION" "$KUBECTL_VERSION_ACTUAL"

echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Verify cluster name"
echo "------------------------------------------------------------------------------------------------------------------------"
run_with_output "$EKSCTL" get cluster "$CLUSTER_NAME"

echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Do these variables look okay?"
echo "------------------------------------------------------------------------------------------------------------------------"

echo -e "Upgrading EKS to version: \e[93m${EKS_TARGET_VERSION}\e[0m"
echo -e "Cluster manifest: \e[93m${CLUSTER_MANIFEST}\e[0m"
echo -e "Cluster name: \e[93m${CLUSTER_NAME}\e[0m"
echo -e "AWS account: \e[93m${AWS_ACCOUNT}\e[0m"
echo -e "AWS region: \e[93m${AWS_REGION}\e[0m"
echo -e "Dry run: \e[93m${DRY_RUN}\e[0m"

echo
# shellcheck disable=SC2162
read -n1 -p "Do these variables look okay? (Y/n) " confirm
if ! echo "$confirm" | grep '^[Yy]\?$'; then
  echo
  echo "Aborting."
  exit 1
fi
echo

# Instructions: https://eksctl.io/usage/cluster-upgrade/
echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Run upgrade of EKS control plane. Estimated time: 10-15 min."
echo "------------------------------------------------------------------------------------------------------------------------"
echo "💡 Tip: You can go to EKS in AWS console to see the status is set to 'Updating'."

if [[ $DRY_RUN == "false" ]]; then
  run_with_output "$EKSCTL" upgrade cluster --name "$CLUSTER_NAME" --version "$EKS_TARGET_VERSION" --approve
else
  run_with_output "$EKSCTL" upgrade cluster --name "$CLUSTER_NAME" --version "$EKS_TARGET_VERSION"
fi

REPLACE_NODE_GROUPS_STEPS="4"
echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Replacing node groups, step 1 of $REPLACE_NODE_GROUPS_STEPS: Create configuration for new node groups."
echo "------------------------------------------------------------------------------------------------------------------------"

cat <<EOF >$NODEGROUP_FILE
addons:
- attachPolicyARNs:
  - arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy
  name: vpc-cni
  permissionsBoundary: arn:aws:iam::$AWS_ACCOUNT:policy/oslokommune/oslokommune-boundary
cloudWatch:
  clusterLogging:
    enableTypes:
    - api
    - audit
    - authenticator
    - controllerManager
    - scheduler
fargateProfiles:
- name: fp-default
  selectors:
  - namespace: default
  - namespace: kube-system
  - namespace: argocd
iam:
  fargatePodExecutionRolePermissionsBoundary: arn:aws:iam::$AWS_ACCOUNT:policy/oslokommune/oslokommune-boundary
  serviceRolePermissionsBoundary: arn:aws:iam::$AWS_ACCOUNT:policy/oslokommune/oslokommune-boundary
  withOIDC: true
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: $CLUSTER_NAME
  region: $AWS_REGION
nodeGroups:
EOF

for AZ_ID in a b c
do
  AZ="${AWS_REGION}${AZ_ID}"

  cat <<EOF >>$NODEGROUP_FILE
  - name: "ng-generic-${EKS_VERSION_WITH_DASH}-1${AZ_ID}"
    availabilityZones: ["$AZ"]
    instanceType: "m5.large"
    desiredCapacity: 0
    minSize: 0
    maxSize: 10
    labels:
      pool: ng-generic-$AZ
    tags:
      k8s.io/cluster-autoscaler/enabled: "true"
      k8s.io/cluster-autoscaler/$CLUSTER_NAME: owned
    privateNetworking: true
EOF
done

echo "Written to: $NODEGROUP_FILE"

echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Replacing node groups, step 2 of $REPLACE_NODE_GROUPS_STEPS: Create new node group"
echo "------------------------------------------------------------------------------------------------------------------------"

if [[ $DRY_RUN == "false" ]]; then
  run_with_output "$EKSCTL" create nodegroup --config-file=$NODEGROUP_FILE
  run_with_output "$KUBECTL" set env daemonset aws-node -n kube-system ENABLE_POD_ENI=true
  run_with_output \
    "$KUBECTL" patch daemonset aws-node \
      -n kube-system \
      -p '{"spec": {"template": {"spec": {"initContainers": [{"env":[{"name":"DISABLE_TCP_EARLY_DEMUX","value":"true"}],"name":"aws-vpc-cni-init"}]}}}}'
else
  run_with_output "$EKSCTL" create nodegroup --config-file=$NODEGROUP_FILE --dry-run
  echo "Would run: $KUBECTL set env daemonset aws-node -n kube-system ENABLE_POD_ENI=true"

  PATCH='{"spec": {"template": {"spec": {"initContainers": [{"env":[{"name":"DISABLE_TCP_EARLY_DEMUX","value":"true"}],"name":"aws-vpc-cni-init"}]}}}}'
  echo "Would run:"
  echo "  $KUBECTL patch daemonset aws-node \\"
  echo "   -n kube-system \\"
  echo "   -p $PATCH"
fi

echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Replacing node groups, step 3 of $REPLACE_NODE_GROUPS_STEPS: Drain old nodes"
echo "------------------------------------------------------------------------------------------------------------------------"

# Nodesgroups we want to keep
NEW_NODE_1A=ng-generic-$EKS_VERSION_WITH_DASH-1a
NEW_NODE_1B=ng-generic-$EKS_VERSION_WITH_DASH-1b
NEW_NODE_1C=ng-generic-$EKS_VERSION_WITH_DASH-1c

EXISTING_NODEGROUPS=$($EKSCTL get nodegroup --cluster "$CLUSTER_NAME" -o yaml | yq eval '.[].Name')

echo "Existing nodegroups:"
echo "$EXISTING_NODEGROUPS"
echo
echo "We want to drain all node groups except these:"
echo "$NEW_NODE_1A"
echo "$NEW_NODE_1B"
echo "$NEW_NODE_1C"
echo
echo "Draining all other node groups now:"
echo

# Drain all nodegroups that are not one of the expected above.
for node_group in $EXISTING_NODEGROUPS
do
  # If node_group is not one of the ones we want to keep, we can drain it
  if [[ \
        "$node_group" != "$NEW_NODE_1A" && \
        "$node_group" != "$NEW_NODE_1B" && \
        "$node_group" != "$NEW_NODE_1C" \
    ]]; then

      echo "Draining node group: $node_group"

      if [[ $DRY_RUN == "false" ]]; then
        run_with_output "$EKSCTL" drain nodegroup --cluster "$CLUSTER_NAME" --name "$node_group"
      else
        echo "Would run: $EKSCTL drain nodegroup --cluster $CLUSTER_NAME --name $node_group"
      fi
  else
    echo "Not draining node group, we want to keep it: $node_group"
  fi
done

echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Replacing node groups, step 4 of $REPLACE_NODE_GROUPS_STEPS: Delete old nodegroups"
echo "------------------------------------------------------------------------------------------------------------------------"

echo "Existing nodegroups:"
echo "$EXISTING_NODEGROUPS"
echo
echo "We want to delete all node groups except these:"
echo "$NEW_NODE_1A"
echo "$NEW_NODE_1B"
echo "$NEW_NODE_1C"
echo
echo "Deleting all other node groups now:"

for node_group in $EXISTING_NODEGROUPS
do
  # If node_group is not one of the ones we want to keep, we can delete it
  if [[ \
        "$node_group" != "$NEW_NODE_1A" && \
        "$node_group" != "$NEW_NODE_1B" && \
        "$node_group" != "$NEW_NODE_1C" \
    ]]; then

      echo "Deleting node group: $node_group"

      if [[ $DRY_RUN == "false" ]]; then
        run_with_output "$EKSCTL" delete nodegroup --cluster "$CLUSTER_NAME" --name "$node_group"
      else
        echo "Would run: $EKSCTL delete nodegroup --cluster $CLUSTER_NAME --name $node_group"
      fi
  else
    echo "Not deleting node group, we want to keep it: $node_group"
  fi
done

UPDATE_ADDONS_STEPS=3
echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Upgrading addon 1 of $UPDATE_ADDONS_STEPS: kube-proxy"
echo "------------------------------------------------------------------------------------------------------------------------"

if [[ $DRY_RUN == "false" ]]; then
  run_with_output "$EKSCTL" utils update-kube-proxy --cluster="$CLUSTER_NAME" --approve
else
  run_with_output "$EKSCTL" utils update-kube-proxy --cluster="$CLUSTER_NAME"
fi


echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Upgrading addon 2 of $UPDATE_ADDONS_STEPS: aws-node"
echo "------------------------------------------------------------------------------------------------------------------------"

if [[ $DRY_RUN == "false" ]]; then
  run_with_output "$EKSCTL" utils update-aws-node --cluster="$CLUSTER_NAME" --approve
else
  run_with_output "$EKSCTL" utils update-aws-node --cluster="$CLUSTER_NAME"
fi


echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Upgrading addon 3 of $UPDATE_ADDONS_STEPS: coredns"
echo "------------------------------------------------------------------------------------------------------------------------"

if [[ $DRY_RUN == "false" ]]; then
  run_with_output "$EKSCTL" utils update-coredns --cluster="$CLUSTER_NAME" --approve
else
  run_with_output "$EKSCTL" utils update-coredns --cluster="$CLUSTER_NAME"
fi

echo
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Done"
echo "------------------------------------------------------------------------------------------------------------------------"
echo "Upgrading to EKS $EKS_TARGET_VERSION complete."
