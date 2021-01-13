"""Common reusable steps for kfctl go testing."""
import datetime
import json
import logging
import os
import tempfile
import urllib
import uuid

import requests
import yaml
from kubeflow.testing import util
from kubeflow.testing.cloudprovider.aws import util as aws_util
from kubeflow.kfctl.testing.util import aws_util as kfctl_aws_util
from kubeflow.testing.cloudprovider.aws import prow_artifacts as aws_prow_artifacts
from retrying import retry


# retry 4 times, waiting 3 minutes between retries
@retry(stop_max_attempt_number=4, wait_fixed=180000)
def run_with_retries(*args, **kwargs):
  util.run(*args, **kwargs)

def build_kfctl_go(kfctl_repo_path):
  """build the kfctl go binary and return the path for the same.

  Args:
    kfctl_repo_path (str): Path to kfctl repo.

  Return:
    kfctl_path (str): Path where kfctl go binary has been built.
            will be Kubeflow/kubeflow/bootstrap/bin/kfctl
  """
  kfctl_path = os.path.join(kfctl_repo_path, "bin", "kfctl")
  # We need to use retry builds because when building in the test cluster
  # we see intermittent failures pulling dependencies
  run_with_retries(["make", "build-kfctl"], cwd=kfctl_repo_path)
  return kfctl_path

def get_or_create_app_path_and_parent_dir(app_path):
  """Get a valid app_path and parent dir. Create if they are not existing.
  """
  if not app_path:
    logging.info("--app_path not specified")
    stamp = datetime.datetime.now().strftime("%H%M")
    parent_dir = tempfile.gettempdir()
    app_path = os.path.join(
      parent_dir, "kfctl-{0}-{1}".format(stamp,
                         uuid.uuid4().hex[0:4]))
  else:
    parent_dir = os.path.dirname(app_path)

  if not os.path.exists(parent_dir):
    os.makedirs(parent_dir)
  if not os.path.exists(app_path):
    os.makedirs(app_path)

  return app_path, parent_dir

def load_config(config_path):
  """Load specified KFDef.

  Args:
    config_path: Path to a YAML file containing a KFDef object.
    Can be a local path or a URI like
    https://raw.githubusercontent.com/kubeflow/manifests/master/kfdef/kfctl_gcp_iap.yaml
  Returns:
    config_spec: KfDef spec
  """
  url_for_spec = urllib.parse.urlparse(config_path)

  if url_for_spec.scheme in ["http", "https"]:
    data = requests.get(config_path)
    return yaml.load(data.content)
  else:
    with open(config_path, 'r') as f:
      config_spec = yaml.load(f)
      return config_spec

def set_env_for_auth(use_basic_auth):
  logging.info("use_basic_auth=%s", use_basic_auth)
  # Set ENV for basic auth username/password.
  if use_basic_auth:
    # Don't log the password.
    # logging.info("Setting environment variables KUBEFLOW_USERNAME and KUBEFLOW_PASSWORD")
    os.environ["KUBEFLOW_USERNAME"] = "kf-test-user"
    os.environ["KUBEFLOW_PASSWORD"] = str(uuid.uuid4().hex)
  else:
    # Owned by project kubeflow-ci-deployment.
    logging.info("Setting environment variables CLIENT_SECRET and CLIENT_ID")
    os.environ["CLIENT_SECRET"] = "CJ4qVPLTi0j0GJMkONj7Quwt"
    os.environ["CLIENT_ID"] = (
      "29647740582-7meo6c7a9a76jvg54j0g2lv8lrsb4l8g"
      ".apps.googleusercontent.com")

def set_env_init_args(config_spec):
  gcp_plugin = {}
  for plugin in config_spec.get("spec", {}).get("plugins", []):
    if plugin.get("kind", "") == "KfGcpPlugin":
      gcp_plugin = plugin
      break
  use_basic_auth = gcp_plugin.get("spec", {}).get("useBasicAuth", False)
  set_env_for_auth(use_basic_auth)

def write_basic_auth_login(filename):
  """Read basic auth login from ENV and write to the filename given. If username/password
  cannot be found in ENV, this function will silently return.

  Args:
    filename: The filename (directory/file name) the login is writing to.
  """
  username = os.environ.get("KUBEFLOW_USERNAME", "")
  password = os.environ.get("KUBEFLOW_PASSWORD", "")

  if not username or not password:
    return

  with open(filename, "w") as f:
    login = {
        "username": username,
        "password": password,
    }
    json.dump(login, f)

