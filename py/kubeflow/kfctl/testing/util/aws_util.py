import os
import logging
from kubeflow.testing import util
from kubeflow.testing.cloudprovider.aws import util as aws_util


def aws_auth_load_kubeconfig(cluster_name):
    logging.info("updating ~/.kube/config file of the EKS cluster")
    util.run(["aws", "eks", "update-kubeconfig", "--name=" + cluster_name])

    aws_util.load_kube_config()