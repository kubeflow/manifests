#!/usr/bin/env python3

from kfp import dsl
import kfp
from time import sleep
import subprocess


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
    second_add_task = add(a=first_add_task.output, b=b)


try:
    print("getting experiment...")
    experiment = client.get_experiment(
        experiment_name=experiment_name, namespace=experiment_namespace
    )
    print("got experiment!")
except Exception:
    print("creating experiment...")
    experiment = client.create_experiment(
        name=experiment_name, namespace=experiment_namespace
    )
    print("created experiment!")

run = client.create_run_from_pipeline_func(
    add_pipeline,
    arguments={"a": 7.0, "b": 8.0},
    experiment_id=experiment.experiment_id,
    enable_caching=False,
)

while True:
    live_run = client.get_run(run_id=run.run_id)
    print(f"{live_run.state=}")

    subprocess.run(["kubectl", "get", "pods"])

    if live_run.finished_at > live_run.created_at:
        print("Finished pipeline!")
        print(f"{live_run.finished_at > live_run.created_at=}")
        print(f"{live_run.state=}")
        print(f"{live_run.created_at=}")
        print(f"{live_run.finished_at=}")
        print(f"{live_run.error=}")
        break
    else:
        print("Waiting for pipeline to finish...")
    sleep(5)
