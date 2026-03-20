#!/usr/bin/env python3

import kfp
import sys
import time
from kubernetes.client.models import V1SecurityContext

def hello_world_operation():
    from kfp.components import func_to_container_op
    
    def hello_world():
        print("Hello World from Kubeflow Pipelines V1!")
        return "Hello World"
    
    return func_to_container_op(hello_world)

def hello_world_pipeline():
    hello_world_task = hello_world_operation()
    hello_world_task()


def apply_security_context(operation):
    operation.container.set_security_context(
        V1SecurityContext(
            run_as_user=1000,
            run_as_group=0,
            run_as_non_root=True,
        )
    )
    return operation

def run_v1_pipeline(token, namespace):
    client = kfp.Client(host="http://localhost:8080/pipeline", existing_token=token)
    
    experiment = client.create_experiment("v1-pipeline-test", namespace=namespace)
    
    pipeline_configuration = kfp.dsl.PipelineConf().add_op_transformer(
        apply_security_context
    )

    pipeline_run = client.create_run_from_pipeline_func(
        hello_world_pipeline,
        experiment_name=experiment.name,
        run_name="v1-hello-world",
        namespace=namespace,
        arguments={},
        pipeline_conf=pipeline_configuration,
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
