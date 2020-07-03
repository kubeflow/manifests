package katib_external_db

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../katib/installs/katib-external-db",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
