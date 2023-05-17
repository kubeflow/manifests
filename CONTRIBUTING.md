# Kubeflow Manifests Contributor Guide

Welcome to the Kubeflow project! We'd love to accept your patches and
contributions to this project. Please read the
[contributor's guide in our docs](https://www.kubeflow.org/docs/about/contributing/).

The contributor's guide
* shows you where to find the Contributor License Agreement (CLA) that you need
  to sign,
* helps you get started with your first contribution to Kubeflow,
* and describes the pull request and review workflow in detail, including the
  OWNERS files and automated workflow tool.

## Leadership Roles

The [kubeflow/manifests](https://github.com/kubeflow/manifests) repo is owned by
the [Kubeflow Manifests Working Group](https://github.com/kubeflow/community/blob/master/wgs.yaml).

Kubeflow's [governance model](https://github.com/kubeflow/community/blob/master/wgs/wg-governance.md)
includes a plethora of different leadership roles.
This section aims to provide a clear description of what these roles mean for
this repo, as well as set expectations from people with these roles and requirements
for people to be promoted in a role.

A Working Group lead is considered someone that has either the role of
**Subproject Owner**, **Tech Lead** or **Chair**. These roles were defined by trying
to provide different responsibility levels for repo owners. For the Manifests WG
we'd like to start by treating *approvers* in the root [OWNERS](https://github.com/kubeflow/manifests/blob/master/OWNERS),
as Subproject Owners, Tech Leads and Chairs. This is done to ensure we have a
simple enough model to start that people can understand and get used to. So for
the Manifests WG we only have Manifests WG Leads, which are the root approvers.

The following sections will aim to define the requirements for someone to become
a reviewer and an approver in the root OWNERS file (Manifests WG Lead).

### Manifests WG Lead Requirements

The requirements for someone to be a Lead come from the processes and work required
to be done in this repo. The main goal with having multiple Leads is to ensure
that in case there's an absence of one of the Leads the rest will be able to ensure
the established processes and the health of the repo will be preserved.

With the above the main pillars of work and responsibilities that we've seen for
this repo throughout the years are the following:
1. Being involved with the release team, since the [release process](https://github.com/kubeflow/community/tree/master/releases) is tightly intertwined with the manifests repo
2. Testing methodologies (GitHub Actions, E2E testing with AWS resources etc)
3. Processes regarding the [contrib/addon](https://github.com/kubeflow/manifests/blob/master/contrib) components
4. [Common manifests](https://github.com/kubeflow/manifests/tree/master/common)  maintained by Manifests WG (Istio, Knative, Cert Manager etc)
5. Community and health of the project

Root approvers, or Manifests WG Leads, are expected to have expertise and be able
to drive all the above areas. Root reviewers on the other hand are expected to
have knowledge in all the above and have as a goal to grow into the approvers
role by helping with reviews throughout the project.

#### Root Reviewer requirements

* Primary reviewer for at least 2 PRs in `/common`
* Primary reviewer for at least 1 PR in `/tests`
* Primary reviewer for at least 1 PR in `/contrib`

#### Root Approver requirements

The goal of the requirements is to quantify the main pillars that we documented
above. The high level reasoning is that approvers should have lead efforts and
have expertise in the different processes and artefacts maintained in this repo
as well as be invested in the community of the WG.

* Need to be a root reviewer
* Have been a WG liaison for a KF release
* Has at least 3 substantial PRs merged for `/common`
    * I.e. updating the versions of Istio, Dex or Knative
    * I.e. update manifests to work with newer versions of kustomize
* Has at least 2 substantial PRs merged for `/testing`
* Has at least 1 substantial PR merged for `/proposals`
* Has been an active participant in issues and PRs for at least 3 months
* Has attended at least 75% of Manifests WG meetings
