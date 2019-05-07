# Kustomize Best Practices

  How to utilize kustomize processing directives to minimize errors and produce simple targets

## 1. Identify the resources that encompass a kustomize target

  In most cases these resources are being move from ksonnet, and per ksonnet component, often include resources like:
  - CustomResourceDefinition
  - ClusterRole, ClusterRoleBinding or Role, RoleBinding
  - ConfigMap
  - Deployment
  - Service
  - ServiceAccount
  - VirtualService 

  In most cases the collection of resources will have a component name from ksonnet or a name identifying its purpose.
  This collection of resources will be moved under `manifests/<component>/base`. 
  

### 1a. Resource naming

  Resources should be organized by kind, where the file name the resource is in is the lower-case hyphenized form of the Resource kind. EG: A Deployment would go in a file named deployment.yaml. A ClusterRoleBinding would go in a file called cluster-role-binding.yaml. If there are multiple resources within a kustomize target eg more than one deployment, you may want to retain a single resource per file and add a prefix|suffix of the resource name to the filename. EG the file name would be `<kind>-<name>.yaml`. The naming should map capital letters to lower hyphenized.

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


### Shared attributes across resources

  There are often repeated attributes across resources. These are often labels, namespace, or perhaps a common prefix used for each resource. You can move these into the kustomization.yaml file and make adjustments within each resource.

> example: /manifests/profiles

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
**namePrefix: profiles-**
**commonLabels:**
**  kustomize.component: profiles**
images:
  - name: gcr.io/kubeflow-images-public/profile-controller
    newName: gcr.io/kubeflow-images-public/profile-controller
    newTag: v20190228-v0.4.0-rc.1-192-g1a802656-dirty-f95773
```


### Identify common overlays

  Certain resources or resource modifications can be further grouped by a particular concept that cuts across components such as a platform type, an Istio Service, etc. Often other components will be split by similar overlays. 


### Parameters

  Ksonnet components that have migrated typically included parameters that will need to be added to the kustomization file and related resource. 
