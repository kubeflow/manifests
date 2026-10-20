#!/usr/bin/env python3
"""Run the Spark Pi example through Kubeflow Pipelines and read the driver logs.

The pipeline runs as the default-editor service account of the profile
namespace, which receives the SparkApplication permissions through the
aggregated kubeflow-spark-edit ClusterRole and the pod log permissions through
the aggregated kubeflow-kubernetes-view ClusterRole.

Usage: spark_pipeline_test.py <token> <namespace>
"""

import sys
from pathlib import Path

import kfp
import yaml
from kfp import dsl
from kfp import kubernetes

REPOSITORY_ROOT = Path(__file__).resolve().parent.parent
SPARK_APPLICATION_YAML = (
    REPOSITORY_ROOT / "applications/spark/sparkapplication_example.yaml"
)
SPARK_APPLICATION_NAME = "spark-pi-pipeline"
KUBECTL_IMAGE = "docker.io/alpine/k8s:1.36.2"
PIPELINE_TIMEOUT_SECONDS = 900

SUBMIT_AND_WAIT_SCRIPT = """
set -eu
export HOME=/tmp
NAMESPACE="$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)"
printf '%s' "$0" | kubectl -n "$NAMESPACE" apply -f -
DEADLINE=$(( $(date +%s) + $2 ))
while true; do
  STATE="$(kubectl -n "$NAMESPACE" get sparkapplication "$1" \
    -o jsonpath='{.status.applicationState.state}')"
  echo "SparkApplication $1 state: ${STATE:-PENDING}"
  case "$STATE" in
    COMPLETED) exit 0 ;;
    FAILED|SUBMISSION_FAILED|INVALIDATING) exit 1 ;;
  esac
  if [ "$(date +%s)" -ge "$DEADLINE" ]; then
    echo "Timed out waiting for SparkApplication $1."
    exit 1
  fi
  sleep 5
done
"""

DRIVER_LOGS_SCRIPT = """
set -eu
export HOME=/tmp
NAMESPACE="$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)"
DRIVER_POD="$(kubectl -n "$NAMESPACE" get sparkapplication "$0" \
  -o jsonpath='{.status.driverInfo.podName}')"
mkdir -p "$(dirname "$1")"
kubectl -n "$NAMESPACE" logs "pod/$DRIVER_POD" | tee "$1"
grep -q "Pi is roughly" "$1"
"""


@dsl.container_component
def submit_spark_application(
    spark_application_manifest: str,
    spark_application_name: str,
    timeout_seconds: int,
):
    return dsl.ContainerSpec(
        image=KUBECTL_IMAGE,
        command=["/bin/sh", "-c"],
        args=[
            SUBMIT_AND_WAIT_SCRIPT,
            spark_application_manifest,
            spark_application_name,
            timeout_seconds,
        ],
    )


@dsl.container_component
def read_spark_driver_logs(
    spark_application_name: str,
    driver_logs: dsl.OutputPath(str),
):
    return dsl.ContainerSpec(
        image=KUBECTL_IMAGE,
        command=["/bin/sh", "-c"],
        args=[DRIVER_LOGS_SCRIPT, spark_application_name, driver_logs],
    )


@dsl.pipeline(
    name="spark-pi",
    description="Submits the Spark Pi SparkApplication and reads the driver logs.",
)
def spark_pi_pipeline(
    spark_application_manifest: str,
    spark_application_name: str = SPARK_APPLICATION_NAME,
    timeout_seconds: int = 600,
):
    submit_task = submit_spark_application(
        spark_application_manifest=spark_application_manifest,
        spark_application_name=spark_application_name,
        timeout_seconds=timeout_seconds,
    )
    driver_logs_task = read_spark_driver_logs(
        spark_application_name=spark_application_name
    ).after(submit_task)
    for task in (submit_task, driver_logs_task):
        task.set_caching_options(False)
        kubernetes.set_security_context(
            task, run_as_user=1000, run_as_group=0, run_as_non_root=True
        )


def spark_application_manifest() -> str:
    """Reuse the repository example so that both Spark tests share one source."""
    manifest = yaml.safe_load(SPARK_APPLICATION_YAML.read_text())
    manifest["metadata"]["name"] = SPARK_APPLICATION_NAME
    return yaml.safe_dump(manifest)


def main(token: str, namespace: str) -> None:
    client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)
    run = client.create_run_from_pipeline_func(
        pipeline_func=spark_pi_pipeline,
        arguments={"spark_application_manifest": spark_application_manifest()},
        experiment_name="spark-pipeline-test",
        run_name="spark-pi",
        namespace=namespace,
        enable_caching=False,
    )
    finished_run = client.wait_for_run_completion(
        run_id=run.run_id, timeout=PIPELINE_TIMEOUT_SECONDS
    )
    print(f"Pipeline run {run.run_id} finished in state {finished_run.state}.")
    if finished_run.state != "SUCCEEDED":
        sys.exit(1)


if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    main(token=sys.argv[1], namespace=sys.argv[2])
