#!/usr/bin/env python3

import kfp
import sys
import time

PIPELINE_TASK_SECURITY_RUN_AS_USER = 1000
PIPELINE_TASK_SECURITY_RUN_AS_GROUP = 1000
PIPELINE_TASK_SECURITY_RUN_AS_NON_ROOT = True
PIPELINE_TASK_SECURITY_CONTEXT = (
    f"runAsUser: {PIPELINE_TASK_SECURITY_RUN_AS_USER}\n"
    f"runAsGroup: {PIPELINE_TASK_SECURITY_RUN_AS_GROUP}\n"
    f"runAsNonRoot: {str(PIPELINE_TASK_SECURITY_RUN_AS_NON_ROOT).lower()}"
)


def hello_world_op():
    from kfp.components import func_to_container_op
    
    def hello_world():
        print("Hello World from Kubeflow Pipelines V1!")
        return "Hello World"
    
    return func_to_container_op(hello_world)

def hello_world_pipeline():
    hello_world_operation = hello_world_op()
    hello_task = hello_world_operation()
    hello_task.container.set_security_context(PIPELINE_TASK_SECURITY_CONTEXT)

def run_v1_pipeline(token, namespace):
    client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)
    
    experiment = client.create_experiment("v1-pipeline-test", namespace=namespace)
    
    pipeline_run = client.create_run_from_pipeline_func(
        hello_world_pipeline,
        experiment_name=experiment.name,
        run_name="v1-hello-world",
        namespace=namespace,
        arguments={}
    )
    
    for iteration in range(15):
        pipeline_status = client.get_run(pipeline_run.run_id).run.status
        
        if pipeline_status == "Succeeded":
            return
        elif pipeline_status not in ["Running", "Pending"]:
            sys.exit(1)
            
        time.sleep(10)
    
    sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit(1)
        
    run_v1_pipeline(sys.argv[1], sys.argv[2]) 
