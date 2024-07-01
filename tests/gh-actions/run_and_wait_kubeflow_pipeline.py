#!/usr/bin/env python3

from kfp import dsl
import kfp
from time import sleep
import subprocess
import logging
import sys
from datetime import datetime, timezone

logger = logging.getLogger("run_and_wait_for_pipeline")
logging.basicConfig(
    stream=sys.stdout,
    level=logging.DEBUG,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)


client = kfp.Client()
experiment_name = "my-experiment"
experiment_namespace = "kubeflow-user-example-com"


@dsl.component
def add(a: float, b: float) -> float:
    """Calculates sum of two arguments"""
    return a + b


@dsl.pipeline(
    name="Addition pipeline",
    description="An example pipeline that performs addition calculations.",
)
def add_pipeline(
    a: float = 1.0,
    b: float = 7.0,
):
    first_add_task = add(a=a, b=4.0)
    add(a=first_add_task.output, b=b)


try:
    logger.info(
        f"Trying to get experiment from {experiment_name=} {experiment_namespace=}."
    )
    experiment = client.get_experiment(
        experiment_name=experiment_name, namespace=experiment_namespace
    )
    logger.info("Experiment found!")
except Exception:
    logger.info("Experiment not found, trying to create experiment.")
    experiment = client.create_experiment(
        name=experiment_name, namespace=experiment_namespace
    )
    logger.info("Experiment created!")

try:
    logger.info("Trying to create Pipeline Run.")
    run = client.create_run_from_pipeline_func(
        add_pipeline,
        arguments={"a": 7.0, "b": 8.0},
        experiment_id=experiment.experiment_id,
        enable_caching=False,
    )
except Exception as e:
    logger.error(
        f"Failed to create Pipeline Run. Exception: {e.__class__.__name__}: {str(e)}"
    )
    raise SystemExit(1)

# For now being able to start a pipeline is enough.
# while True:
#     live_run = client.get_run(run_id=run.run_id)
#     logger.info(f"Pipeline Run State: {live_run.state}.")

#     minutes_from_pipeline_run_start = (
#         datetime.now(timezone.utc) - live_run.created_at
#     ).seconds / 60

#     if minutes_from_pipeline_run_start > 5:
#         logger.debug(
#             "Pipeline is running for more than 5 minutes, "
#             f"showing pod states in {experiment_namespace=}."
#         )
#         subprocess.run(["kubectl", "get", "pods"])

#     if live_run.finished_at > live_run.created_at:
#         logger.info("Finished Pipeline Run!")
#         logger.info(
#             f"Pipeline was running for {minutes_from_pipeline_run_start:0.2} minutes."
#         )
#         logger.info(f"Pipeline Run finished in state: {live_run.state}.")
#         logger.info(f"Pipeline Run finished with error: {live_run.error}.")

#         if live_run.state != "SUCCEEDED":
#             logger.warn("The Pipeline Run finished but has failed...")

#             logger.warn("Running 'kubectl get pods':")
#             subprocess.run(["kubectl", "get", "pods"])

#             logger.warn("Running 'kubectl describe wf':")
#             subprocess.run(["kubectl", "describe", "wf"])

#             raise SystemExit(1)
#         break
#     else:
#         logger.info("Waiting for pipeline to finish...")
#     sleep(5)
