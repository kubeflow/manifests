package jupyter_web_app

import (
	"github.com/kubeflow/manifests/tests"
	"testing"
)

func TestKustomize(t *testing.T) {
	testCase := &tests.KustomizeTestCase{
		Package:  "../../../../../stacks/aws/application/jupyter-web-app",
		Expected: "test_data/expected",
	}

	tests.RunTestCase(t, testCase)
}
