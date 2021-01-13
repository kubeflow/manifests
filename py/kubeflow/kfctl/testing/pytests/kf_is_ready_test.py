import logging
import os
import yaml

import pytest

from kubeflow.testing import util
from kubeflow.kfctl.testing.util import deploy_utils
from kubeflow.kfctl.testing.util import aws_util as kfctl_aws_util


def set_logging():
    logging.basicConfig(level=logging.INFO,
                        format=('%(levelname)s|%(asctime)s'
                                '|%(pathname)s|%(lineno)d| %(message)s'),
                        datefmt='%Y-%m-%dT%H:%M:%S',
                        )
    logging.getLogger().setLevel(logging.INFO)


def get_platform_app_name(app_path):
    with open(os.path.join(app_path, "tmp.yaml")) as f:
        kfdef = yaml.safe_load(f)
    app_name = kfdef["metadata"]["name"]
    platform = ""
    apiVersion = kfdef["apiVersion"].strip().split("/")
    if len(apiVersion) != 2:
        raise RuntimeError("Invalid apiVersion: " + kfdef["apiVersion"].strip())
    if apiVersion[-1] == "v1alpha1":
        platform = kfdef["spec"]["platform"]
    elif apiVersion[-1] in ["v1beta1", "v1"]:
        for plugin in kfdef["spec"].get("plugins", []):
            if plugin.get("kind", "") == "KfGcpPlugin":
                platform = "gcp"
            elif plugin.get("kind", "") == "KfAwsPlugin":
                platform = "aws"
            elif plugin.get("kind", "") == "KfExistingArriktoPlugin":
                platform = "existing_arrikto"
            else:
                # Indicate agnostic Kubeflow Platform
                platform = "agnostic"
    else:
        raise RuntimeError("Unknown version: " + apiVersion[-1])
    return platform, app_name


def check_deployments_ready(record_xml_attribute, namespace, name, deployments, cluster_name):
    """Test that Kubeflow deployments are successfully deployed.

  Args:
    namespace: The namespace Kubeflow is deployed to.
  """
    set_logging()
    util.set_pytest_junit(record_xml_attribute, name)

    kfctl_aws_util.aws_auth_load_kubeconfig(cluster_name)

    api_client = deploy_utils.create_k8s_client()

    for deployment_name in deployments:
        logging.info("Verifying that deployment %s started...", deployment_name)
        util.wait_for_deployment(api_client, namespace, deployment_name, 10)


def test_admission_is_ready(record_xml_attribute, namespace, cluster_name):
    deployment_names = [
        "admission-webhook-deployment"
    ]
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_admission_is_ready", deployment_names,
                            cluster_name)


def test_katib_is_ready(record_xml_attribute, namespace, cluster_name):
    deployment_names = [
        "katib-controller",
        "katib-mysql",
        "katib-db-manager",
        "katib-ui",
    ]
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_katib_is_ready", deployment_names,
                            cluster_name)


def test_metadata_is_ready(record_xml_attribute, namespace, cluster_name):
    deployment_names = [
        "metadata-grpc-deployment",
        "metadata-db",
        "metadata-envoy-deployment",
        "metadata-writer",
    ]
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_metadata_is_ready", deployment_names,
                            cluster_name)


def test_pipeline_is_ready(record_xml_attribute, namespace, cluster_name):
    deployment_names = [
        "argo-ui",
        "cache-deployer-deployment",
        "cache-server",
        "kubeflow-pipelines-profile-controller",
        "minio",
        "ml-pipeline",
        "ml-pipeline-persistenceagent",
        "ml-pipeline-scheduledworkflow",
        "ml-pipeline-ui",
        "ml-pipeline-viewer-crd",
        "ml-pipeline-visualizationserver",
        "mysql",
    ]
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_pipeline_is_ready", deployment_names,
                            cluster_name)


def test_notebook_is_ready(record_xml_attribute, namespace, cluster_name):
    deployment_names = [
        "jupyter-web-app-deployment",
        "notebook-controller-deployment",
    ]
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_notebook_is_ready", deployment_names,
                            cluster_name)


def test_centraldashboard_is_ready(record_xml_attribute, namespace, cluster_name):
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_centraldashboard_is_ready", ["centraldashboard"],
                            cluster_name)


def test_profiles_is_ready(record_xml_attribute, namespace, cluster_name):
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_profile_is_ready", ["profiles-deployment"],
                            cluster_name)


def test_seldon_is_ready(record_xml_attribute, namespace, cluster_name):
    deployment_names = [
        "seldon-controller-manager"
    ]
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_seldon_is_ready", deployment_names,
                            cluster_name)


def test_spark_is_ready(record_xml_attribute, namespace, cluster_name):
    deployment_names = [
        "spark-operatorsparkoperator"
    ]
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_spark_is_ready", deployment_names,
                            cluster_name)


def test_training_operators_are_ready(record_xml_attribute, namespace, cluster_name):
    deployment_names = [
        "mpi-operator",
        "mxnet-operator",
        "pytorch-operator",
        "tf-job-operator",
    ]

    check_deployments_ready(record_xml_attribute, namespace,
                            "test_training_is_ready", deployment_names,
                            cluster_name)


