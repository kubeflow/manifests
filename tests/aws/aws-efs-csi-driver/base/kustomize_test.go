package base

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package: "../../../../aws/aws-efs-csi-driver/base",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}