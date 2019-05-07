# Kustomize Best Practices

  How to utilize kustomize processing directives to minimize errors and produce simple targets

## Identify the resources that encompass a kustomize target

  In most cases these resources are being move from ksonnet, and per ksonnet component, often include resources like:
  - CustomResourceDefinition
  - ClusterRole, ClusterRoleBinding or Role, RoleBinding
  - ConfigMap
  - Deployment
  - Service
  - ServiceAccount
  - VirtualService 

  In most cases the collection of resources will have a component name from ksonnet or a name identifying its purpose.
  Move the collection of resources under manifests/<component>/base. 
  

### Resource naming

  Resources should be organized by kind, where the file name is the plural form of the Resource kind. A Deployment would go in a file named deployment.yaml. If there is a need to separate multiple deployments across 'deployment' files, you should add a prefix of the name to the filename. EG the file name would be `<kind plural>-<name>.yaml`. The naming should map capital letters to lower with dashes eg the file name for the ClusterRoleMapping resource would be cluster-rol-mapping.yaml.


### Shared attributes across resources

  Look for common, repeated attributes across resources. These are often labels, namespace, a common prefix used for each resource. Move these into the kustomization.yaml file as commonLabels, namespace and nameprefix respectively.


### Identify common overlays

  Certain resources or resource modifications can be further grouped by a particular concept that cuts across components such as a platform type, an Istio Service, etc. Often other components will be split by similar overlays. 


### Parameters

  Ksonnet components that have migrated typically included parameters that will need to be added to the kustomization file and related resource. 
