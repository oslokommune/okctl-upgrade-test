# EKS upgrade

This document describes how to use the [upgrade script](upgrade.sh) in this repository in order to upgrade your EKS cluster.

The upgrade script basically just does what is described in step 1-3 in the official guide, https://eksctl.io/usage/cluster-upgrade/, which is:

> 1. upgrade control plane version with `eksctl upgrade cluster`
> 2. replace each of the nodegroups by creating a new one and deleting the old one
> 3. update default add-ons:
>     - `kube-proxy`
>     - `aws-node`
>     - `coredns`

## Tips / Read this!

* All steps in the scripts are written to be idempotent. This means if the script breaks, or if you edit it and want to re-run it, you can, and it should still work.
  * It also means you can abort the script at any time with CTRL+C, and re-run it again. (But don't do it for no reason, we cannot ever be really sure.)
* If the upgrade script breaks, or you want to customize it in any way, it's possible to edit the script to suit your needs.

## Prerequisites

**Make sure** you have already followed this guide previsouly to get to EKS 1.20:
https://github.com/oslokommune/okctl-upgrade/blob/main/gists/bump-eks-to-1-20/README.md. The exception to this if you have used Okctl to create a 1.20 cluster, as then it does not need upgrading.

# Step 1: Download or update tools

The upgrade script expects the following tools to exist on your machine, so make sure to install these.

## aws CLI

Follow instructions in https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html

## jq

jq is a tool for parsing JSON.

### Linux / apt

```sh
sudo apt-get install jq
```

### macOS / Linux via Homebrew

```sh
brew install jq
```

### Other

See: https://stedolan.github.io/jq/download/.

## yq

yq is a tool for parsing YAML.

### Linux / snap

```sh
snap install yq
```

###  macOS / Linux via Homebrew

```sh
brew install yq
```

### Other

See https://github.com/mikefarah/yq.

## watch

### Linux

No need to install, this ususally comes preinstalled in most distributions.

### macOS

```
brew install watch
```

# Step 2: Log in to the environment

Set `AWS_PROFILE` to the correct AWS profile from `~/.aws/config`. If you have not set this up, have a look at [Authenticating to AWS](https://www.okctl.io/authenticating-to-aws/#aws-single-sign-on-sso).

```sh
export AWS_PROFILE=some-account
aws sso login
```

# Step 3: Configure applications for no downtime

To avoid downtime, **make sure** you have completed the steps described in this guide: https://github.com/oslokommune/okctl-upgrade/tree/main/gists/bump-eks-to-1-20#prepare-applications

# Step 4: Adapt to EKS version specific requirements

Some EKS versions deprecate old resources types (i.e. how your Kubernetes resources must look like). Or
perhaps an EKS plugin must be upgraded.

This section describes what you must do to be able to upgrade to specific EKS versions.

The source for these suggestions are:
* https://docs.aws.amazon.com/eks/latest/userguide/kubernetes-versions.html
* https://docs.aws.amazon.com/eks/latest/userguide/update-cluster.html
* https://docs.aws.amazon.com/eks/latest/userguide/platform-versions.html
* https://kubernetes.io/docs/reference/using-api/deprecation-guide/

## EKS 1.21

Okctl doesn't support EKS 1.21 yet. We will update this guide when it is. (Note to ourselves:
"Replace this text line with: Make sure you have upgraded to Okctl version x.x.x or later.")

You can still upgrade to EKS 1.21, but you cannot create a cluster from scratch with version 1.21. You
will have to create it with version 1.20 and upgade to 1.21.

## EKS 1.22

If you're upgrading to EKS 1.22, you must follow the below steps. If you do not, your application **probably will stop working**.

AWS describes the necessary changes we need to take into account in detail: https://docs.aws.amazon.com/eks/latest/userguide/update-cluster.html#update-1.22. However, we have attempted to extract everything that is necessary, so you shouldn't need to read that guide.

### Update Ingress manifests

First `cd` into the directory where you store your Kubernetes manifests/YAMLs. The default way in Okctl is to put these in your IAC repository.

```yaml
cd your-okctl-iac-repository
```

Find occurrences of old Ingress resources by running

```
grep -nRsH "apiVersion: networking.k8s.io/v1beta1"
grep -nRsH "apiVersion: extensions/v1beta1"
```

Example output:

```
$ grep -nRsH "apiVersion: networking.k8s.io/v1beta1"
--
infrastructure/applications/okctl-reference-app/base/ingress.yaml:1:apiVersion: networking.k8s.io/v1beta1
infrastructure/applications/okctl-reference-app/base/ingress.yaml-2-kind: Ingress
--
infrastructure/applications/hello/base/ingress.yaml:1:apiVersion: networking.k8s.io/v1beta1
infrastructure/applications/hello/base/ingress.yaml-2-kind: Ingress
```

#### Update the YAML

* For every file in the result, edit it and replace the `apiVersion` so it becomes like this:

```yaml
apiVersion: networking.k8s.io/v1
```

* Also, you need to change the Ingress YAML somewhat:

Old ingress:

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
# ... (some more YAML)
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /
        backend:
          serviceName: echo
          servicePort: 80
```

You must rewrite this to:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
# ... (some more YAML)
spec:
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: test
                port:
                  number: 80
```

Some useful links regarding this change:
* [Example](https://docs.konghq.com/kubernetes-ingress-controller/latest/concepts/ingress-versions/)
* [Kubernetes Ingress documentation](https://kubernetes.io/docs/concepts/services-networking/ingress/#the-ingress-resource)

### Make sure Okctl version is x.x.x or later (ToDo: update this when Okctl supports 1.22)

Okctl doesn't support EKS 1.22 yet. We will update this guide when it is. (Note to ourselves:
"Replace this text line with: Make sure you have upgraded to Okctl version x.x.x or later.")

You can still upgrade to EKS 1.22, but you cannot create a cluster from scratch with version 1.21. You
will have to create it with version 1.20 and upgade to 1.21.

We need to bump the AWS load balancer controller to 2.4.1 or later. `okctl upgrade` handles this for us.

From update-1.22 documentation:

> - If you currently have the AWS Load Balancer Controller deployed to your cluster, you must update it to version `2.4.1` before updating your cluster to Kubernetes version `1.22`.

```
okctl venv -a aws-profile -c my-cluster.yaml

# TODO
kubectl -n kube-system get pod -o=jsonpath='{$.spec.template.spec.containers[:0].image}'
```

# Step 5: Monitor everything while upgrading

It's nice to see that stuff changes while upgrading, so while we run the upgrade script in the next step, we want to monitor pods and nodes. We'll monitor these things in a separate terminal.

## Pods

Open a new terminal window. Log in to AWS and your kubernetes cluster. The default Okctl way is to run `okctl venv` with your usual arguments.

```sh
export AWS_PROFILE=some-account
# Change my-cluster-dev.yaml to the correct file for your environment
okctl venv -a aws-profile -c my-cluster-dev.yaml
```

Then start monitoring pods:

```sh
watch -n 1 kubectl get pod --all-namespaces -o wide
```

What to look for when running the upgrade
* Are at least on of your applications' pods in 1/1 Running state at all times? If not, you probably haven't configured the `PodDisruptionBudget` correctly (see the step to configure applications above).
* During the upgrade, pods will be removed from old nodes and started on new nodes. Which nodes are new nodes? See the `kubectl get node`node age in the next section below.

## Nodes

Open a new terminal tab. Log in to your kubernetes cluster like above.

Then start monitoring nodes:

```sh
watch -n 4 kubectl get node -o wide
```

What to look for when running the upgrade:
* See that new nodes is launched (see the age column), and that old nodes are removed.
* See that they get the correct Kubernetes version

## Node groups

Open a new terminal tab. Log in to your kubernetes cluster like above.

In the following command, replace `my-cluster-dev.yaml` with your Okctl cluster manifest. Then start monitoring node groups:

```sh
watch -n 15 eksctl get nodegroup --cluster $(yq e '.metadata.name' "my-cluster-dev.yaml")
```

What to look for when running the upgrade:
* See that new node groups are created, and the old ones are removed

# Step 6: Run the upgrade

## Download upgrade script

Download latest version upgrade script (it may be updated at any time):

```sh
curl --silent --location "https://raw.githubusercontent.com/oslokommune/okctl-upgrade/main/gists/bump-eks/upgrade.sh"
```

## Run the upgrade

### Usage

```sh
USAGE:
upgrade.sh <cluster-manifest file> <aws-region> <EKS target version> [dry-run={false|true}] | tee logfile.txt

cluster-manifest file      The Okctl cluster manifest
aws-region                 AWS region
EKS target version         Example: 1.21
dry-run                    Default true. Set to false to actually run upgrade.
```

* You can upgrade only one minor version at the time. So if you are on EKS 1.20 and want to upgrade to EKS 1.22, you must first upgrade to 1.21, then to 1.22.

* :information_source: The `tee` thing in the following commands is there to create a nice upgrade log. You do not have to, but we recommend storing this (in git or somewhere else), because
  * It gives a pretty nice and accurate way of telling what you have done with your cluster, which can be useful for future reference.
  * It helps immensely for debugging in case something wrong happens.

### Example, upgrading EKS 1.20 to 1.22

```sh
export AWS_PROFILE=some-account
mkdir -p logs

# Dry run the upgrade, hoping to catch any errors before actually upgrading
./upgrade.sh cluster-dev.yaml eu-west-1 1.21 | tee "logs/eks-upgrade-1-21-$(date +"%Y-%m-%dx%H-%M-%S").log"

# Actually run the upgrade
./upgrade.sh cluster-dev.yaml eu-west-1 1.21 dry-run=false | tee "logs/eks-upgrade-1-21-$(date +"%Y-%m-%dx%H-%M-%S").log"

# Store the logs
git add logs
git commit -m "Add log for upgrade to EKS 1.22"

# Dry run the upgrade, hoping to catch any errors before actually upgrading
./upgrade.sh cluster-dev.yaml eu-west-1 1.22 | tee "logs/eks-upgrade-1-22-$(date +"%Y-%m-%dx%H-%M-%S").log"

# Actually run the upgrade
./upgrade.sh cluster-dev.yaml eu-west-1 1.22 dry-run=false | tee "logs/eks-upgrade-1-22-$(date +"%Y-%m-%dx%H-%M-%S").log"

# Store the log
git add logs
git commit -m "Add log for upgrade to EKS 1.22"
```

# Something wrong happened

## eksctl delete nodegroup cannot evict pods

Abort/CTRL+C your execution of `eksctl delete nodegroup` if it's running, because we will be running the
command below, which we don't want to run at the same time.

In the following command, replace

* `/tmp/eks-upgrade/1-21` with `/tmp/eks-upgrade/1-22` or whatever version you're running on. Run:
* `ng-generic-1-10-1c` with whichever node is failing to drain. You can see which in the output of the upgrade script.

```shell
/tmp/eks-upgrade/1-21/kubectl drain -l 'alpha.eksctl.io/nodegroup-name=ng-generic-1-20-1c' --ignore-daemonsets --delete-emptydir-data
```

This should output exactly which pods that cannot be evicted due to its `PodDisruptionBudget` (or for other reasons?).

## My applications have downtime when draining nodes

* Your application's Deployment must have `replicas: 2`.
* You need a working `PodDisruptionBudget`.

How to setup these correctly is described in https://github.com/oslokommune/okctl-upgrade/blob/main/gists/bump-eks-to-1-20/README.md.

# Resources

- https://eksctl.io/usage/cluster-upgrade/

## Commands

## Set desiredCapacity

If you really need to, you can in nodegroup_config.yaml set desiredCapacity to `1`. Or run:

```
aws autoscaling set-desired-capacity --desired-capacity 1 --auto-scaling-group-name eksctl-my-cluster-nodegroup-ng-generic-1-20-1c-NodeGroup-DFG36JFJY345
```

to have less down time. This is at the cost of having more nodes than needed.
