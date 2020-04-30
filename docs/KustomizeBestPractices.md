<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Kustomize Best Practices](#kustomize-best-practices)
  - [Package layout.](#package-layout)
  - [Command Line substitution](#command-line-substitution)
  - [Eschew vars](#eschew-vars)
    - [Internal subsitution of fields Kustomize isn't aware of.](#internal-subsitution-of-fields-kustomize-isnt-aware-of)
    - [Global substitution](#global-substitution)
  - [Have separate packages for CR's and instances of the custom resource](#have-separate-packages-for-crs-and-instances-of-the-custom-resource)
  - [CommonLabels should be immutable](#commonlabels-should-be-immutable)
    - [1b. Removing common attributes across resources](#1b-removing-common-attributes-across-resources)
    - [1c. Parameters](#1c-parameters)
    - [1d. Overlays](#1d-overlays)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Kustomize Best Practices

  This doc provides best practices for writing Kubeflow kustomize packages.

## Package layout.

If your application consists of loosely coupled component e.g. backend, front-end, database consider defining these as separate kustomize packages
and then using kustomize to compose these applications into different installs e.g 

```
components/
          /app-front
          /app-backend
          /app-db
installs
          /app-standalone
          /app-onprem
```

Defining separate packages for each component makes it easier to use composition to define new configurations; e.g. using an external database as opposed
to a database running in cluster.

## Command Line substitution

To make it easy for users to override command line arguments use the following pattern.

1. Use a config map generator to store the parameters
1. On Deployments/StatefulSets/etc... set environment variables based on the config map
1. Rely on Kubernetes to substitute environment variables into container arguments ([ref](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/#using-environment-variables-inside-of-your-config))

Users can then override the parameters by defining [config map overlays](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/configGeneration.md).

Using a [ConfigMapGenerator](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/configGeneration.md#configmap-generation-and-rolling-updates) and including a content hash is highly prefered as this ensures
rolling updates are triggered if the config map is changed.

**Deprecated patterns**

* vars should no longer be used to do command line substitution see [bit.ly/kf_kustomize_v3](https://docs.google.com/document/d/1jBayuR5YvhuGcIVAgB1F_q4NrlzUryZPyI9lCdkFTcw/edit?pli=1#heading=h.ychbuvw81fj7)

## Eschew vars

As noted in [kubernetes-sigs/kustomize#2052](https://github.com/kubernetes-sigs/kustomize/issues/2052) vars have a lot of downsides.
For Kubeflow in particular vars have made it difficult to compose kustomize packages because they need to be unique globally ([kubeflow/manifests#1007](https://github.com/kubeflow/manifests/issues/1007)).

Vars should be used sparingly. Below are some guidance on acceptable use cases.


### Internal subsitution of fields Kustomize isn't aware of.

One ok use case for vars is getting kustomize to subsitute a value into a field kustomize wouldn't normally do substitution into. 
This often happens with CRDs. For example, consider the virtual service below from [jupyter-web-app](https://github.com/kubeflow/manifests/blob/master/jupyter/jupyter-web-app/overlays/istio/virtual-service.yaml).

```
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: jupyter-web-app
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - ...
    route:
    - destination:
        host: jupyter-web-app-service.$(jupyter-web-app-namespace).svc.$(clusterDomain)
        port:
          number: 80
```

We would like kustomize to substitute namespace into the destination host. We do this by

1. Defining a [vars](https://github.com/kubeflow/manifests/blob/393ec700e7834ca69a0832ec01ea2ecd90fb5bc4/jupyter/jupyter-web-app/base/kustomization.yaml#L63) to get the value for namespace
1. Defining a [custom configuration](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/transformerconfigs/README.md#customizing-transformer-configurations) so that the vars will be substituted into the virtual service host.

This use of vars is acceptable because the var is internal to the kustomize package and can be given a unique enough name to prevent
conflicts when the package is composed with other applications.

### Global substitution

One of the most problematic use cases for vars in Kubeflow today is substituting a user supplied value into multiple applications.

Currently we only have one primary use case for this today which is substituting in cluster domain into virtual services ([ref](https://docs.google.com/document/d/1jBayuR5YvhuGcIVAgB1F_q4NrlzUryZPyI9lCdkFTcw/edit#heading=h.vyq4iltpirga)).

We would ultimately like to get rid of the use of vars in these cases but have not settled on precise solutions. Some possible options are

1. Using [kpt setters](https://googlecontainertools.github.io/kpt/reference/cfg/create-subst/)

   * kpt is still relatively new and we don't want to mandate/require using it
   * consider adding kpt setters as appropriate so users who are willing to use kpt can avoid dealing with vars

1. Defining custom transformers 

   * e.g. we could define a new transformer for virtual services as discussed in [kubeflow/manifests#1007](https://github.com/kubeflow/manifests/issues/1007#issuecomment-599257347)


## Have separate packages for CR's and instances of the custom resource

If you are adding a custom resource (e.g. CertManager) and also defining instances of those resources (e.g. ClusterIssuer) (see [kubeflow/manifests#1121](https://github.com/kubeflow/manifests/issues/1121)).

Having separate packages makes it easier during deployment to ensure the custom resource is deployed and ready before trying to create instances
of the CR.

## CommonLabels should be immutable

As noted [here](https://kubectl.docs.kubernetes.io/pages/reference/kustomize.html#commonlabels) commonLabels get applied to
selectors which are immutable. Therefore, commonLabels should be immutable across versions of a package to avoid causing
problems during upgrades.

For more info see [kubeflow/manifests#1131](https://github.com/kubeflow/manifests/issues/1131)

###Resource file naming

  Resources should be organized by kind, where the resource is in a file that is the lower-case hyphenized form of the Resource kind. For example: a Deployment would go in a file named deployment.yaml. A ClusterRoleBinding would go in a file called cluster-role-binding.yaml. If there are multiple resources within a kustomize target (eg more than one deployment), you may want to maintain a single resource per file and add a prefix|suffix of the resource name to the filename. For example the file name would be `<kind>-<name>.yaml`. See below for an example.

> example: /manifests/profiles

```
profiles
└── base
    ├── README.md
    ├── cluster-role-binding.yaml
    ├── crd.yaml
    ├── deployment.yaml
    ├── kustomization.yaml
    ├── role-binding.yaml
    ├── role.yaml
    ├── service-account.yaml
    └── service.yaml
```

### 1b. Removing common attributes across resources

  There are often repeated attributes across resources: labels, namespace, or perhaps a common prefix used for each resource. You can move name prefixes into the kustomization.yaml file and then make adjustments within each resource; removing the prefix from its name. Additionaly you can move labels and their selectors into the kustomization.yaml. Yo can move the namespace into the kustomization.yaml. All of these will be added back into the resource by running `kustomize build`.

> example: /manifests/profiles/base/kustomization.yaml. Contains namespace, nameprefix, commonLabels.

```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- crd.yaml
- service-account.yaml
- cluster-role-binding.yaml
- role.yaml
- role-binding.yaml
- service.yaml
- deployment.yaml
namespace: kubeflow
namePrefix: profiles-
commonLabels:
  kustomize.component: profiles
images:
  - name: gcr.io/kubeflow-images-public/profile-controller
    newName: gcr.io/kubeflow-images-public/profile-controller
    newTag: v20190228-v0.4.0-rc.1-192-g1a802656-dirty-f95773
```


  The original deployment in profiles looked like:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    kustomize.component: profiles
  name: profiles-deployment
  namespace: kubeflow
spec:
  selector:
    matchLabels:
      kustomize.component: profiles
  template:
    metadata:
      labels:
        kustomize.component: profiles
    spec:
      containers:
      - command:
        - /manager
        image: gcr.io/kubeflow-images-public/profile-controller:v20190228-v0.4.0-rc.1-192-g1a802656-dirty-f95773
        imagePullPolicy: Always
        name: manager
      serviceAccountName: profiles-controller-service-account
```

  Moving labels, namespace and the nameprefix 'profiles-' to kustomization.yaml reduces deployment.yaml to

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  template:
    spec:
      containers:
      - name: manager
        command:
        - /manager
        image: gcr.io/kubeflow-images-public/profile-controller:v20190228-v0.4.0-rc.1-192-g1a802656-dirty-f95773
        imagePullPolicy: Always
      serviceAccountName: controller-service-account
```

  Note: A kustomize target should always 'build', so you should add what's needed to allow a `kustomize build` to succeed (and for unittests to work). Defining a namespace in kustomization.yaml is required to run `kustomize build`, even though there is a namespace override in the parent kustomization.yaml generated by kfctl under /manifests/profiles. This generated kustomization.yaml provides overrides using values from app.yaml and will appear within the manifest cache after running `kfctl generate...`. 

### 1c. Parameters

  There are 3 types of kustomize parameters that can be leveraged when building a kustomize target. The vars, patchesStrategicMerge and patchesJson6209 are described in detail in the [kustomization.yaml file](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/kustomization.yaml).

  1. [vars](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/kustomization.yaml#L226)<br/> 
     Kustomize vars should be used whenever a simple text replacement is needed. The text must be a string, for example you cannot use kustomize vars for deployment.spec.replicas which takes a int. You will need to use a Json Patch for this.  Kustomize vars require a section within the kustomization.yaml under `vars:` that identifies the variable name, its object reference and optionally its field reference in the object. This uniquely identifies the source of the variable. A configuration yaml may also be needed to identify those resources referencing the variable. In many cases the variable will reference a ConfigMap that is generated in the kustomization.yaml file using values contained in a params.env file. The configuration yaml will then list resources and their json paths where the variable is used eg $(varname). See [ambassador](https://github.com/kubeflow/manifests/tree/master/common/ambassador/base) for an example of using a var named [ambassadorServiceType](https://github.com/kubeflow/manifests/blob/master/common/ambassador/base/kustomization.yaml#L28) that is generated from a [configMapGenerator](https://github.com/kubeflow/manifests/blob/master/common/ambassador/base/kustomization.yaml#L15) which reads this value from [params.env](https://github.com/kubeflow/manifests/blob/master/common/ambassador/base/params.env#L1). In order to use ambassadorServiceType, you need to list which [resource](https://github.com/kubeflow/manifests/blob/master/common/ambassador/base/params.yaml#L1) uses it in params.yaml and finally, actually use the [variable](https://github.com/kubeflow/manifests/blob/master/common/ambassador/base/service.yaml#L30).
  2. [patchesStrategicMerge](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/kustomization.yaml#L149)<br/>
     This contains a list of resources which have one or more changes from the base resource. Only the changes and enough information to uniquely describe the resource should be in this resource file. This should not be used if you're changing an array in the base resource. For this you should use a Json Patch.
  3. [patchesJson6902](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/kustomization.yaml#L167)<br/>
     Json Patches are the most versatile, having the ability to add, replace or remove array entries in resources (eg `deployment/spec/template/spec/containers/env/[-+]`) as well as other sections that may be parts of the schema or integers. [Ambassador]'s kustomization.yaml also contains an example of a [json patch](https://github.com/kubeflow/manifests/blob/master/common/ambassador/base/kustomization.yaml#L20) that replaces an integer using the patch found in deployment-amabassador-patch.yaml in the ambassador [base directory](https://github.com/kubeflow/manifests/tree/master/common/ambassador/base).

### 1d. Overlays

  Certain resources or resource modifications can be further grouped by a particular concept that cuts across kustomize base targets such as platform type, an Istio service, basic-auth, etc. There are a number of examples where base targets are changed by specific overlays and can be used as a reference when adding a new target or overlay. One example of an overlay is under [profiles debug overlay](https://github.com/kubeflow/manifests/tree/master/profiles/overlays/debug). In this case, there is a need to debug the profile-controller using dlv/GoLand. This requires replacing the deployment command with '/go/bin/dlv' and adding args required by dlv. It also requires replacing the image with one that has been modified and matches with the debugger source code.

