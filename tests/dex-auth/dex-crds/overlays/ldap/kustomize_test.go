package ldap

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../dex-auth/dex-crds/overlays/ldap",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
