import kfp
from kfp import dsl
import sys
import time
import traceback

@dsl.component
def echo_op():
    print("Test pipeline")

@dsl.pipeline(name="test-pipeline", description="A test pipeline.")
def hello_world_pipeline():
    echo_task = echo_op()

def run_pipeline(token, namespace):
    print(f"[DEBUG] Starting run_pipeline for namespace: {namespace}")
    try:
        print("[DEBUG] Creating KFP client")
        kfp_client = kfp.Client(
            host="http://localhost:8080/pipeline", 
            namespace=namespace,
            cookies=""
        )
        
        print("[DEBUG] Setting up authentication headers")
        kfp_client.runs.api_client.default_headers.update(
            {"Authorization": f"Bearer {token}", "kubeflow-userid": namespace}
        )
        
        print("[DEBUG] Creating pipeline run")
        kfp_client.create_run_from_pipeline_func(
            hello_world_pipeline,
            namespace=namespace,
            arguments={},
        )
        print("[DEBUG] Pipeline run created successfully")
    except Exception as e:
        print(f"[ERROR] Failed to run pipeline: {e}")
        print(f"[ERROR] Exception type: {type(e).__name__}")
        traceback.print_exc()
        sys.exit(1)

def test_unauthorized_access(token, namespace):
    print(f"[DEBUG] Starting test_unauthorized_access for namespace: {namespace}")
    try:
        print("[DEBUG] Creating KFP client for unauthorized test")
        kfp_client = kfp.Client(
            host="http://localhost:8080/pipeline",
            namespace=namespace,
            cookies=""
        )
        
        print("[DEBUG] Setting up authentication headers")
        kfp_client.runs.api_client.default_headers.update(
            {"Authorization": f"Bearer {token}", "kubeflow-userid": namespace}
        )
        
        print("[DEBUG] Attempting to list pipelines (should fail for unauthorized users)")
        pipelines = kfp_client.list_pipelines()
        
        print(f"[WARNING] Successfully accessed {len(pipelines.pipelines or [])} pipelines - this should have failed!")
        if pipelines and len(pipelines.pipelines or []) > 0:
            print("[ERROR] Unauthorized access was allowed - security issue detected")
            sys.exit(1)
        else:
            print("[INFO] No pipelines were returned, but request should have been denied")
            sys.exit(0)
    except Exception as e:
        print(f"[DEBUG] Expected exception encountered: {e}")
        print(f"[DEBUG] Exception type: {type(e).__name__}")
        if 'ApiException' in str(type(e).__name__):
            print(f"[DEBUG] Status code: {getattr(e, 'status', 'unknown')}")
        print("[INFO] Request was properly denied as expected")
        sys.exit(0)

if __name__ == "__main__":
    print(f"[DEBUG] Starting pipeline_test.py with args: {sys.argv}")
    if len(sys.argv) < 3:
        print("[ERROR] Missing required arguments")
        print("Usage: python pipeline_test.py [run_pipeline|test_unauthorized_access] TOKEN NAMESPACE")
        sys.exit(1)
    
    action = sys.argv[1]
    token = sys.argv[2]
    namespace = sys.argv[3]
    
    print(f"[DEBUG] Running action: {action}")
    if action == "run_pipeline":
        run_pipeline(token, namespace)
    elif action == "test_unauthorized_access":
        test_unauthorized_access(token, namespace)
    else:
        print(f"[ERROR] Unknown action: {action}")
        sys.exit(1)