def filter_spartakus(spec):
  """Filter our Spartakus from KfDef spec.

  Args:
    spec: KfDef spec

  Returns:
    spec: Filtered KfDef spec
  """
  for i, app in enumerate(spec["applications"]):
    if app["name"] == "spartakus":
      spec["applications"].pop(i)
      break
  return spec

def get_config_spec(config_path, app_path, cluster_name):
  """Generate KfDef spec.

  Args:
    config_path: Path to a YAML file containing a KFDef object.
    Can be a local path or a URI like
    https://raw.githubusercontent.com/kubeflow/manifests/master/kfdef/kfctl_gcp_iap.yaml
    app_path: The path to the Kubeflow app.
    cluster_name: Name of EKS cluster
  Returns:
    config_spec: Updated KfDef spec
  """
  # TODO(https://github.com/kubeflow/kubeflow/issues/2831): Once kfctl
  # supports loading version from a URI we should use that so that we
  # pull the configs from the repo we checked out.
  config_spec = load_config(config_path)

  repos = config_spec["spec"]["repos"]
  manifests_repo_name = "manifests"
  if os.getenv("REPO_NAME") == manifests_repo_name:
    # kfctl_go_test.py was triggered on presubmit from the kubeflow/manifests
    # repository. In this case we want to use the specified PR of the
    # kubeflow/manifests repository; so we need to change the repo specification
    # in the KFDef spec.
    # TODO(jlewi): We should also point to a specific commit when triggering
    # postsubmits from the kubeflow/manifests repo
    for repo in repos:
      if repo["name"] !=  manifests_repo_name:
        continue

      version = None

      if os.getenv("PULL_PULL_SHA"):
        # Presubmit
        version = os.getenv("PULL_PULL_SHA")

      # See https://github.com/kubernetes/test-infra/blob/45246b09ed105698aa8fb928b7736d14480def29/prow/jobs.md#job-environment-variables  # pylint: disable=line-too-long
      elif os.getenv("PULL_BASE_SHA"):
        version = os.getenv("PULL_BASE_SHA")

      if version:
        repo["uri"] = ("https://github.com/kubeflow/manifests/archive/"
                       "{0}.tar.gz").format(version)
        logging.info("Overwriting the URI")
      else:
        # Its a periodic job so use whatever value is set in the KFDef
        logging.info("Not overwriting manifests version")
    logging.info(str(config_spec))
  return config_spec

def kfctl_deploy_kubeflow(app_path, config_path, kfctl_path, build_and_apply, cluster_name):
  """Deploy kubeflow.

  Args:
  app_path: The path to the Kubeflow app.
  config_path: Path to the KFDef spec file.
  kfctl_path: Path to the kfctl go binary
  build_and_apply: whether to build and apply or apply
  cluster_name: Name of EKS cluster
  Returns:
  app_path: Path where Kubeflow is installed
  """
  # build_and_apply is a boolean used for testing both the new semantics
  # test case 1: build_and_apply
  # kfctl build -f <config file>
  # kfctl apply
  # test case 2: apply
  # kfctl apply -f <config file>

  kfctl_aws_util.aws_auth_load_kubeconfig(cluster_name)

  if not os.path.exists(kfctl_path):
    msg = "kfctl Go binary not found: {path}".format(path=kfctl_path)
    logging.error(msg)
    raise RuntimeError(msg)

  app_path, parent_dir = get_or_create_app_path_and_parent_dir(app_path)

  logging.info("app path %s", app_path)
  logging.info("kfctl path %s", kfctl_path)

  config_spec = get_config_spec(config_path, app_path, cluster_name)
  with open(os.path.join(app_path, "tmp.yaml"), "w") as f:
    yaml.dump(config_spec, f)

  # build_and_apply
  logging.info("running kfctl with build and apply: %s \n", build_and_apply)

  logging.info("switching working directory to: %s \n", app_path)
  os.chdir(app_path)

  # push newly built kfctl to S3
  push_kfctl_to_s3(kfctl_path)

  # Workaround to fix issue
  # msg="Encountered error applying application bootstrap:  (kubeflow.error): Code 500 with message: Apply.Run
  # : error when creating \"/tmp/kout927048001\": namespaces \"kubeflow-test-infra\" not found"
  # filename="kustomize/kustomize.go:266"
  # TODO(PatrickXYS): fix the issue permanentely rather than work-around
  util.run(["kubectl", "create", "namespace", "kubeflow-test-infra"])

  # Do not run with retries since it masks errors
  logging.info("Running kfctl with config:\n%s", yaml.safe_dump(config_spec))
  if build_and_apply:
    build_and_apply_kubeflow(kfctl_path, app_path)
  else:
    apply_kubeflow(kfctl_path, app_path)
  return app_path

