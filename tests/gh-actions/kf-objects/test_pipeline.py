import kfp
from kfp import dsl
import kfp.components as comp


@comp.create_component_from_func
def echo_op():
    print("Test pipeline")

@dsl.pipeline(
    name='test-pipeline',
    description='A test pipeline.'
)
def hello_world_pipeline():
    echo_task = echo_op()

if __name__ == "__main__":
    # Run the Kubeflow Pipeline in the user's namespace.
    kfp_client = kfp.Client(host="http://localhost:3000",
                            namespace="kubeflow-user-example-com")
    kfp_client.runs.api_client.default_headers.update(
        {"kubeflow-userid": "kubeflow-user-example-com"})
    # create the KFP run
    run_id = kfp_client.create_run_from_pipeline_func(
        hello_world_pipeline,
        namespace="kubeflow-user-example-com",
        arguments={},
    ).run_id