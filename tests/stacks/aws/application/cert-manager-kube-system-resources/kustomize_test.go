package cert_manager_kube_system_resources

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/aws/application/cert-manager-kube-system-resources",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
