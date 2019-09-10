package tests_test

import (
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/target"
	"testing"
)

func writeDexIngressBase(th *KustTestHarness) {
	th.writeF("/manifests/dex-auth/dex-ingress/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: dex-ingress-admin
  namespace: istio-system
`)

	th.writeF("/manifests/dex-auth/dex-ingress/base/cluster-role.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: dex-ingress-admin
rules:
- apiGroups:
  - '*'
  resources:
  - secrets
  verbs:
  - create
  - delete
  - deletecollection
  - get
  - list
  - patch
  - update
  - watch
`)
	th.writeF("/manifests/dex-auth/dex-ingress/base/cluster-role-binding.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: dex-ingress-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: dex-ingress-admin
subjects:
- kind: ServiceAccount
  name: dex-ingress-admin
  namespace: istio-system
`)
	th.writeF("/manifests/dex-auth/dex-ingress/base/config-map.yaml", `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ingress-config
  namespace: istio-system
data:
  ingress_gen_self_cert.sh: |
    #!/bin/bash

    SSL_DIR=/var/ingress-certs
    mkdir -p ${SSL_DIR}

    IFS=', ' read -r -a array <<< "$DOMAINS"
    if [ ${#array[@]} -eq 0 ];
    then
      echo "Enter domain names for gateway server as arguments"
      exit
    fi

    for ((i = 0; i < ${#array[@]}; ++i)); do
        position=$(( $i + 1 ))
        ALT_NAMES="${ALT_NAMES}DNS.${position}=${array[$i]}\n"
    done

    cat << EOF > ${SSL_DIR}/req.cnf
    [req]
    req_extensions = v3_req
    distinguished_name = req_distinguished_name

    [req_distinguished_name]

    [ v3_req ]
    basicConstraints = CA:FALSE
    keyUsage = nonRepudiation, digitalSignature, keyEncipherment
    subjectAltName = @alt_names

    [alt_names]
    $(printf ${ALT_NAMES})
    EOF

    openssl genrsa -out ${SSL_DIR}/ca-key.pem 2048
    openssl req -x509 -new -nodes -key ${SSL_DIR}/ca-key.pem -days 1000 -out ${SSL_DIR}/ca.pem -subj "/CN=istio-ingressgateway-certs-ca"

    openssl genrsa -out ${SSL_DIR}/key.pem 2048
    openssl req -new -key ${SSL_DIR}/key.pem -out ${SSL_DIR}/csr.pem -subj "/CN=istio-ingressgateway-certs-ca" -config ${SSL_DIR}/req.cnf
    openssl x509 -req -in ${SSL_DIR}/csr.pem -CA ${SSL_DIR}/ca.pem -CAkey ${SSL_DIR}/ca-key.pem -CAcreateserial -out ${SSL_DIR}/cert.pem -days 1000 -extensions v3_req -extfile ${SSL_DIR}/req.cnf

    kubectl create secret tls istio-ingressgateway-certs  --cert=${SSL_DIR}/cert.pem --key=${SSL_DIR}/key.pem -n $NAMESPACE
    kubectl create secret tls istio-ingressgateway-certs-ca  --cert=${SSL_DIR}/ca.pem --key=${SSL_DIR}/ca-key.pem -n $NAMESPACE
`)
	th.writeF("/manifests/dex-auth/dex-ingress/base/job.yaml", `
---
apiVersion: batch/v1
kind: Job
metadata:
  name: ingress-gen-self-cert
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - command:
        - /var/ingress-configs/ingress_gen_self_cert.sh
        env:
        - name: NAMESPACE
          valueFrom:
            configMapKeyRef:
              key: istioNamespace
              name: dex-ingress-parameters
        - name: DOMAINS
          valueFrom:
            configMapKeyRef:
              key: domains
              name: dex-ingress-parameters
        image: krishnadurai/ingress-gen-self-cert:latest
        name: gen-self-cert
        volumeMounts:
        - mountPath: /var/ingress-configs/
          name: ingress-config
      restartPolicy: OnFailure
      serviceAccountName: dex-ingress-admin
      volumes:
      - configMap:
          defaultMode: 493
          name: ingress-config
        name: ingress-config
`)
	th.writeF("/manifests/dex-auth/dex-ingress/base/params.env", `
istioNamespace=istio-system
domains=example.org
`)
	th.writeK("/manifests/dex-auth/dex-ingress/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: istio-system

resources:
- service-account.yaml
- cluster-role.yaml
- cluster-role-binding.yaml
- config-map.yaml
- job.yaml

configMapGenerator:
- name: dex-ingress-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true

vars:
- name: istioNamespace
  objref:
    kind: ConfigMap
    name: dex-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.istioNamespace
- name: domains
  objref:
    kind: ConfigMap
    name: dex-ingress-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.domains
`)
}

func TestDexIngressBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/dex-auth/dex-ingress/base")
	writeDexIngressBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.EncodeAsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../dex-auth/dex-ingress/base"
	fsys := fs.MakeRealFS()
	_loader, loaderErr := loader.NewLoader(targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl())
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