def push_kfctl_to_s3(kfctl_path):
  bucket = aws_prow_artifacts.AWS_PROW_RESULTS_BUCKET
  logging.info("Bucket name: %s", aws_prow_artifacts.get_s3_dir(bucket))
  s3_path = os.path.join(aws_prow_artifacts.get_s3_dir(bucket) + "/artifacts/build_bin/kfctl")
  aws_util.upload_file_to_s3(kfctl_path, s3_path)

def apply_kubeflow(kfctl_path, app_path):
  util.run([kfctl_path, "apply", "-V", "-f=" + os.path.join(app_path, "tmp.yaml")], cwd=app_path)
  return app_path

def build_and_apply_kubeflow(kfctl_path, app_path):
  util.run([kfctl_path, "build", "-V", "-f=" + os.path.join(app_path, "tmp.yaml")], cwd=app_path)
  util.run([kfctl_path, "apply", "-V", "-f=" + os.path.join(app_path, "tmp.yaml")], cwd=app_path)
  return app_path

def upgrade_kubeflow(kfctl_path, parent_dir):
  util.run([kfctl_path, "apply", "-V", "-f=" + os.path.join(parent_dir, "upgrade.yaml")], cwd=parent_dir)

def verify_kubeconfig(app_path):
  """Verify kubeconfig.

  Args:
    app_path: KfDef spec path
  """
  name = os.path.basename(app_path)
  context = util.run(["kubectl", "config", "current-context"]).strip()
  if name == context:
    logging.info("KUBECONFIG current context name matches app name: %s", name)
  else:
    msg = "KUBECONFIG not having expected context: {expected} v.s. {actual}".format(
      expected=name, actual=context)
    logging.error(msg)
    raise RuntimeError(msg)

def kfctl_upgrade_kubeflow(app_path, kfctl_path, upgrade_spec_path, use_basic_auth=False):
  """Upgrade kubeflow.

  Args:
  app_path: The path to the Kubeflow app to be upgraded.
  kfctl_path: The path to the kfctl binary.
  upgrade_spec_path: The path to the upgrade sepc.
  use_basic_auth: True if we are using basic auth for GCP.
  """
  if not os.path.exists(kfctl_path):
    msg = "kfctl Go binary not found: {path}".format(path=kfctl_path)
    logging.error(msg)
    raise RuntimeError(msg)

  app_path, parent_dir = get_or_create_app_path_and_parent_dir(app_path)
  app_name = os.path.basename(app_path)

  logging.info("app path %s", app_path)
  logging.info("app name %s", app_name)
  logging.info("parent dir %s", parent_dir)
  logging.info("kfctl path %s", kfctl_path)

  upgrade_spec = load_config(upgrade_spec_path)
  upgrade_spec["spec"]["currentKfDef"]["name"] = app_name
  upgrade_spec["spec"]["newKfDef"]["name"] = app_name

  with open(os.path.join(parent_dir, "upgrade.yaml"), "w") as f:
    yaml.dump(upgrade_spec, f)

  # Set ENV for credentials IAP/basic auth needs.
  set_env_for_auth(use_basic_auth)

  # Write basic auth login username/password to a file for later tests.
  # If the ENVs are not set, this function call will be noop.
  write_basic_auth_login(os.path.join(app_path, "login.json"))

  logging.info("switching working directory to: %s \n", parent_dir)
  os.chdir(parent_dir)

  # Run upgrade
  logging.info("Running kfctl with upgrade spec:\n%s", yaml.safe_dump(upgrade_spec))
  upgrade_kubeflow(kfctl_path, parent_dir)