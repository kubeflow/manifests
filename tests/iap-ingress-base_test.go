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

func writeIapIngressBase(th *KustTestHarness) {
	th.writeF("/manifests/gcp/iap-ingress/base/backend-config.yaml", `
apiVersion: cloud.google.com/v1beta1
kind: BackendConfig
metadata:
  name: iap-backendconfig
spec:
  iap:
    enabled: true
    oauthclientCredentials:
      secretName: $(oauthSecretName)
`)
	th.writeF("/manifests/gcp/iap-ingress/base/certificate.yaml", `
apiVersion: certmanager.k8s.io/v1alpha1
kind: Certificate
metadata:
  name: $(tlsSecretName)
spec:
  acme:
    config:
    - domains:
      - $(hostname)
      http01:
        ingress: $(ingressName)
  commonName: $(hostname)
  dnsNames:
  - $(hostname)
  issuerRef:
    kind: ClusterIssuer
    name: $(issuer)
  secretName: $(tlsSecretName)
`)
	th.writeF("/manifests/gcp/iap-ingress/base/cloud-endpoint.yaml", `
apiVersion: ctl.isla.solutions/v1
kind: CloudEndpoint
metadata:
  name: $(appName)
spec:
  project: $(project)
  targetIngress:
    name: $(ingressName)
    namespace: $(istioNamespace)
`)
	th.writeF("/manifests/gcp/iap-ingress/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: envoy
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: envoy
subjects:
- kind: ServiceAccount
  name: envoy
`)
	th.writeF("/manifests/gcp/iap-ingress/base/cluster-role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: envoy
rules:
- apiGroups:
  - ""
  resources:
  - services
  - configmaps
  - secrets
  verbs:
  - get
  - list
  - patch
  - update
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - get
  - list
  - update
  - patch
- apiGroups:
  - authentication.istio.io
  resources:
  - policies
  verbs:
  - '*'
- apiGroups:
  - networking.istio.io
  resources:
  - gateways
  - virtualservices
  verbs:
  - '*'
`)
	th.writeF("/manifests/gcp/iap-ingress/base/config-map.yaml", `
---
apiVersion: v1
data:
  healthcheck_route.yaml: |
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: default-routes
      namespace: $(namespace)
    spec:
      hosts:
      - "*"
      gateways:
      - kubeflow-gateway
      http:
      - match:
        - uri:
            exact: /healthz
        route:
        - destination:
            port:
              number: 80
            host: whoami-app.kubeflow.svc.cluster.local
      - match:
        - uri:
            exact: /whoami
        route:
        - destination:
            port:
              number: 80
            host: whoami-app.kubeflow.svc.cluster.local
    ---
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: kubeflow-gateway
      namespace: $(namespace)
    spec:
      selector:
        istio: ingressgateway
      servers:
      - port:
          number: 80
          name: http
          protocol: HTTP
        hosts:
        - "*"
  setup_backend.sh: |
    #!/usr/bin/env bash
    #
    # A simple shell script to configure the backend timeouts and health checks by using gcloud.
    [ -z ${NAMESPACE} ] && echo Error NAMESPACE must be set && exit 1
    [ -z ${SERVICE} ] && echo Error SERVICE must be set && exit 1
    [ -z ${INGRESS_NAME} ] && echo Error INGRESS_NAME must be set && exit 1

    PROJECT=$(curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/project/project-id)
    if [ -z ${PROJECT} ]; then
      echo Error unable to fetch PROJECT from compute metadata
      exit 1
    fi

    PROJECT_NUM=$(curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/project/numeric-project-id)
    if [ -z ${PROJECT_NUM} ]; then
      echo Error unable to fetch PROJECT_NUM from compute metadata
      exit 1
    fi

    # Activate the service account
    gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}
    # Print out the config for debugging
    gcloud config list

    NODE_PORT=$(kubectl --namespace=${NAMESPACE} get svc ${SERVICE} -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}')
    echo "node port is ${NODE_PORT}"

    while [[ -z ${BACKEND_NAME} ]]; do
      BACKENDS=$(kubectl --namespace=${NAMESPACE} get ingress ${INGRESS_NAME} -o jsonpath='{.metadata.annotations.ingress\.kubernetes\.io/backends}')
      echo "fetching backends info with ${INGRESS_NAME}: ${BACKENDS}"
      BACKEND_NAME=$(echo $BACKENDS | grep -o "k8s-be-${NODE_PORT}--[0-9a-z]\+")
      echo "backend name is ${BACKEND_NAME}"
      sleep 2
    done

    while [[ -z ${BACKEND_ID} ]]; do
      BACKEND_ID=$(gcloud compute --project=${PROJECT} backend-services list --filter=name~${BACKEND_NAME} --format='value(id)')
      echo "Waiting for backend id PROJECT=${PROJECT} NAMESPACE=${NAMESPACE} SERVICE=${SERVICE} filter=name~${BACKEND_NAME}"
      sleep 2
    done
    echo BACKEND_ID=${BACKEND_ID}

    JWT_AUDIENCE="/projects/${PROJECT_NUM}/global/backendServices/${BACKEND_ID}"

    # For healthcheck compare.
    mkdir -p /var/shared
    echo "JWT_AUDIENCE=${JWT_AUDIENCE}" > /var/shared/healthz.env
    echo "NODE_PORT=${NODE_PORT}" >> /var/shared/healthz.env
    echo "BACKEND_ID=${BACKEND_ID}" >> /var/shared/healthz.env

    if [[ -z ${USE_ISTIO} ]]; then
      # TODO(https://github.com/kubeflow/kubeflow/issues/942): We should publish the modified envoy
      # config as a config map and use that in the envoy sidecars.
      kubectl get configmap -n ${NAMESPACE} envoy-config -o jsonpath='{.data.envoy-config\.json}' |
        sed -e "s|{{JWT_AUDIENCE}}|${JWT_AUDIENCE}|g" > /var/shared/envoy-config.json
    else
      # Use kubectl patch.
       echo patch JWT audience: ${JWT_AUDIENCE}
       kubectl -n ${NAMESPACE} patch policy ingress-jwt --type json -p '[{"op": "replace", "path": "/spec/origins/0/jwt/audiences/0", "value": "'${JWT_AUDIENCE}'"}]'
    fi

    echo "Clearing lock on service annotation"
    kubectl patch svc "${SERVICE}" -p "{\"metadata\": { \"annotations\": {\"backendlock\": \"\" }}}"

    checkBackend() {
      # created by init container.
      . /var/shared/healthz.env

      # If node port or backend id change, so does the JWT audience.
      CURR_NODE_PORT=$(kubectl --namespace=${NAMESPACE} get svc ${SERVICE} -o jsonpath='{.spec.ports[0].nodePort}')
      read -ra toks <<<"$(gcloud compute --project=${PROJECT} backend-services list --filter=name~k8s-be-${CURR_NODE_PORT}- --format='value(id,timeoutSec)')"
      CURR_BACKEND_ID="${toks[0]}"
      CURR_BACKEND_TIMEOUT="${toks[1]}"
      [[ "$BACKEND_ID" == "$CURR_BACKEND_ID" && "${CURR_BACKEND_TIMEOUT}" -eq 3600 ]]
    }

    # Verify configuration every 10 seconds.
    while true; do
      if ! checkBackend; then
        echo "$(date) WARN: Backend check failed, restarting container."
        exit 1
      fi
      sleep 10
    done
  update_backend.sh: |
    #!/bin/bash
    #
    # A simple shell script to configure the backend timeouts and health checks by using gcloud.

    [ -z ${NAMESPACE} ] && echo Error NAMESPACE must be set && exit 1
    [ -z ${SERVICE} ] && echo Error SERVICE must be set && exit 1
    [ -z ${INGRESS_NAME} ] && echo Error INGRESS_NAME must be set && exit 1

    PROJECT=$(curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/project/project-id)
    if [ -z ${PROJECT} ]; then
      echo Error unable to fetch PROJECT from compute metadata
      exit 1
    fi

    # Activate the service account, allow 5 retries
    for i in {1..5}; do gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS} && break || sleep 10; done

    NODE_PORT=$(kubectl --namespace=${NAMESPACE} get svc ${SERVICE} -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}')
    echo node port is ${NODE_PORT}

    while [[ -z ${BACKEND_NAME} ]]; do
      BACKENDS=$(kubectl --namespace=${NAMESPACE} get ingress ${INGRESS_NAME} -o jsonpath='{.metadata.annotations.ingress\.kubernetes\.io/backends}')
      echo "fetching backends info with ${INGRESS_NAME}: ${BACKENDS}"
      BACKEND_NAME=$(echo $BACKENDS | grep -o "k8s-be-${NODE_PORT}--[0-9a-z]\+")
      echo "backend name is ${BACKEND_NAME}"
      sleep 2
    done

    while [[ -z ${BACKEND_SERVICE} ]];
    do BACKEND_SERVICE=$(gcloud --project=${PROJECT} compute backend-services list --filter=name~k8s-be-${NODE_PORT}- --uri);
    echo "Waiting for the backend-services resource PROJECT=${PROJECT} NODEPORT=${NODE_PORT} SERVICE=${SERVICE}...";
    sleep 2;
    done

    while [[ -z ${HEALTH_CHECK_URI} ]];
    do HEALTH_CHECK_URI=$(gcloud compute --project=${PROJECT} health-checks list --filter=name~${BACKEND_NAME} --uri);
    echo "Waiting for the healthcheck resource PROJECT=${PROJECT} NODEPORT=${NODE_PORT} SERVICE=${SERVICE}...";
    sleep 2;
    done

    echo health check URI is ${HEALTH_CHECK_URI}

    # Since we create the envoy-ingress ingress object before creating the envoy
    # deployment object, healthcheck will not be configured correctly in the GCP
    # load balancer. It will default the healthcheck request path to a value of
    # / instead of the intended /healthz.
    # Manually update the healthcheck request path to /healthz
    if [[ ${HEALTHCHECK_PATH} ]]; then
      # This is basic auth
      echo Running health checks update ${HEALTH_CHECK_URI} with ${HEALTHCHECK_PATH}
      gcloud --project=${PROJECT} compute health-checks update http ${HEALTH_CHECK_URI} --request-path=${HEALTHCHECK_PATH}
    else
      # /healthz/ready is the health check path for istio-ingressgateway
      echo Running health checks update ${HEALTH_CHECK_URI} with /healthz/ready
      gcloud --project=${PROJECT} compute health-checks update http ${HEALTH_CHECK_URI} --request-path=/healthz/ready
      # We need the nodeport for istio-ingressgateway status-port
      STATUS_NODE_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="status-port")].nodePort}')
      gcloud --project=${PROJECT} compute health-checks update http ${HEALTH_CHECK_URI} --port=${STATUS_NODE_PORT}
    fi

    # Since JupyterHub uses websockets we want to increase the backend timeout
    echo Increasing backend timeout for JupyterHub
    gcloud --project=${PROJECT} compute backend-services update --global ${BACKEND_SERVICE} --timeout=3600

    echo "Backend updated successfully. Waiting 1 hour before updating again."
    sleep 3600
kind: ConfigMap
metadata:
  name: envoy-config
---
apiVersion: v1
data:
  ingress_bootstrap.sh: |
    #!/usr/bin/env bash

    set -x
    set -e

    # This is a workaround until this is resolved: https://github.com/kubernetes/ingress-gce/pull/388
    # The long-term solution is to use a managed SSL certificate on GKE once the feature is GA.

    # The ingress is initially created without a tls spec.
    # Wait until cert-manager generates the certificate using the http-01 challenge on the GCLB ingress.
    # After the certificate is obtained, patch the ingress with the tls spec to enable SSL on the GCLB.

    # Wait for certificate.
    until kubectl -n ${NAMESPACE} get secret ${TLS_SECRET_NAME} 2>/dev/null; do
      echo "Waiting for certificate..."
      sleep 2
    done

    kubectl -n ${NAMESPACE} patch ingress ${INGRESS_NAME} --type='json' -p '[{"op": "add", "path": "/spec/tls", "value": [{"secretName": "'${TLS_SECRET_NAME}'", "hosts":["'${TLS_HOST_NAME}'"]}]}]'

    echo "Done"
kind: ConfigMap
metadata:
  name: ingress-bootstrap-config
---
`)
	th.writeF("/manifests/gcp/iap-ingress/base/deployment.yaml", `
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: whoami-app
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
      - env:
        - name: PORT
          value: "8081"
        image: gcr.io/cloud-solutions-group/esp-sample-app:1.0.0
        name: app
        ports:
        - containerPort: 8081
        readinessProbe:
          failureThreshold: 2
          httpGet:
            path: /healthz
            port: 8081
            scheme: HTTP
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 5
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: iap-enabler
spec:
  replicas: 1
  template:
    metadata:
      labels:
        service: iap-enabler
    spec:
      containers:
      - command:
        - bash
        - /var/envoy-config/setup_backend.sh
        env:
        - name: NAMESPACE
          value: $(istioNamespace)
        - name: SERVICE
          value: istio-ingressgateway
        - name: INGRESS_NAME
          value: $(ingressName)
        - name: ENVOY_ADMIN
          value: http://localhost:8001
        - name: GOOGLE_APPLICATION_CREDENTIALS
          value: /var/run/secrets/sa/admin-gcp-sa.json
        - name: USE_ISTIO
          value: "true"
        image: gcr.io/kubeflow-images-public/ingress-setup:latest
        name: iap
        volumeMounts:
        - mountPath: /var/envoy-config/
          name: config-volume
        - mountPath: /var/run/secrets/sa
          name: sa-key
          readOnly: true
      restartPolicy: Always
      serviceAccountName: envoy
      volumes:
      - configMap:
          name: envoy-config
        name: config-volume
      - name: sa-key
        secret:
          secretName: $(adminSaSecretName)
`)
	th.writeF("/manifests/gcp/iap-ingress/base/ingress.yaml", `
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    certmanager.k8s.io/issuer: $(issuer)
    ingress.kubernetes.io/ssl-redirect: "true"
    kubernetes.io/ingress.global-static-ip-name: $(ipName)
    kubernetes.io/tls-acme: "true"
  name: envoy-ingress
spec:
  rules:
  - host: $(hostname)
    http:
      paths:
      - backend:
          serviceName: istio-ingressgateway
          servicePort: 80
        path: /*
`)
	th.writeF("/manifests/gcp/iap-ingress/base/job.yaml", `
apiVersion: batch/v1
kind: Job
metadata:
  name: ingress-bootstrap
spec:
  template:
    spec:
      containers:
      - command:
        - /var/ingress-config/ingress_bootstrap.sh
        env:
        - name: NAMESPACE
          valueFrom:
            configMapKeyRef:
              name: parameters
              key: istioNamespace
        - name: TLS_SECRET_NAME
          valueFrom:
            configMapKeyRef:
              name: parameters
              key: tlsSecretName
        - name: TLS_HOST_NAME
          valueFrom:
            configMapKeyRef:
              name: parameters
              key: hostname
        - name: INGRESS_NAME
          valueFrom:
            configMapKeyRef:
              name: parameters
              key: ingressName
        image: gcr.io/kubeflow-images-public/ingress-setup:latest
        name: bootstrap
        volumeMounts:
        - mountPath: /var/ingress-config/
          name: ingress-config
      restartPolicy: OnFailure
      serviceAccountName: envoy
      volumes:
      - configMap:
          defaultMode: 493
          name: ingress-bootstrap-config
        name: ingress-config
`)
	th.writeF("/manifests/gcp/iap-ingress/base/policy.yaml", `
apiVersion: authentication.istio.io/v1alpha1
kind: Policy
metadata:
  name: ingress-jwt
spec:
  origins:
  - jwt:
      audiences:
      - TO_BE_PATCHED
      issuer: https://cloud.google.com/iap
      jwksUri: https://www.gstatic.com/iap/verify/public_key-jwk
      jwtHeaders:
      - x-goog-iap-jwt-assertion
      trigger_rules:
      - excluded_paths:
        - exact: /healthz
        - prefix: /.well-known/acme-challenge
  principalBinding: USE_ORIGIN
  targets:
  - name: istio-ingressgateway
    ports:
    - number: 80
`)
	th.writeF("/manifests/gcp/iap-ingress/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: envoy
`)
	th.writeF("/manifests/gcp/iap-ingress/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: whoami
  name: whoami-app
spec:
  ports:
  - port: 80
    targetPort: 8081
  selector:
    app: whoami
  type: ClusterIP
`)
	th.writeF("/manifests/gcp/iap-ingress/base/stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    service: backend-updater
  name: backend-updater
spec:
  serviceName: backend-updater
  selector:
    matchLabels:
      service: backend-updater
  template:
    metadata:
      labels:
        service: backend-updater
    spec:
      containers:
      - command:
        - bash
        - /var/envoy-config/update_backend.sh
        env:
        - name: NAMESPACE
          value: $(istioNamespace)
        - name: SERVICE
          value: istio-ingressgateway
        - name: GOOGLE_APPLICATION_CREDENTIALS
          value: /var/run/secrets/sa/admin-gcp-sa.json
        - name: INGRESS_NAME
          value: $(ingressName)
        - name: USE_ISTIO
          value: "true"
        image: gcr.io/kubeflow-images-public/ingress-setup:latest
        name: backend-updater
        volumeMounts:
        - mountPath: /var/envoy-config/
          name: config-volume
        - mountPath: /var/run/secrets/sa
          name: sa-key
          readOnly: true
      serviceAccountName: envoy
      volumes:
      - configMap:
          name: envoy-config
        name: config-volume
      - name: sa-key
        secret:
          secretName: admin-gcp-sa
  volumeClaimTemplates: []
`)
	th.writeF("/manifests/gcp/iap-ingress/base/params.yaml", `
varReference:
- path: metadata/name
  kind: Certificate
- path: spec/origins/jwt/issuer
  kind: Policy
- path: metadata/annotations/getambassador.io\/config
  kind: Service
- path: spec/dnsNames
  kind: Certificate
- path: spec/issuerRef/name
  kind: Certificate
- path: metadata/annotations/kubernetes.io\/ingress.global-static-ip-name
  kind: Ingress
- path: spec/commonName
  kind: Certificate
- path: spec/secretName
  kind: Certificate
- path: spec/acme/config/domains
  kind: Certificate
- path: spec/acme/config/http01/ingress
  kind: Certificate
- path: metadata/name
  kind: Ingress
- path: spec/rules/host
  kind: Ingress
- path: metadata/annotations/certmanager.k8s.io\/issuer
  kind: Ingress
- path: spec/template/spec/volumes/secret/secretName
  kind: Deployment
- path: spec/template/spec/volumes/secret/secretName
  kind: StatefulSet
- path: metadata/name
  kind: CloudEndpoint
- path: spec/project
  kind: CloudEndpoint
- path: spec/targetIngress/name
  kind: CloudEndpoint
- path: spec/targetIngress/namespace
  kind: CloudEndpoint
- path: spec/iap/oauthclientCredentials/secretName
  kind: BackendConfig
- path: data/healthcheck_route.yaml
  kind: ConfigMap
`)
	th.writeF("/manifests/gcp/iap-ingress/base/params.env", `
namespace=kubeflow
appName=kubeflow
hostname=
ingressName=envoy-ingress
ipName=
issuer=letsencrypt-prod
oauthSecretName=kubeflow-oauth
project=
adminSaSecretName=admin-gcp-sa
tlsSecretName=envoy-ingress-tls
istioNamespace=istio-system
`)
	th.writeK("/manifests/gcp/iap-ingress/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- backend-config.yaml
- certificate.yaml
- cloud-endpoint.yaml
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- deployment.yaml
- ingress.yaml
- job.yaml
- policy.yaml
- service-account.yaml
- service.yaml
- stateful-set.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: iap-ingress
images:
  - name: gcr.io/kubeflow-images-public/envoy
    newName: gcr.io/kubeflow-images-public/envoy
    newTag: v20180309-0fb4886b463698702b6a08955045731903a18738
  - name: gcr.io/kubeflow-images-public/ingress-setup
    newName: gcr.io/kubeflow-images-public/ingress-setup
    newTag: latest
  - name: gcr.io/cloud-solutions-group/esp-sample-app
    newName: gcr.io/cloud-solutions-group/esp-sample-app
    newTag: 1.0.0
configMapGenerator:
- name: parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: namespace
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
- name: appName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.appName
- name: hostname
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.hostname
- name: ipName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ipName
- name: ingressName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.ingressName
- name: issuer
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.issuer
- name: oauthSecretName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.oauthSecretName
- name: project
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.project
- name: adminSaSecretName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.adminSaSecretName
- name: tlsSecretName
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.tlsSecretName
- name: istioNamespace
  objref:
    kind: ConfigMap
    name: parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.istioNamespace
configurations:
- params.yaml
`)
}

func TestIapIngressBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/gcp/iap-ingress/base")
	writeIapIngressBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../gcp/iap-ingress/base"
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
	n, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := n.EncodeAsYaml()
	th.assertActualEqualsExpected(m, string(expected))
}
