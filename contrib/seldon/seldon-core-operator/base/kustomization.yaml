# List of resource files that kustomize reads, modifies
# and emits as a YAML string
resources:
- resources.yaml

configurations:
  - kustomizeconfig.yaml

vars:
- name: CERTIFICATE_NAMESPACE # namespace of the certificate CR
  objref:
    kind: Certificate
    group: cert-manager.io
    version: v1
    name: seldon-serving-cert # this name should match the one in certificate.yaml
  fieldref:
    fieldpath: metadata.namespace
- name: CERTIFICATE_NAME
  objref:
    kind: Certificate
    group: cert-manager.io
    version: v1
    name: seldon-serving-cert # this name should match the one in certificate.yaml
- name: SERVICE_NAMESPACE # namespace of the service
  objref:
    kind: Deployment
    version: v1
    group: apps
    name: seldon-controller-manager
  fieldref:
    fieldpath: metadata.namespace
