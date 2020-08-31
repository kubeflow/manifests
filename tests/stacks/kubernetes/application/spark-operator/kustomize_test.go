package spark_operator

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/kubernetes/application/spark-operator",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
