import kfp
from kfp import dsl
import sys


@dsl.component
def echo_op():
    print("Test pipeline")


@dsl.pipeline(name="test-pipeline", description="A test pipeline.")
def hello_world_pipeline():
    echo_task = echo_op()


def run_pipeline(token, namespace):
    kfp_client = kfp.Client(
        host="http://localhost:8080/pipeline", 
        namespace=namespace,
        cookies=""
    )
    kfp_client.runs.api_client.default_headers.update(
        {"Authorization": f"Bearer {token}", "kubeflow-userid": namespace}
    )
    kfp_client.create_run_from_pipeline_func(
        hello_world_pipeline,
        namespace=namespace,
        arguments={},
    )


def test_unauthorized_access(token, namespace):
    try:
        kfp_client = kfp.Client(
            host="http://localhost:8080/pipeline", 
            namespace=namespace,
            cookies=""
        )
        kfp_client.runs.api_client.default_headers.update(
            {"Authorization": f"Bearer {token}", "kubeflow-userid": namespace}
        )
        
        kfp_client.list_pipelines()
        return False
    except Exception:
        return True


if __name__ == "__main__":
    if len(sys.argv) < 3:
        sys.exit(1)
    
    action = sys.argv[1]
    token = sys.argv[2]
    namespace = sys.argv[3]
    
    if action == "run_pipeline":
        run_pipeline(token, namespace)
    elif action == "test_unauthorized_access":
        success = test_unauthorized_access(token, namespace)
        if not success:
            sys.exit(1)
    else:
        sys.exit(1)
