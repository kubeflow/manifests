from time import sleep, time
import kfp
import sys
from kfp_server_api.exceptions import ApiException


def run_pipeline(token, namespace):
    client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)

    pipelines = client.list_pipelines().pipelines
    if not pipelines:
        sys.exit(1)
        
    pipeline = pipelines[0]
    pipeline_name = pipeline.display_name
    pipeline_id = pipeline.pipeline_id

    pipeline_versions = client.list_pipeline_versions(pipeline_id).pipeline_versions
    if not pipeline_versions:
        print(f"No versions found for pipeline {pipeline_name}")
        sys.exit(1)
        
    pipeline_version_id = pipeline_versions[0].pipeline_version_id
    
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

    timeout = time() + 300
    
    while time() < timeout:
        status = client.get_run(run_id=run_id).state
        if status in ["PENDING", "RUNNING"]:
            print(f"Waiting for run_id: {run_id}, status: {status}.")
            sleep(10)
        else:
            print(f"Run with id {run_id} finished with status: {status}.")
            if status != "SUCCEEDED":
                print("Pipeline failed")
                sys.exit(1)
            return
            
    print(f"Pipeline run timed out after 5 minutes")
    sys.exit(1)


def test_unauthorized_access(token, namespace):
    client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)

    try:
        pipeline = client.list_runs(namespace=namespace)
        sys.exit(1)
    except ApiException as e:
        if e.status != 403:
            sys.exit(1)


if __name__ == "__main__":
    if len(sys.argv) != 4:
        sys.exit(1)
        
    action = sys.argv[1]
    token = sys.argv[2]
    namespace = sys.argv[3]

    if action == "run_pipeline":
        run_pipeline(token, namespace)
    elif action == "test_unauthorized_access":
        test_unauthorized_access(token, namespace)
    else:
        sys.exit(1)
