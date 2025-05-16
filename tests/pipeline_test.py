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
    client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)
    
    experiment = client.create_experiment("default", namespace=namespace)
    
    client.create_run_from_pipeline_func(
        hello_world_pipeline,
        experiment_name=experiment.name,
        namespace=namespace,
        arguments={},
    )

def test_unauthorized_access(token, namespace):
    try:
        client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)
        
        client.list_runs(namespace=namespace)
        sys.exit(1)
    except Exception:
        sys.exit(0)

if __name__ == "__main__":
    if len(sys.argv) < 3:
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