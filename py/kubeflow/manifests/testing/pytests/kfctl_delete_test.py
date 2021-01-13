"""Run kfctl delete as a pytest.

We use this in order to generate a junit_xml file.
"""
import logging
import os
from retrying import retry
import pytest
from kubeflow.testing import util
from kubeflow.kfctl.testing.util import aws_util as kfctl_aws_util


# TODO(https://github.com/kubeflow/kfctl/issues/56): test_kfctl_delete is flaky
# and more importantly failures block upload of GCS artifacts so for now we mark
# it as expected to fail.
@pytest.mark.xfail
def test_kfctl_delete(record_xml_attribute, kfctl_path, app_path,
                      cluster_name):
  util.set_pytest_junit(record_xml_attribute, "test_kfctl_delete")

  # TODO(PatrickXYS): do we need to load kubeconfig again?

  if not kfctl_path:
    raise ValueError("kfctl_path is required")

  if not app_path:
    raise ValueError("app_path is required")

  logging.info("Using kfctl path %s", kfctl_path)
  logging.info("Using app path %s", app_path)

  kfdef_path = os.path.join(app_path, "tmp.yaml")
  logging.info("Using kfdef file path %s", kfdef_path)

  kfctl_aws_util.aws_auth_load_kubeconfig(cluster_name)

  # We see failures because delete operation will delete cert-manager and
  # knative-serving, and encounter timeout. To deal with this we do retries.
  # This has a potential downside of hiding errors that are fixed by retrying.
  @retry(stop_max_delay=60*3*1000)
  def run_delete():
    util.run([kfctl_path, "delete", "-V", "-f", kfdef_path],
             cwd=app_path)

  run_delete()


if __name__ == "__main__":
  logging.basicConfig(level=logging.INFO,
                      format=('%(levelname)s|%(asctime)s'
                              '|%(pathname)s|%(lineno)d| %(message)s'),
                      datefmt='%Y-%m-%dT%H:%M:%S',
                      )
  logging.getLogger().setLevel(logging.INFO)
  pytest.main()