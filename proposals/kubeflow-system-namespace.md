# Proposal: kubeflow-system Namespace Migration

**Author:** Aniket Patil (@aniketpatil1121)  
**Status:** Draft – Seeking Maintainer Feedback

---

## Overview

This document proposes a gradual and safe migration of Kubeflow control-plane
components from the hard-coded `kubeflow` namespace to `kubeflow-system`,
or alternatively making the namespace configurable where required.

The goal is to improve security, isolation, and long-term maintainability
while preserving backward compatibility.

---

## Problem Statement

Kubeflow manifests currently assume a fixed control-plane namespace (`kubeflow`)
across multiple components and deployment methods.

This tight coupling:

- Limits namespace isolation and multi-tenant use cases
- Conflicts with Kubernetes and CNCF best practices
- Increases security and compliance risks
- Makes gradual refactoring and component isolation harder

---

## Motivation

Migrating control-plane components to `kubeflow-system`, or removing hard-coded
namespace assumptions, enables:

- Secure and rootless deployments
- Better separation between system and user workloads
- Easier enterprise and regulated-environment adoption
- Cleaner long-term architecture

---

## Why Now?

Kubeflow has already started partial adoption of `kubeflow-system`
(e.g. Training Operator v2 and related components).

Addressing namespace assumptions early helps:

- Avoid further hard-coded dependencies
- Reduce future migration complexity
- Align new work with an improved namespace model

---

## Initial Repository Scan Results

A repository-wide scan shows extensive hard-coded usage of the `kubeflow`
namespace across multiple layers.

### Affected Areas

#### Helm Charts (`experimental/`)
- Multiple charts define `namespace: kubeflow` in `values.yaml`
- Example components:
  - Katib
  - Notebook Controller
  - Model Registry
  - KServe Models Web App

This suggests Helm-based deployments need explicit namespace parameterization.

---

#### Kustomize Overlays (`applications/`, `common/`)
- Widespread use of `namespace: kubeflow` in `kustomization.yaml`
- Found across:
  - Pipelines
  - Katib
  - KServe
  - Central Dashboard
  - Jupyter Web App
  - Spark Operator

These references are deeply embedded and represent medium-to-high migration risk.

---

#### Istio and Network Policies (`common/`)
- Istio `AuthorizationPolicy` and `VirtualService` objects reference
  `kubeflow` explicitly
- NetworkPolicies are tightly scoped to the `kubeflow` namespace

This area is considered **high risk** due to cross-namespace
service communication assumptions.

---

#### Partial Adoption of `kubeflow-system`

Some components already deploy into `kubeflow-system`, including:

- Training Operator v2
- Trainer Controller
- Select NetworkPolicies

Other components (e.g. Katib) already support configurable namespaces,
which further indicates that a gradual and backward-compatible migration
strategy is technically feasible.


---

## Proposed Approach

### 1. Documentation First

- Identify and document all `kubeflow` namespace assumptions
- Propose a clear and staged migration path
- Collect maintainer feedback before code changes

---

### 2. Component-by-Component Migration

- Migrate one component at a time
- Start with a pilot component (e.g. Kubeflow Pipelines)
- Maintain **one PR per component** to reduce review complexity
- Avoid cross-component changes in a single PR

---

### 3. Testing Strategy

- Validate deployments in `kubeflow-system`
- Verify cross-namespace communication
- Review NetworkPolicies and Istio rules
- Ensure CI pipelines pass without namespace-specific assumptions

---

### 4. Gradual Rollout

- No breaking changes in a single release
- Preserve existing `kubeflow` namespace behavior
- Introduce configurability where needed
- Merge changes only after successful validation

---

## Non-Goals

This proposal does **not** aim to:

- Migrate all components in a single release
- Remove the `kubeflow` namespace immediately
- Change user-facing defaults in this phase
- Introduce new installation requirements

---

## Risk Mitigation

- Perform migrations incrementally
- Keep backward compatibility throughout
- Isolate each change to a single component and PR
- Explicitly validate networking and security policies
- Pause or revert if regressions are detected

---

## Success Criteria

This effort will be considered successful if:

- Components deploy successfully in `kubeflow-system`
- Existing `kubeflow` namespace deployments continue to work
- CI pipelines pass without namespace-specific assumptions
- No regressions are introduced across components

---

## Documentation Impact

No user-facing documentation will be updated in this phase.

`README.md` files and installation guides will be updated only after:

- Maintainer approval
- A successful pilot migration
- Confirmed backward compatibility

---

## Next Steps

1. Submit this proposal for maintainer feedback and scope confirmation
2. Incorporate review comments and finalize the phased migration strategy
3. Identify and confirm the pilot component (Kubeflow Pipelines)
4. Start a backward-compatible pilot migration for the selected component
5. Document implementation findings, risks, and follow-up recommendations
   to inform future component migrations

