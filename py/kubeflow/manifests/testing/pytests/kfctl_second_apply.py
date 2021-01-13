import logging
import os

import pytest

from kubeflow.kfctl.testing.util import kfctl_go_test_utils as kfctl_util
from kubeflow.testing import util


def test_second_apply(record_xml_attribute, app_path, kfctl_path):
  """Test that we can run kfctl apply again with error.

  Args:
    app_path: The app dir of kubeflow deployment.
    kfctl_path: The path to kfctl binary.
  """
  if not os.path.exists(kfctl_path):
    msg = "kfctl Go binary not found: {path}".format(path=kfctl_path)
    logging.error(msg)
    raise RuntimeError(msg)

  # Need to activate account for scopes.
  if os.getenv("GOOGLE_APPLICATION_CREDENTIALS"):
    util.run(["gcloud", "auth", "activate-service-account",
              "--key-file=" + os.environ["GOOGLE_APPLICATION_CREDENTIALS"]])

  util.run([kfctl_path, "apply", "-V", "-f=" + os.path.join(app_path, "tmp.yaml")], cwd=app_path)
