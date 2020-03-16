package tests_test

import (
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
	"testing"
)

func writePrivateutilBase(th *KustTestHarness) {
	th.writeF("/manifests/gcp/privateutil/base/iap-jwt-key.yaml", `
apiVersion: v1
data:
  public_key-jwk: |
    {
       "keys" : [
          {
             "alg" : "ES256",
             "crv" : "P-256",
             "kid" : "6BEeoA",
             "kty" : "EC",
             "use" : "sig",
             "x" : "lmi1hJdqtbvdX1INOf5B9dWvkydYoowHUXiw8ELWzk8",
             "y" : "2BxEja_L10KMjrizhLS2XgkGxZHi1KsWKdbEwKyjbvw"
          },
          {
             "alg" : "ES256",
             "crv" : "P-256",
             "kid" : "2nMJtw",
             "kty" : "EC",
             "use" : "sig",
             "x" : "9e1x7YRZg53A5zIJ0p2ZQ9yTrgPLGIf4ntOk-4O2R28",
             "y" : "q8iDm7nsnpz1xPdrWBtTZSowzciS3O7bMYtFFJ8saYo"
          },
          {
             "alg" : "ES256",
             "crv" : "P-256",
             "kid" : "LYyP2g",
             "kty" : "EC",
             "use" : "sig",
             "x" : "SlXFFkJ3JxMsXyXNrqzE3ozl_0913PmNbccLLWfeQFU",
             "y" : "GLSahrZfBErmMUcHP0MGaeVnJdBwquhrhQ8eP05NfCI"
          },
          {
             "alg" : "ES256",
             "crv" : "P-256",
             "kid" : "mpf0DA",
             "kty" : "EC",
             "use" : "sig",
             "x" : "fHEdeT3a6KaC1kbwov73ZwB_SiUHEyKQwUUtMCEn0aI",
             "y" : "QWOjwPhInNuPlqjxLQyhveXpWqOFcQPhZ3t-koMNbZI"
          },
          {
             "alg" : "ES256",
             "crv" : "P-256",
             "kid" : "b9vTLA",
             "kty" : "EC",
             "use" : "sig",
             "x" : "qCByTAvci-jRAD7uQSEhTdOs8iA714IbcY2L--YzynI",
             "y" : "WQY0uCoQyPSozWKGQ0anmFeOH5JNXiZa9i6SNqOcm7w"
          }
       ]
    }
kind: ConfigMap
metadata:
  name: pubkey
  namespace: istio-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pubkey
  namespace: istio-system
spec:
  replicas: 1
  selector:
    matchLabels:
      service: pubkey
  template:
    metadata:
      labels:
        service: pubkey
    spec:
      containers:
        - command:
          - /pub-key-server
          image: gcr.io/kubeflow-images-public/jwtpubkey:v20200311-v0.7.0-rc.5-109-g641fb40b-dirty-eb1cdc
          name: pubkey
          volumeMounts:
            - mountPath: /var/pubkey/
              name: config-volume
      restartPolicy: Always
      serviceAccountName: kf-admin
      volumes:
        - configMap:
            name: pubkey
          name: config-volume
---
apiVersion: v1
kind: Service
metadata:
  name: pubkey
  namespace: istio-system
spec:
  ports:
    - port: 8087
  selector:
    service: pubkey`)
	th.writeK("/manifests/gcp/privateutil/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- iap-jwt-key.yaml
`)
}

func TestPrivateutilBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/gcp/privateutil/base")
	writePrivateutilBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../gcp/privateutil/base"
	fsys := fs.MakeRealFS()
	lrc := loader.RestrictionRootOnly
	_loader, loaderErr := loader.NewLoader(lrc, validators.MakeFakeValidator(), targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pc := plugins.DefaultPluginConfig()
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl(), plugins.NewLoader(pc, rf))
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
