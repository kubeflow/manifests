package ibm_storage_config

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package: "../../../../../katib/katib-controller/overlays/ibm-storage-config",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}