import logging

import pytest

from kubeflow.testing import util
import json
from retrying import retry
from time import sleep
import uuid
from kubeflow.kfctl.testing.util import aws_util as kfctl_aws_util


def test_kfam(record_xml_attribute, cluster_name):
  util.set_pytest_junit(record_xml_attribute, "test_kfam_e2e")
  kfctl_aws_util.aws_auth_load_kubeconfig(cluster_name)

  getcmd = "kubectl get pods -n kubeflow -l=app=jupyter-web-app --template '{{range.items}}{{.metadata.name}}{{end}}'"
  jupyterpod = util.run(getcmd.split(' '))[1:-1]

  logging.info("accessing kfam svc from jupyter pod %s" % jupyterpod)

  sleep(10)
  # Profile Creation
  profile_name = "testprofile-%s" % uuid.uuid4().hex[0:7]
  util.run(['kubectl', 'exec', jupyterpod, '-n', 'kubeflow', '--', 'curl',
            '--silent', '-X', 'POST', '-d',
            '{"metadata":{"name":"%s"},"spec":{"owner":{"kind":"User","name":"user1@kubeflow.org"}}}' % profile_name,
            'profiles-kfam.kubeflow:8081/kfam/v1/profiles'])

  assert verify_profile_creation(jupyterpod, profile_name)


@retry(wait_fixed=2000, stop_max_delay=20 * 1000)
def verify_profile_creation(jupyterpod, profile_name):
  # Verify Profile Creation
  bindingsstr = util.run(['kubectl', 'exec', jupyterpod, '-n', 'kubeflow', '--', 'curl', '--silent',
                          'profiles-kfam.kubeflow:8081/kfam/v1/bindings'])
  bindings = json.loads(bindingsstr)

  if profile_name not in [binding['referredNamespace'] for binding in bindings['bindings']]:
    raise Exception("testprofile not created yet!")
  return True


if __name__ == "__main__":
  logging.basicConfig(level=logging.INFO,
                      format=('%(levelname)s|%(asctime)s'
                              '|%(pathname)s|%(lineno)d| %(message)s'),
                      datefmt='%Y-%m-%dT%H:%M:%S',
                      )
  logging.getLogger().setLevel(logging.INFO)
  pytest.main()