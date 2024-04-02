# kubeflow-kluctl

This directory contains a [Kluctl](https://kluctl.io) based deployment project for [Kubeflow](https://www.kubeflow.org/).


## Table of Content

<!-- TOC -->
* [kubeflow-kluctl](#kubeflow-kluctl)
  * [Table of Content](#table-of-content)
  * [Motivation](#motivation)
  * [Why Kluctl?](#why-kluctl)
  * [Prerequisites](#prerequisites)
  * [Create a Kind cluster](#create-a-kind-cluster)
  * [Configuration](#configuration)
  * [Deploying](#deploying)
  * [Future](#future)
<!-- TOC -->

## Motivation

This project has started as a PoC to demonstrate the capabilities of Kluctl as a deployment tool for complex Kubernetes
deployment projects. Kubeflow turns out to be a much more complex deployment than usual Kubernetes projects, leading to a lot
of complicated and partially manual steps required to install and maintain a Kubeflow instance.

I believe that Kluctl is able to simplify this process a lot, so I started building this PoC. My motivation is to
get attention on the Kluctl project and in best case find users that feel that their needs are fulfilled, ultimately
bringing adoption and maintainers to the project.

At the same time, I believe Kubeflow could potentially benefit from this effort. This project might even turn out to be
a viable distribution of Kubeflow as it makes installing and long-term maintaining it so much easier.

I have read into a lot of issues in the [manifests](https://github.com/kubeflow/manifests) and also looked into other
distributions (platform/cloud specific and more generic solutions like deployKF). What I found so far confirms my
assumptions.

## Why Kluctl?

To make it short: Because `kluctl deploy -a config=./my-config.yaml` is enough to install a fully functional Kubeflow
instance to an existing cluster. With the same command, you'll do upgrades and cleanups, re-deploy with new
configuration, and so on. There is no need for ArgoCD, FluxCD, or any other additional tooling, because the CLI is able
to perform all necessary configuration management and orchestration.

If you're already familiar with the [manifests](https://github.com/kubeflow/manifests) repo, you can think of Kluctl as a
replacement for the top-level `kustomization.yaml` and the top-level `kustomize build | kubectly apply -f-` invocation.

The difference is that it allows much better control over the deployment process. For example, the
[`deployment.yaml`](https://kluctl.io/docs/kluctl/deployments/deployment-yml/) files found on every folder level
(starting at the root), allow to control deployment order by introducing [`barriers`](https://kluctl.io/docs/kluctl/deployments/deployment-yml/#barriers)
between individual deployment items. The deployment items itself are either
[includes of sub-deployments](https://kluctl.io/docs/kluctl/deployments/deployment-yml/#includes), simple
[Kustomize](https://kluctl.io/docs/kluctl/deployments/deployment-yml/#kustomize-deployments) deployments or
[Helm Charts](https://kluctl.io/docs/kluctl/deployments/helm/).

Also, whole sub-deployments and individual items can be disabled conditionally, as seen for example in `common/cert-manager/deployment.yaml`.

Ultimately, these features allow to remove most (if not all?) manual interventions required by the user while installing
or upgrading Kubeflow. For example, there is no need to choose different sets of Kustomize overlays based on which type
of auth setup you want to install. Instead, this can be configured via configuration files and will automatically lead
to all required modifications to the deployment process.

This also allows to easily implement features like "bring your own cert-manager or istio", simply by changing
configuration and using appropriate conditionals and templating in the deployment project.

Another advantage that comes for free is the integration Helm Charts with proper support of templated values. This can
for example be seen in `common/dex/helm-values.yaml`. 

## Prerequisites

You'll need the following things before you can start with this deployment project.

1. Kluctl must be installed. Follow the [installation](https://kluctl.io/docs/kluctl/installation/) instructions.
   The minimum Kluctl version is 2.24.0, which might not be released at the time this README.md was written.
   Please download the [devel release](https://github.com/kluctl/kluctl/releases/tag/devel) in that case.
2. You need a Kubernetes cluster. You can use a naked cluster without anything pre-installed or a cluster with istio and cert-manager pre-installed. See the next chapter for instructions to setup a local cluster.
3. You need to clone this repo to a local directory.

## Create a Kind cluster

You can skip this if you bring your own cluster. Otherwise, [install Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
anc create a local cluster:

```sh
cat <<EOF | kind create cluster --name=kubeflow  --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: ClusterConfiguration
    apiServer:
      extraArgs:
        "service-account-issuer": "kubernetes.default.svc"
        "service-account-signing-key-file": "/etc/kubernetes/pki/sa.key"
- role: worker
- role: worker
- role: worker
EOF
```

Please note that the above Kind config creates 3 worker nodes. We need these as you'll otherwise run out of CPU resources.
A future version of this deployment project will support running with lower resource quotas to allow easier local testing.

You'll now have the Kubernetes context `kind-kubeflow` setup and configured as the current context. We'll later invoke
Kluctl with `--context kind-kubeflow`, which you can actually skip if your current context is setup properly. We'll still
do this to avoid that you end up messing another cluster up. Future versions of this deployment project will properly
support/describe using [targets](https://kluctl.io/docs/kluctl/kluctl-project/targets/) with [contexts](https://kluctl.io/docs/kluctl/kluctl-project/targets/#context)
being bound.

## Configuration

Now it's time to create your own configuration by copying `sample-config.yaml` to `my-config.yaml` and perform
the desired modifications. Check the contents of the `sample-config.yaml` and `config/*-defaults.yaml` for all available
configuration options.

If you look into [`deployment.yaml`](./deployment.yaml) you'll see multiple
[vars sources](https://kluctl.io/docs/kluctl/templating/variable-sources/#file) being loaded. These are merged together
and later used in the deployment sources itself to perform some [templating](https://kluctl.io/docs/kluctl/templating/).

There is also an optional (marked via `when: args.config`) that allows to load additional configuration from an externally
provided configuration file. This is done via the `-a config=my-config.yaml` later.

## Deploying

To actually deploy Kubeflow, run:

```sh
kluctl deploy --context=kind-kubeflow -a config=my-config.yaml
```

This will perform a dry-run first and show a diff before actually deploying. The diff must be approved by pressing `y`.
After that, it shows what actually happened.

This command is basically all you'll need to do to re-deploy with new configuration or later update to newer versions.
The dry-run based diff will give you some confidence in what you're doing as you'll always know what's going to happen
before it actually happens.

Try it out and modify the configuration. A good test is to change the authentication mode from auth-service to
oauth2-proxy by switching the `enabled` flags in your config appropriately. After this, re-deploy:

```sh
kluctl deploy --context=kind-kubeflow -a config=my-config.yaml --prune
```

Please note the `--prune` that we added. It will instruct Kluctl to not just detect orphan resources, but actually prune
them. If omitted, you can also use `kluctl prune --context=kind-kubeflow -a config=my-config.yaml` to prune/cleanup.

## Future

- More configuration
- Integration of manifests repo
- keeping components up-to-date
- SOPS
- GitOps
