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

export PYTHONPATH=${PYTHONPATH}:${KFCTL_REPO}/py:${KUBEFLOW_TESTING_REPO}/py

python -m kubeflow.testing.e2e_tool apply \
  kubeflow.kfctl.testing.ci.kfctl_e2e_workflow.create_workflow
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

from kubeflow.testing import argo_build_util
from kubeflow.testing import util
import logging
import os
import uuid

# The name of the NFS volume claim to use for test files.
NFS_VOLUME_CLAIM = "nfs-external"
# The name to use for the volume to use to contain test data
DATA_VOLUME = "kubeflow-test-volume"

# This is the main dag with the entrypoint
UNIT_DAG_NAME = "unit"
EXIT_DAG_NAME = "exit-handler"

TEMPLATE_LABEL = "manifests_unit"

DEFAULT_REPOS = [
    "kubeflow/testing@HEAD",
]

class Builder(object):
  def __init__(self, name=None, namespace="kubeflow-test-infra",
               config_path=("https://raw.githubusercontent.com/kubeflow"
                            "/manifests/master/kfdef/kfctl_istio_dex.v1.1.0.yaml"),
               bucket="aws-kubernetes-jenkins",
               use_basic_auth=False,
               build_and_apply=False,
               test_target_name=None,
               eks_cluster_version="1.17",
               extra_repos="",
               **kwargs):
    """Initialize a builder.
    Args:
      name: Name for the workflow.
      namespace: Namespace for the workflow.
      config_path: Path to the KFDef spec file.
      bucket: The bucket to upload artifacts to. If not set use default determined by prow_artifacts.py.
      test_endpoint: Whether to test the endpoint is ready.
      use_basic_auth: Whether to use basic_auth.
      test_target_name: (Optional) Name to use as the test target to group
        tests.
      delete_kf: (Optional) Don't run the step to delete Kubeflow. Set to
        true if you want to leave the deployment up for some reason.
    """
    self.name = name
    self.namespace = namespace
    self.bucket = bucket if bucket else self.bucket
    self.config_path = config_path
    self.build_and_apply = build_and_apply
    #****************************************************************************
    # Define directory locations
    #****************************************************************************
    # mount_path is the directory where the volume to store the test data
    # should be mounted.
    self.mount_path = "/mnt/" + "test-data-volume"
    # test_dir is the root directory for all data for a particular test run.
    self.test_dir = self.mount_path + "/" + self.name
    # output_dir is the directory to sync to S3 to contain the output for this
    # job.
    self.output_dir = self.test_dir + "/output"
    self.artifactsDir = self.output_dir + "/artifacts"

    # We prefix the artifacts directory with junit because
    # that's what spyglass/prow requires. This ensures multiple
    # instances of a workflow triggered by the same prow job
    # don't end up clobbering each other
    self.artifacts_dir = self.output_dir + "/artifacts/junit_{0}".format(name)

    # source directory where all repos should be checked out
    self.src_root_dir = self.test_dir + "/src"
    # The directory containing the kubeflow/kfctl repo
    self.src_dir = self.src_root_dir + "/kubeflow/manifests"
    self.kubeflow_dir = self.src_root_dir + "/kubeflow/kubeflow"
    self.manifests_test_dir = self.src_dir + "/tests"

    # Directory in kubeflow/kfctl containing the pytest files.
    self.kfctl_pytest_dir = os.path.join(self.src_dir, "py/kubeflow/kfctl/testing/pytests")

    # Top level directories for python testing code in kfctl.
    self.kfctl_py = os.path.join(self.src_dir, "py")

    # Build a string of key value pairs that can be passed to various test
    # steps to allow them to do substitution into different values.
    values = {
      "srcrootdir": self.src_root_dir,
    }

    value_pairs = ["{0}={1}".format(k,v) for k,v in values.items()]
    self.values_str = ",".join(value_pairs)

    # The directory within the kubeflow_testing submodule containing
    # py scripts to use.
    self.kubeflow_testing_py = self.src_root_dir + "/kubeflow/testing/py"

    self.tf_operator_root = os.path.join(self.src_root_dir,
                                         "kubeflow/tf-operator")
    self.tf_operator_py = os.path.join(self.tf_operator_root, "py")

    self.go_path = self.test_dir

    # Name for the Kubeflow app.
    # This needs to be unique for each test run because it is
    # used to name AWS resources
    # TODO(jlewi): Might be good to include pull number or build id in the name
    # Not sure if being non-deterministic is a good idea.
    # A better approach might be to hash the workflow name to generate a unique
    # name dependent on the workflow name. We know there will be one workflow
    # per cluster.
    self.uuid = uuid.uuid4().hex[0:8]

    # Name for ephemeral EKS cluster
    self.cluster_name = "eks-cluster-" + self.uuid

    # Version for ephemeral EKS clsuter
    self.eks_cluster_version = eks_cluster_version

    # Config name is the name of the config file. This is used to give junit
    # files unique names.
    self.config_name = os.path.splitext(os.path.basename(config_path))[0]

    # The class name to label junit files.
    # We want to be able to group related tests in test grid.
    # Test grid allows grouping by target which corresponds to the classname
    # attribute in junit files.
    # So we set an environment variable to the desired class name.
    # The pytest modules can then look at this environment variable to
    # explicitly override the classname.
    # The classname should be unique for each run so it should take into
    # account the different parameters
    if test_target_name:
      self.test_target_name = test_target_name
    else:
      self.test_target_name = self.config_name

    # app_name is the name of the Kubeflow deployment.
    # This needs to be unique per run since we name AWS resources with it.
    self.app_name = self.cluster_name

    # AWS service accounts can only be max 100 characters. Service account names
    # are generated by taking the app_name and appending suffixes like "user"
    # and "admin"
    if len(self.app_name) > 99:
      raise ValueError(("app_name {0} is longer than 100 characters; this will"
                        "likely exceed AWS naming restrictions.").format(
                          self.app_name))
    # Directory for the KF app.
    self.app_dir = os.path.join(self.test_dir, self.app_name)
    self.use_basic_auth = use_basic_auth

    # The name space we create KF artifacts in; e.g. TFJob and notebooks.
    # TODO(jlewi): These should no longer be running the system namespace but
    # should move into the namespace associated with the default profile.
    self.steps_namespace = "kubeflow"

    self.kfctl_path = os.path.join(self.src_dir, "bin/kfctl")

    # Fetch the main repo from Prow environment.
    self.main_repo = argo_build_util.get_repo_from_prow_env()

    # extra_repos is a list of comma separated repo names with commits,
    # in the format <repo_owner>/<repo_name>@<commit>,
    # e.g. "kubeflow/tf-operator@12345,kubeflow/manifests@23456".
    # This will be used to override the default repo branches.
    self.extra_repos = []
    if extra_repos:
      self.extra_repos = extra_repos.split(',')


  def _build_workflow(self):
    """Create the scaffolding for the Argo workflow"""
    workflow = {
      "apiVersion": "argoproj.io/v1alpha1",
      "kind": "Workflow",
      "metadata": {
        "name": self.name,
        "namespace": self.namespace,
        "labels": argo_build_util.add_dicts([{
            "workflow": self.name,
            "workflow_template": TEMPLATE_LABEL,
          }, argo_build_util.get_prow_labels()]),
      },
      "spec": {
        "entrypoint": UNIT_DAG_NAME,
        # Have argo garbage collect old workflows otherwise we overload the API
        # server.
        "ttlSecondsAfterFinished": 7 * 24 * 60 * 60,
        "volumes": [
          {
            "name": DATA_VOLUME,
            "persistentVolumeClaim": {
              "claimName": NFS_VOLUME_CLAIM,
            },
          },
        ],
        "onExit": EXIT_DAG_NAME,
        "templates": [
          {
           "dag": {
                "tasks": [],
                },
           "name": UNIT_DAG_NAME,
          },
          {
            "dag": {
              "tasks": [],
              },
              "name": EXIT_DAG_NAME,
            }
        ],
      },  # spec
    } # workflow

    return workflow

  def _build_task_template(self):
    """Return a template for all the tasks"""

    task_template = {'activeDeadlineSeconds': 3000,
     'container': {'command': [],
      'env': [
        {
          "name": "AWS_ACCESS_KEY_ID",
          "valueFrom": {
            "secretKeyRef": {
              "name": "aws-credentials",
              "key": "AWS_ACCESS_KEY_ID",
            },
          },
        },
        {
          "name": "AWS_SECRET_ACCESS_KEY",
          "valueFrom": {
            "secretKeyRef": {
              "name": "aws-credentials",
              "key": "AWS_SECRET_ACCESS_KEY",
            },
          },
        },
        {
          "name": "AWS_DEFAULT_REGION",
          "value": "us-west-2",
        },
        {
          "name": "GITHUB_TOKEN",
          "valueFrom": {
            "secretKeyRef": {
              "name": "github-token",
              "key": "github_token",
            },
          },
        },
        {"name": "TEST_TARGET_NAME",
         "value": self.test_target_name},
       ],
      'image': '527798164940.dkr.ecr.us-west-2.amazonaws.com/aws-kubeflow-ci/test-worker:v1.2-branch',
      'imagePullPolicy': 'Always',
      'name': '',
      'resources': {'limits': {'cpu': '4', 'memory': '4Gi'},
       'requests': {'cpu': '1', 'memory': '1536Mi'}},
      'volumeMounts': [{'mountPath': '/mnt/test-data-volume',
        'name': 'kubeflow-test-volume'}]},
     'metadata': {'labels': {
       'workflow_template': TEMPLATE_LABEL}},
     'outputs': {}}

    # Define common environment variables to be added to all steps
    common_env = [
      {'name': 'PYTHONPATH',
       'value': ":".join([self.kubeflow_testing_py,
                          self.kfctl_py,
                          self.tf_operator_py])},
      {'name': 'GOPATH',
        'value': self.go_path},
    ]

    task_template["container"]["env"].extend(common_env)

    task_template = argo_build_util.add_prow_env(task_template)

    return task_template

  def _build_step(self, name, workflow, dag_name, task_template,
                  command, dependences):
    """Syntactic sugar to add a step to the workflow"""

    step = argo_build_util.deep_copy(task_template)

    step["name"] = name
    step["container"]["command"] = command

    return argo_build_util.add_task_to_dag(workflow, dag_name, step, dependences)

  def build(self):
    self.workflow = self._build_workflow()
    task_template = self._build_task_template()
    py3_template = argo_build_util.deep_copy(task_template)
    py3_template["container"]["image"] = "527798164940.dkr.ecr.us-west-2.amazonaws.com/aws-kubeflow-ci/test-worker:v1.2-branch"

    #**************************************************************************
    # Checkout
    # create the checkout step

    checkout = argo_build_util.deep_copy(task_template)

    # Construct the list of repos to checkout
    list_of_repos = DEFAULT_REPOS
    list_of_repos.append(self.main_repo)
    list_of_repos.extend(self.extra_repos)
    repos = util.combine_repos(list_of_repos)
    repos_str = ','.join(['%s@%s' % (key, value) for (key, value) in repos.items()])


    # If we are using a specific branch (e.g. periodic tests for release branch)
    # then we need to use depth = all; otherwise checkout out the branch
    # will fail. Otherwise we checkout with depth=30. We want more than
    # depth=1 because the depth will determine our ability to find the common
    # ancestor which affects our ability to determine which files have changed
    depth = 30
    if os.getenv("BRANCH_NAME"):
      logging.info("BRANCH_NAME=%s; setting detph=all",
                   os.getenv("BRANCH_NAME"))
      depth = "all"

    checkout["name"] = "checkout"
    checkout["container"]["command"] = ["/usr/local/bin/checkout_repos.sh",
                                        "--repos=" + repos_str,
                                        "--depth={0}".format(depth),
                                        "--src_dir=" + self.src_root_dir]

    argo_build_util.add_task_to_dag(self.workflow, UNIT_DAG_NAME, checkout, [])

    # Change the working directory for all subsequent steps
    task_template["container"]["workingDir"] = os.path.join(
      self.manifests_test_dir)
    py3_template["container"]["workingDir"] = os.path.join(self.kfctl_pytest_dir)

    #***************************************************************************
    # create_pr_symlink
    #***************************************************************************
    # TODO(jlewi): run_e2e_workflow.py should probably create the PR symlink
    step_name = "create-pr-symlink"
    command = ["python",
               "-m",
               "kubeflow.testing.cloudprovider.aws.prow_artifacts",
               "--artifacts_dir=" + self.output_dir,
               "create_pr_symlink_s3",
               "--bucket=" + self.bucket]

    dependences = [checkout["name"]]
    symlink = self._build_step(step_name, self.workflow, UNIT_DAG_NAME, task_template,
                               command, dependences)

    #***************************************************************************
    # mnaifests unit tests
    #***************************************************************************
    step_name = "mnaifests-unittests"
    command = ["make",
               "test",
               ]

    dependences = [checkout["name"]]
    kfctl_go_unittests = self._build_step(step_name, self.workflow, UNIT_DAG_NAME,
                                          task_template, command, dependences)

    # Set the labels on all templates
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