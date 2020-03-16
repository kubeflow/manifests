package oidc

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package: "../../../../../aws/istio-ingress/overlays/oidc",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}