def test_workflow_controller_is_ready(record_xml_attribute, namespace, cluster_name):
    check_deployments_ready(record_xml_attribute, namespace,
                            "test_workflow_controller_is_ready", ["workflow-controller"],
                            cluster_name)


def test_kf_is_ready(record_xml_attribute, namespace, use_basic_auth,
                     app_path, cluster_name):
    """Test that Kubeflow was successfully deployed.

  Args:
    namespace: The namespace Kubeflow is deployed to.
  """
    set_logging()
    util.set_pytest_junit(record_xml_attribute, "test_kf_is_ready")

    kfctl_aws_util.aws_auth_load_kubeconfig(cluster_name)

    api_client = deploy_utils.create_k8s_client()

    # Verify that components are actually deployed.
    deployment_names = []

    stateful_set_names = []

    daemon_set_names = []

    platform, _ = get_platform_app_name(app_path)

    # TODO(PatrickXYS): not sure why istio-galley can't found
    ingress_related_deployments = [
        "cluster-local-gateway",
        "istio-citadel",
        "istio-ingressgateway",
        "istio-pilot",
        "istio-policy",
        "istio-sidecar-injector",
        "istio-telemetry",
        "prometheus",
    ]
    ingress_related_stateful_sets = []

    knative_namespace = "knative-serving"
    knative_related_deployments = [
        "activator",
        "autoscaler",
        "controller",
        "networking-istio",
        "webhook",
    ]

    if platform == "gcp":
        deployment_names.extend(["cloud-endpoints-controller"])
        stateful_set_names.extend(["kfserving-controller-manager"])
        if use_basic_auth:
            deployment_names.extend(["basic-auth-login"])
            ingress_related_stateful_sets.extend(["backend-updater"])
        else:
            ingress_related_deployments.extend(["iap-enabler"])
            ingress_related_stateful_sets.extend(["backend-updater"])
    elif platform == "existing_arrikto":
        deployment_names.extend(["dex"])
        ingress_related_deployments.extend(["authservice"])
        knative_related_deployments = []
    elif platform == "aws":
        # TODO(PatrickXYS): Extend List with AWS Deployment
        deployment_names.extend(["alb-ingress-controller"])
        daemon_set_names.extend(["nvidia-device-plugin-daemonset"])

    # TODO(jlewi): Might want to parallelize this.
    for deployment_name in deployment_names:
        logging.info("Verifying that deployment %s started...", deployment_name)
        util.wait_for_deployment(api_client, namespace, deployment_name, 10)

    ingress_namespace = "istio-system"
    for deployment_name in ingress_related_deployments:
        logging.info("Verifying that deployment %s started...", deployment_name)
        util.wait_for_deployment(api_client, ingress_namespace, deployment_name, 10)

    all_stateful_sets = [(namespace, name) for name in stateful_set_names]
    all_stateful_sets.extend([(ingress_namespace, name) for name in ingress_related_stateful_sets])

    for ss_namespace, name in all_stateful_sets:
        logging.info("Verifying that stateful set %s.%s started...", ss_namespace, name)
        try:
            util.wait_for_statefulset(api_client, ss_namespace, name)
        except:
            # Collect debug information by running describe
            util.run(["kubectl", "-n", ss_namespace, "describe", "statefulsets", name])
            raise

    all_daemon_sets = [(namespace, name) for name in daemon_set_names]

    for ds_namespace, name in all_daemon_sets:
        logging.info("Verifying that daemonset set %s.%s started...", ds_namespace, name)
        try:
            util.wait_for_daemonset(api_client, ds_namespace, name)
        except:
            # Collect debug information by running describe
            util.run(["kubectl", "-n", ds_namespace, "describe", "daemonset", name])
            raise

    ingress_names = ["istio-ingress"]
    # Check if Ingress is Ready and Healthy
    if platform in ["aws"]:
        for ingress_name in ingress_names:
            logging.info("Verifying that ingress %s started...", ingress_name)
            util.wait_for_ingress(api_client, ingress_namespace, ingress_name, 10)

    for deployment_name in knative_related_deployments:
        logging.info("Verifying that deployment %s started...", deployment_name)
        util.wait_for_deployment(api_client, knative_namespace, deployment_name, 10)

    # Check if Dex is Ready and Healthy
    dex_deployment_names = ["dex"]
    dex_namespace = "auth"
    for dex_deployment_name in dex_deployment_names:
        logging.info("Verifying that deployment %s started...", dex_deployment_name)
        util.wait_for_deployment(api_client, dex_namespace, dex_deployment_name, 10)

    # Check if Cert-Manager is Ready and Healthy
    cert_manager_deployment_names = [
        "cert-manager",
        "cert-manager-cainjector",
        "cert-manager-webhook",
    ]
    cert_manager_namespace = "cert-manager"
    for cert_manager_deployment_name in cert_manager_deployment_names:
        logging.info("Verifying that deployment %s started...", cert_manager_deployment_name)
        util.wait_for_deployment(api_client, cert_manager_namespace, cert_manager_deployment_name, 10)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO,
                        format=('%(levelname)s|%(asctime)s'
                                '|%(pathname)s|%(lineno)d| %(message)s'),
                        datefmt='%Y-%m-%dT%H:%M:%S',
                        )
    logging.getLogger().setLevel(logging.INFO)
    pytest.main()
