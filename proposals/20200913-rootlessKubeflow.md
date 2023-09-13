# Rootless Kubeflow

### Goals

We want to run Kubeflow 99 % rootless accoring to CNCF/Kubernetes best practices.

The main steps are adding an additional profile for istio-cni or ambient mesh, updating the documentation and manifest generation process. 
Then adding the baseline and restricted PSS as kustomize component to /contrib and extending the profile controller to annotate user namespaces with configurable PSS labels.

We want to use a staged approach.
First Stage:

1. Implement Istio 1.17.5 and use it by default. This is important for the Kubeflow 1.8  feature freeze
2. Implement istio-cni (--set components.cni.enabled=true --set components.cni.namespace=kube-system) as second option.
3. Add simple tests similar to tests/gh-actions/install_istio.sh and tests/gh-actions/install_knative.sh for istio-cni and support both rootfull and rootless istio at the same time and give users one release to test

Second stage in a second PR:
4. Add pod security standards (https://kubernetes.io/docs/concepts/security/pod-security-standards/) base/restricted to manifests/contrib
5. Add istio-ambient as an option to Kubeflow 1.9
6. Enforce PSS baseline (here you can still build OCI containers via Podman and buildah) in Kubeflow 1.9. It works with any istio
7. Warning for violations of restricted PSS
8. Optionally Enforce PSS restricted (this is where minor corner cases are affected) in Kubeflow 1.10

Third stage:
10. Upgrade Istio to 1.19 and use ambient service mesh by default in Kubeflow 1.10

### Non-Goals
THis does not cover Application level CVEs, only cluster level security.

### Does this break any existing functionality?
So far not. Only PSS restricted may block Docker in Docker. In PSS baseline you can still build OCI images with Podman

### Does this fix/solve any outstanding issues?
We are not following best practices and this is forbidden in most enterprise environments.
