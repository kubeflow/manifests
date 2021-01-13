""""Define the E2E workflow for kfctl.
Rapid iteration.
Here are some pointers for rapidly iterating on the workflow during development.
1. You can use the e2e_tool.py to directly launch the workflow on a K8s cluster.
   If you don't have CLI access to the kubeflow-ci cluster (most folks) then
   you would need to setup your own test cluster.
2. To avoid redeploying on successive runs set the following parameters
   --app_name=name for kfapp
   --delete_kubeflow=False
   Setting these parameters will cause the same KF deployment to be reused
   across invocations. As a result successive runs won't have to redeploy KF.
Example running with E2E tool
export PYTHONPATH=${PYTHONPATH}:${KUBEFLOW_REPO}/py:${KUBEFLOW_TESTING_REPO}/py
python -m kubeflow.testing.e2e_tool apply \
  kubeflow.kubeflow.ci.kfctl_e2e_workflow.create_workflow
  --name=${USER}-kfctl-test-$(date +%Y%m%d-%H%M%S) \
  --namespace=kubeflow-test-infra \
  --test-endpoint=true \
  --kf-app-name=${KFAPPNAME} \
  --delete-kf=false
  --open-in-chrome=true
We set kf-app-name and delete-kf to false to allow reusing the deployment
across successive runs.
To use code from a pull request set the prow envariables; e.g.
export JOB_NAME="jlewi-test"
export JOB_TYPE="presubmit"
export BUILD_ID=1234
export PROW_JOB_ID=1234
export REPO_OWNER=kubeflow
export REPO_NAME=kubeflow
export PULL_NUMBER=4148
"""

import datetime
from kubeflow.kfctl.testing.ci import kfctl_e2e_workflow
from kubeflow.testing import argo_build_util
import logging
import os
import uuid

UPGRADE_DAG_NAME = "upgrade-dag"

READY_AFTER_UPGRADE = "ready-after-upgrade"

TEMPLATE_LABEL = "kfctl_upgrade_e2e"

class Builder(kfctl_e2e_workflow.Builder):
    def __init__(self,
                 upgrade_spec_path=("https://raw.githubusercontent.com/kubeflow"
                                    "/manifests/v0.7-branch/kfdef/kfctl_upgrade_gcp_iap_0.7.1.yaml"),
                 **kwargs):
        """Initialize a builder."""
        super(Builder, self).__init__(**kwargs)
        self.upgrade_spec_path = upgrade_spec_path

    def _build_upgrade_dag(self):
        """Build the dag of steps to run upgrade."""

        # Add the dag to the workflow.
        self.workflow["spec"]["templates"].append({
           "dag": {
                "tasks": [],
                },
           "name": UPGRADE_DAG_NAME,
          })
        task_template = self._build_task_template()

        # Change the workfing directory for all subsequent steps
        task_template["container"]["workingDir"] = os.path.join(self.kfctl_pytest_dir)

        #***************************************************************************
        # Run kfctl upgrade
        step_name = "upgrade"
        command = [
           "pytest",
           "kfctl_upgrade_test.py",
           "-s",
           "--log-cli-level=info",
           "--junitxml=" + os.path.join(self.artifacts_dir,
                                        "junit_kfctl-upgrade-test-" +
                                        self.config_name + ".xml"),
           # Test suite name needs to be unique based on parameters
           "-o", "junit_suite_name=test_kfctl_upgrade_" + self.config_name,
           "--app_path=" + self.app_dir,
           "--kfctl_path=" + self.kfctl_path,
           "--upgrade_spec_path=" + self.upgrade_spec_path,
         ]
        dependences = []
        upgrade_step = self._build_step(step_name, self.workflow, UPGRADE_DAG_NAME, task_template,
                                        command, dependences)
    def build(self):
        self.workflow = super(Builder, self).build()
        task_template = self._build_task_template()

        # Change the workflow_template labels
        self.workflow["metadata"]["labels"]["workflow_template"] = TEMPLATE_LABEL

        # Add the dag to upgrade Kubeflow to the workflow
        self._build_upgrade_dag()

        # Add a task to the E2E dag to run the dag to upgrade Kubeflow.
        dependencies = [self._run_tests_step_name]

        if self._test_endpoint_step_name:
            dependencies.append(self._test_endpoint_step_name)

        step_name = UPGRADE_DAG_NAME
        template_name = UPGRADE_DAG_NAME
        argo_build_util.add_task_only_to_dag(self.workflow, kfctl_e2e_workflow.UNIT_DAG_NAME, step_name,
                                             template_name, dependencies)

        # Wait for Kubeflow to be ready after upgrading
        step_name = READY_AFTER_UPGRADE
        template_name = "kubeflow-is-ready"
        command = [
           "pytest",
           "kf_is_ready_test.py",
           "-s",
           "--log-cli-level=info",
           "--junitxml=" + os.path.join(self.artifacts_dir,
                                        "junit_ready-after-upgrade-test-" +
                                        self.config_name + ".xml"),
           "-o", "junit_suite_name=test_ready_after_upgrade_" + self.config_name,
           "--app_path=" + self.app_dir,
         ]

        dependencies = [UPGRADE_DAG_NAME]
        argo_build_util.add_task_only_to_dag(self.workflow, kfctl_e2e_workflow.UNIT_DAG_NAME, step_name,
                                             template_name, dependencies)
   
        #****************************************************************************
        # Add tests DAG
        #****************************************************************************
        # After running upgrade we want to rerun the DAG(s) that validate the deployment is healthy

        step_name = "test-after-upgrade"
        template_name = kfctl_e2e_workflow.TESTS_DAG_NAME
        dependencies = [READY_AFTER_UPGRADE]
        argo_build_util.add_task_only_to_dag(self.workflow, kfctl_e2e_workflow.UNIT_DAG_NAME, step_name,
                                             template_name, dependencies)

        # Test the endpoint after upgrade
        if self.test_endpoint:
            dependencies = [UPGRADE_DAG_NAME]
            step_name = "upgraded-endpoint-ready"
            argo_build_util.add_task_only_to_dag(self.workflow, kfctl_e2e_workflow.UNIT_DAG_NAME, step_name,
                                                 self._test_endpoint_template_name, dependencies)

        # Reset the labels on all templates to pick up the updated workflow template label
        self.workflow = argo_build_util.set_task_template_labels(self.workflow)

        return self.workflow

# TODO(jlewi): This is an unnecessary layer of indirection around the builder
# We should allow py_func in prow_config to point to the builder and
# let e2e_tool take care of this.
def create_workflow(**kwargs): # pylint: disable=too-many-statements
    """Create workflow returns an Argo workflow to test kfctl upgrades.
    Args:
      name: Name to give to the workflow. This can also be used to name things
       associated with the workflow.
    """

    builder = Builder(**kwargs)

    return builder.build()

