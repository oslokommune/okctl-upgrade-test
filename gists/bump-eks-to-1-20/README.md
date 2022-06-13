This guide describes how to upgrade EKS from 1.19 to 1.20 in an EKS cluster.

# Upgrade Okctl environments

Download okctl 0.0.95 and run `okctl upgrade`.

This is required in order to make sure Loki spawns in the correct AZ.

# Update tools

* Download the latest version of [eksctl](https://github.com/weaveworks/eksctl/releases). (This guide is tested with 0.98.0). (Important: You need to run okctl upgrade before running this, as this breaks the 0.0.95 Loki upgrade)

```shell
curl -LO "https://dl.k8s.io/release/v1.20.15/bin/linux/amd64/kubectl"
wget -qO- https://dl.k8s.io/release/v1.20.15/bin/linux/amd64/kubectl | tar xvz -C /usr/local/bin/kubectl2
curl -s some_url | tar xvz - -C /tmp
```

macOS:

```shell
curl -LO "https://dl.k8s.io/release/v1.20.15/bin/darwin/amd64/kubectl"
```


* Download [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) CLI version 1.20:

Linux:

```shell
curl -LO "https://dl.k8s.io/release/v1.20.15/bin/linux/amd64/kubectl"
```

macOS:

```shell
curl -LO "https://dl.k8s.io/release/v1.20.15/bin/darwin/amd64/kubectl"
```

# Prepare applications

Your applications need one of the following configureations.

## Alternative 1: No downtime

* Deployment: Use `RollingUpdate` strategy
* Deployment: Use `replicas: 2` or more
* Create a `PodDisruptionBudget`

Google how to apply using `maxUnavailable=0` and `type: RollingUpdate`.

Example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello
  namespace: hello
spec:
  replicas: 2
  strategy:
    rollingUpdate:
      maxUnavailable: 0
```

## Alternative 2: For stateful applications that must not co-exist 

* Deployment: Use `Recreate` strategy
* Deployment: Use `replicas: 1`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello
  namespace: hello
spec:
  replicas: 1
  strategy:
    type: Recreate
```

## Add node selectors to pods using PVCs

Before we can bump nodes, we need to make sure that pods that use volumes (via PVCs), spawn on a node in the same AZ as the
volumes. If not the pod will not start, as it cannot find the PV.

To do this, we need to specify which AZ pods in Kubernetes should spawn on. The AZ should be the same as the AZ of the PVC the
application is using.

### List PVCs

To get a list of PVCs, run:

```shell
kubectl get pvc -A

kubectl -n NAMESPACE describe pv PV_ID
# Replace NAMESPACE and PV_ID with values from above command. PV_ID = VOLUME.
```

Look for a label like this `failure-domain.beta.kubernetes.io/zone=eu-west-1c`, `eu-west-1c` is the AZ.

### Update deployments

Now update all your Deployments (or Pods or StatefulSets) that refers to the PVCs of these PVs to use a `nodeSelector` with the same AZ as the PVC.

So for instance, in `deployment.yaml`, you can change from

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello
spec:
  template:
    spec:
      containers:
        - name: hello
```

to

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello
spec:
  template:
    spec:
      nodeSelector:
        failure-domain.beta.kubernetes.io/zone: eu-west-1c
      containers:
        - name: hello
```

## Add a PodDisruptionBudget for every application

A `PodDisruptionBudget` can be used to make sure for instance 1 pod is always in Running state when draining nodes. For more details, see [documentation](https://kubernetes.io/docs/tasks/run-application/configure-pdb/).

For each application, create a `infrastructure/applications/hello/base/pod-disruption-budget.yaml` with contents:

```yaml
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: hello
  namespace: hello
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: hello
```

**Important!** Your app's deployment must have `replicas: 2` or more. If not, it will be impossible to drain a node, because the pod can never be moved to a new node. Follow steps above to set `replicas` on your deployment.

Update `infrastructure/applications/hello/base/kustomization.yaml` so it includes `pod-disruption-budget.yaml`. For isntance:

```yaml
resources:
- service.yaml
- ingress.yaml
- namespace.yaml
- deployment.yaml
- pod-disruption-budget.yaml
```

## Apply changes

Run

```shell
git add .
git commit -m "Add node selector to deployments"
git push
``` 

ArgoCD will then update your apps.

If you don't want to wait, you can run

```shell
CLUSTER_NAME="my-cluster" # See "eksctl get cluster"
kustomize build infrastructure/applications/hello/overlays/$CLUSTER_NAME | kubectl apply -f -
```

# Bump EKS control plane

Remember to upgrade your okctl cluster to latest version before continuing.

Bump your EKS control plane, by running.

```shell
okctl venv ...
```

```shell
eksctl get cluster
```

# Replace my-cluster with the name of cluster you want to upgrade from above command.

```shell
CLUSTER_NAME="my-cluster"
```

```shell
eksctl upgrade cluster --name $CLUSTER_NAME --version 1.20
```

```shell
eksctl upgrade cluster --name $CLUSTER_NAME --version 1.20 --approve 
```

Estimated time: 11 min

# Update EKS add-ons

Remember to set

```sh
CLUSTER_NAME="my-cluster" # See "eksctl get cluster"
```

## Default addons

```shell
eksctl utils update-kube-proxy --cluster=$CLUSTER_NAME --approve
```

```shell
eksctl utils update-coredns --cluster=$CLUSTER_NAME --approve
```

```shell
eksctl utils update-aws-node --cluster=$CLUSTER_NAME --approve
```

```shell
kubectl -n kube-system set env daemonset aws-node ENABLE_POD_ENI=true --v=9
```

```shell
kubectl patch daemonset aws-node \
  -n kube-system \
  -p '{"spec": {"template": {"spec": {"initContainers": [{"env":[{"name":"DISABLE_TCP_EARLY_DEMUX","value":"true"}],"name":"aws-vpc-cni-init"}]}}}}'
```

## VPC-CNI addon

The recommended vpc-cni addon version for all Kubernetes versions is `1.11.0-eksbuild.1`
([source](https://docs.aws.amazon.com/eks/latest/userguide/managing-vpc-cni.html)).

Get the IAM role the VPC-CNI addon uses:

```shell
eksctl get addon --cluster $CLUSTER_NAME --name vpc-cni -o json
```

See field "IAMRole", it should be something like

```
arn:aws:iam::123456789012:role/eksctl-mycluster-addon-vpc-cni-Role1-DMGPR03HYLWR
```

Put it into an environment variable:

```shell
ROLE_ARN="arn:aws:iam::123456789012:role/eksctl-mycluster-addon-vpc-cni-Role1-DMGPR03HYLWR"
```

Then upgrade addon with commands below.

To roll back, just run the `eksctl update addon` command that last worked.

```shell
# Update vpc-cni addon
eksctl update addon \
  --cluster $CLUSTER_NAME \
  --name vpc-cni \
  --version 1.7.10-eksbuild.1 \
  --service-account-role-arn $ROLE_ARN
```

Wait until

```shell
watch -n 5 eksctl get addon --cluster $CLUSTER_NAME --name vpc-cni -o json
```

says `"Status": "ACTIVE"`.

```shell
eksctl update addon \
  --cluster $CLUSTER_NAME \
  --name vpc-cni \
  --version 1.8.0-eksbuild.1 \
  --service-account-role-arn $ROLE_ARN
```

Wait like above

```shell
eksctl update addon \
  --cluster $CLUSTER_NAME \
  --name vpc-cni \
  --version 1.9.3-eksbuild.1 \
  --service-account-role-arn $ROLE_ARN
```

Wait like above.

For the next step, for some reason it results in a configuration conflict if we don't use the `--force` flag, so we have  to add `--force here`.

```shell
eksctl update addon \
  --cluster $CLUSTER_NAME \
  --name vpc-cni \
  --version 1.10.1-eksbuild.1 \
  --service-account-role-arn $ROLE_ARN \
  --force 
```

```shell
eksctl update addon \
  --cluster $CLUSTER_NAME \
  --name vpc-cni \
  --version 1.10.3-eksbuild.1 \
  --service-account-role-arn $ROLE_ARN
```

Wait like above

```shell
eksctl update addon \
  --cluster $CLUSTER_NAME \
  --name vpc-cni \
  --version 1.11.0-eksbuild.1 \
  --service-account-role-arn $ROLE_ARN
```

Wait like above

# Bump EC2 nodes in your cluster

## Spin up new nodes

We're using 3 nodes to ensure we have 1 node for every AZ. We need one in every AZ to ensure that any applications using PVCs can
be placed on a node in the same AZ as the PVC. For instance: If some-app use a PVC in AZ B, we need to have a node in AZ B as
well.

In the following code snippet, replace:
* `CLUSTER_NAME` with the name from `eksctl get cluster`
* `REGION` with your region.

Then run it.

```shell
CLUSTER_NAME="my-cluster" # See "eksctl get cluster"
REGION="eu-west-1"
ACCOUNT="123456789012"

cat <<EOF >nodegroup_config.yaml
addons:
- attachPolicyARNs:
  - arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy
  name: vpc-cni
  permissionsBoundary: arn:aws:iam::$ACCOUNT:policy/oslokommune/oslokommune-boundary
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
  fargatePodExecutionRolePermissionsBoundary: arn:aws:iam::$ACCOUNT:policy/oslokommune/oslokommune-boundary
  serviceRolePermissionsBoundary: arn:aws:iam::$ACCOUNT:policy/oslokommune/oslokommune-boundary
  withOIDC: true
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: $CLUSTER_NAME
  region: $REGION
nodeGroups:
EOF

for AZ_ID in a b c
do
  AZ="${REGION}${AZ_ID}"

  cat <<EOF >>nodegroup_config.yaml
  - name: "ng-generic-1-20-1${AZ_ID}"
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

```

Now, create a new nodegroup:

```shell
eksctl create nodegroup --config-file=nodegroup_config.yaml
```

This might take 5 minutes. You can run

```shell
eksctl get nodegroup --cluster $CLUSTER_NAME
```

to have a look at the new node groups.

Make sure these nodes have correct settings (not sure if this is really needed):

```shell
kubectl -n kube-system set env daemonset aws-node ENABLE_POD_ENI=true --v=9
```

```shell
kubectl patch daemonset aws-node \
  -n kube-system \
  -p '{"spec": {"template": {"spec": {"initContainers": [{"env":[{"name":"DISABLE_TCP_EARLY_DEMUX","value":"true"}],"name":"aws-vpc-cni-init"}]}}}}'
```

## Optional: Set desiredCapacity

If you really need to, you can in nodegroup_config.yaml set desiredCapacity to `1`. Or run:

```
aws autoscaling set-desired-capacity --desired-capacity 1 --auto-scaling-group-name eksctl-my-cluster-nodegroup-ng-generic-1-20-1c-NodeGroup-DFG36JFJY345
```

to have less down time. This is at the cost of having more nodes than needed.

# Delete old node(s)

While you do all this, you can run

```
watch -n 2 kubectl get node -o wide
``` 

## Verify node(s) to delete before deleting them

You can skip this if you are sure what nodes are being deleted in the next step. Use `eksctl get nodegroup` to find names of 
node groups.

(Draining also sets a taint on the nodes, i.e. prohibits new pods to be scheduled on them. So there is no need to taint nodes before draining them.)

To see which nodes and pods are going to be drained, run the command below. `kubectl2` with the `2` means
we're running Kubectl version 1.20.15, not the version shipped with Okctl.

```shell
kubectl2 drain -l 'alpha.eksctl.io/nodegroup-name=ng-generic' --ignore-daemonsets --delete-emptydir-data --dry-run=client
```

Verify that the list of nodes above are indeed the nodes you want to drain.

### Optional: Drain nodes

This isn't needed as the next delete command do this. But if you want, you can drain nodes before deleting the node group.

```shell
kubectl2 drain -l 'alpha.eksctl.io/nodegroup-name=ng-generic' --ignore-daemonsets --delete-emptydir-data
```

## Delete the old nodegroup

Use `eksctl get nodegroup --cluster $CLUSTER_NAME` to verify names of the old node group. It should be `ng-generic`.

Then delete the nodegroup:

```shell
eksctl delete nodegroup --cluster $CLUSTER_NAME --name ng-generic
```

Verify that no pods is running on the nodes in the old nodegroup:

```shell
kubectl get pod -o wide
```

# Something wrong happened

## Apps have downtime when draining nodes

* Your app's Deployment must have `replicas: 2`.
* You need a working `PodDisruptionBudget`.

## Upgrading VPC-CNI addon fails

When upgrading vpc-cni addon to 1.10.3 without --force, this error is returned from eksctl:

```

  Every 2,0s: eksctl get addon --cluster okctl-reference-dev --name vpc-cni -o json                                                                                                                                                                       yngvarxd: Mon May 23 15:47:55 2022

2022-05-23 15:47:56 [ℹ]  eksctl version 0.98.0
2022-05-23 15:47:56 [ℹ]  using region eu-west-1
2022-05-23 15:47:56 [ℹ]  Kubernetes version "1.20" in use by cluster "okctl-reference-dev"
2022-05-23 15:47:57 [ℹ]  to see issues for an addon run `eksctl get addon --name <addon-name> --cluster <cluster-name>`
[
    {
        "Name": "vpc-cni",
        "Version": "v1.9.3-eksbuild.1",
        "NewerVersion": "v1.11.0-eksbuild.1,v1.10.3-eksbuild.1,v1.10.2-eksbuild.1,v1.10.1-eksbuild.1",
        "IAMRole": "arn:aws:iam::123456789012:role/eksctl-okctl-reference-dev-addon-vpc-cni-Role1-131WLM79CLTQ4",
        "Status": "DEGRADED",
        "Issues": [
            {
                "Code": "ConfigurationConflict",
                "Message": "Apply failed with 3 conflicts: conflicts with \"eksctl\" using apps/v1:\n- .spec.template.spec.containers[name=\"aws-node\"].livenessProbe.timeoutSeconds\nconflicts with \"kubectl-client-side-apply\" using apps/v1:\n- .spec.template.spec.containers[name=
\"aws-node\"].resources.requests\n- .spec.template.spec.containers[name=\"aws-node\"].resources.requests.cpu",
                "ResourceIDs": null
            }
        ]
    }
]Issue: {Code:ConfigurationConflict Message:Apply failed with 3 conflicts: conflicts with "eksctl" using apps/v1:
- .spec.template.spec.containers[name="aws-node"].livenessProbe.timeoutSeconds
conflicts with "kubectl-client-side-apply" using apps/v1:
- .spec.template.spec.containers[name="aws-node"].resources.requests
- .spec.template.spec.containers[name="aws-node"].resources.requests.cpu ResourceIDs:[]}
```

# Resources

- https://eksctl.io/usage/cluster-upgrade/
