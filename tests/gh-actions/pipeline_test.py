from time import sleep
import kfp
import sys
from kfp_server_api.exceptions import ApiException


def run_pipeline(token, namespace):
    client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)

    pipeline = client.list_pipelines().pipelines[0]
    pipeline_name = pipeline.display_name
    pipeline_id = pipeline.pipeline_id
    pipeline_version_id = (
        client.list_pipeline_versions(pipeline_id)
        .pipeline_versions[0]
        .pipeline_version_id
    )
    experiment_id = client.create_experiment(
        "m2m-test", namespace=namespace
    ).experiment_id

    print(f"Starting pipeline {pipeline_name}.")
    run_id = client.run_pipeline(
        experiment_id=experiment_id,
        job_name="m2m-test",
        pipeline_id=pipeline_id,
        version_id=pipeline_version_id,
    ).run_id

    while True:
        status = client.get_run(run_id=run_id).state
        if status in ["PENDING", "RUNNING"]:
            print(f"Waiting for run_id: {run_id}, status: {status}.")
            sleep(10)
        else:
            print(f"Run with id {run_id} finished with status: {status}.")
            if status != "SUCCEEDED":
                print("Pipeline failed")
                raise SystemExit(1)
            break


def test_unauthorized_access(token, namespace):
    client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)

    try:
        pipeline = client.list_runs(namespace=namespace)
    except ApiException as e:
        assert (
            e.status == 403
        ), "This API Call should return unauthorized/forbidden error."


if __name__ == "__main__":
    action = sys.argv[1]
    token = sys.argv[2]
    namespace = sys.argv[3]

    if action == "run_pipeline":
        run_pipeline(token, namespace)
    elif action == "test_unauthorized_access":
        test_unauthorized_access(token, namespace)
