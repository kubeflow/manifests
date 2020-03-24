package mysqlPd

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../pipeline/mysql/overlays/mysqlPd",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
