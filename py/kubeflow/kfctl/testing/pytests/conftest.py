import pytest

def pytest_addoption(parser):
  parser.addoption(
      "--app_path", action="store", default="",
      help="Path where the KF application should be stored")

  parser.addoption(
      "--app_name", action="store", default="",
      help="Name of the KF application")

  parser.addoption(
      "--cluster_name", action="store", default="",
      help="Name of ephemeral EKS cluster for e2e test.")

  parser.addoption(
      "--kfctl_path", action="store", default="",
      help="Path to kfctl.")

  parser.addoption(
      "--kfctl_repo_path", action="store", default="",
      help="Path to kfctl repo.")

  parser.addoption(
      "--namespace", action="store", default="kubeflow",
      help="Namespace to use.")

  parser.addoption(
      "--project", action="store", default="kubeflow-ci-deployment",
      help="GCP project to deploy Kubeflow to")

  parser.addoption(
      "--config_path", action="store", default="",
      help=("The config to use for kfctl init. The path can use python style "
            "go format strings; e.g. "
            "--config_path=/{srcdir}/kubeflow/manifests/kfdef/kfctl_gcp.yaml"
            " The values should be supplied by --values."))

  parser.addoption(
      "--values", action="store", default="",
      help=("A comma separated list of key value pairs to be used for "
            "subsitution in --config_path"))

  parser.addoption(
      "--build_and_apply", action="store", default="False",
      help="Whether to build and apply or apply in kfctl"
  )
  # TODO(jlewi): This flag is deprecated this should be determined now from the KFDef spec.
  parser.addoption(
      "--use_basic_auth", action="store", default="False",
      help="Use basic auth.")

  # TODO(jlewi): This flag is deprecated this should be determined now from the KFDef spec
  parser.addoption(
      "--use_istio", action="store", default="True",
      help="Use istio.")

  parser.addoption(
      "--eks_cluster_version", action="store", default="",
      help="Version of ephemeral EKS cluster")

  parser.addoption(
      "--cluster_creation_script", action="store", default="",
      help="The script to use to create a K8s cluster before running kfctl.")

  parser.addoption(
      "--cluster_deletion_script", action="store", default="",
      help="The script to use to delete a K8s cluster before running kfctl.")

  parser.addoption(
      "--self_signed_cert", action="store", default="False",
      help="Whether to use self-signed cert for ingress.")

  parser.addoption(
      "--upgrade_spec_path", action="store", default="",
      help=("The spec for upgrading Kubeflow. The path can use python style "
            "go format strings; e.g. "
            "--upgrade_spec_path=/{srcdir}/kubeflow/manifests/kfdef/kfctl_gcp_upgrade.yaml"))

@pytest.fixture
def app_path(request):
  return request.config.getoption("--app_path")

@pytest.fixture
def app_name(request):
  return request.config.getoption("--app_name")

@pytest.fixture
def kfctl_path(request):
  return request.config.getoption("--kfctl_path")

@pytest.fixture
def kfctl_repo_path(request):
  return request.config.getoption("--kfctl_repo_path")

@pytest.fixture
def namespace(request):
  return request.config.getoption("--namespace")

@pytest.fixture
def project(request):
  return request.config.getoption("--project")

@pytest.fixture
def config_path(request):
  return request.config.getoption("--config_path")

@pytest.fixture
def values(request):
  return request.config.getoption("--values")

@pytest.fixture
def eks_cluster_version(request):
  return request.config.getoption("--eks_cluster_version")

@pytest.fixture
def cluster_creation_script(request):
  return request.config.getoption("--cluster_creation_script")

@pytest.fixture
def cluster_deletion_script(request):
  return request.config.getoption("--cluster_deletion_script")

@pytest.fixture
def cluster_name(request):
  return request.config.getoption("--cluster_name")

@pytest.fixture
def build_and_apply(request):
  value = request.config.getoption("--build_and_apply").lower()

  if value in ["t", "true"]:
    return True
  else:
    return False


@pytest.fixture
def use_basic_auth(request):
  value = request.config.getoption("--use_basic_auth").lower()

  if value in ["t", "true"]:
    return True
  else:
    return False

@pytest.fixture
def use_istio(request):
  value = request.config.getoption("--use_istio").lower()

  if value in ["t", "true"]:
    return True
  else:
    return False

@pytest.fixture
def self_signed_cert(request):
  value = request.config.getoption("--self_signed_cert").lower()
  if value in ["t", "true"]:
    return True
  else:
    return False

@pytest.fixture
def upgrade_spec_path(request):
  return request.config.getoption("--upgrade_spec_path")