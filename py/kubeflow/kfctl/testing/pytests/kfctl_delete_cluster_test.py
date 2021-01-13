"""Run kfctl delete as a pytest.

We use this in order to generate a junit_xml file.
"""
import logging
import pytest
import os
from kubeflow.testing import util


def test_kfctl_delete(record_xml_attribute, cluster_deletion_script,
                      cluster_name):
    util.set_pytest_junit(record_xml_attribute, "test_cluster_delete")

    if cluster_deletion_script:
        logging.info("cluster_deletion_script specified: %s", cluster_deletion_script)
        os.environ["CLUSTER_NAME"] = cluster_name
        util.run(["/bin/bash", "-c", cluster_deletion_script])
        return


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO,
                        format=('%(levelname)s|%(asctime)s'
                                '|%(pathname)s|%(lineno)d| %(message)s'),
                        datefmt='%Y-%m-%dT%H:%M:%S',
                        )
    logging.getLogger().setLevel(logging.INFO)
    pytest.main()