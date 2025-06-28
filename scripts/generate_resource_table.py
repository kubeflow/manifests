#!/usr/bin/env python3

import subprocess
import re
import sys
import os
import glob
from collections import defaultdict

try:
    import yaml
    YAML_AVAILABLE = True
except ImportError:
    print("Warning: PyYAML not available. Install with: pip install PyYAML")
    print("Falling back to actual usage only...")
    YAML_AVAILABLE = False

def run_kubectl_top():
    """Run kubectl top pods commands and return the output"""
    try:
        # Get pod resource usage
        pod_result = subprocess.run(['kubectl', 'top', 'pods', '--all-namespaces'], 
                                  capture_output=True, text=True, check=True)
        return pod_result.stdout
    except subprocess.CalledProcessError as e:
        print(f"Error running kubectl top: {e}")
        print(f"Make sure metrics-server is installed and running in your cluster")
        print(f"Error output: {e.stderr if hasattr(e, 'stderr') else 'N/A'}")
        sys.exit(1)
    except FileNotFoundError:
        print("Error: kubectl command not found. Please install kubectl and ensure it's in your PATH")
        sys.exit(1)

def parse_cpu_to_millicores(cpu_str):
    """Convert CPU string to millicores (m)"""
    if cpu_str.endswith('m'):
        return int(cpu_str[:-1])
    elif cpu_str.endswith('n'):
        return int(cpu_str[:-1]) // 1000000  # nanocores to millicores
    else:
        return int(float(cpu_str) * 1000)  # cores to millicores

def parse_memory_to_mib(memory_str):
    """Convert memory string to MiB"""
    if memory_str.endswith('Mi'):
        return int(memory_str[:-2])
    elif memory_str.endswith('Gi'):
        return int(float(memory_str[:-2]) * 1024)
    elif memory_str.endswith('Ki'):
        return int(memory_str[:-2]) // 1024
    elif memory_str.endswith('Bi') or memory_str.endswith('B'):
        return int(memory_str[:-2]) // (1024 * 1024)
    else:
        # Assume bytes
        return int(memory_str) // (1024 * 1024)

def categorize_pod(namespace, pod_name):
    """Categorize pods into components based on namespace and pod name"""
    
    # Skip system pods
    if namespace in ['kube-system', 'local-path-storage']:
        return None
        
    # Dex + OAuth2-Proxy
    if namespace in ['auth', 'oauth2-proxy'] or 'dex' in pod_name or 'oauth2-proxy' in pod_name:
        return 'Dex + OAuth2-Proxy'
    
    # Cert Manager
    if namespace == 'cert-manager' or 'cert-manager' in pod_name:
        return 'Cert Manager'
    
    # Istio
    if namespace == 'istio-system' or any(x in pod_name for x in ['istio', 'cluster-local-gateway', 'cluster-jwks-proxy']):
        return 'Istio'
    
    # Knative (part of Istio/infrastructure)
    if namespace == 'knative-serving' or 'knative' in pod_name:
        return 'Istio'
    
    # Katib
    if 'katib' in pod_name:
        return 'Katib'
    
    # KServe  
    if 'kserve' in pod_name or 'predictor' in pod_name:
        return 'KServe'
    
    # Metadata
    if any(x in pod_name for x in ['metadata', 'mysql']) and 'model-registry' not in pod_name:
        return 'Metadata'
    
    # Model Registry
    if 'model-registry' in pod_name:
        return 'Model Registry'
    
    # Pipelines
    if any(x in pod_name for x in ['ml-pipeline', 'workflow-controller', 'cache-server', 'minio']):
        return 'Pipelines'
    
    # Spark
    if 'spark' in pod_name:
        return 'Spark'
    
    # Training
    if 'training' in pod_name:
        return 'Training'
    
    # Kubeflow Core (central dashboard, jupyter, profiles, etc.)
    if any(x in pod_name for x in ['centraldashboard', 'jupyter', 'notebook-controller', 
                                   'profiles', 'admission-webhook', 'volumes-web-app', 
                                   'tensorboard', 'pvcviewer']):
        return 'Kubeflow Core'
    
    # Everything else in kubeflow namespace or kubeflow-user namespaces
    if namespace.startswith('kubeflow'):
        return 'Other'
    
    return None

def parse_kubectl_output(output):
    """Parse kubectl top output and categorize by component"""
    lines = output.strip().split('\n')[1:]  # Skip header
    
    component_resources = defaultdict(lambda: {'cpu': 0, 'memory': 0})
    
    for line in lines:
        parts = line.split()
        if len(parts) < 4:
            continue
            
        namespace = parts[0]
        pod_name = parts[1]
        cpu_str = parts[2]
        memory_str = parts[3]
        
        component = categorize_pod(namespace, pod_name)
        if component is None:
            continue
            
        try:
            cpu_millicores = parse_cpu_to_millicores(cpu_str)
            memory_mib = parse_memory_to_mib(memory_str)
            
            component_resources[component]['cpu'] += cpu_millicores
            component_resources[component]['memory'] += memory_mib
        except (ValueError, IndexError) as e:
            print(f"Warning: Could not parse resource values for {pod_name}: {e}")
            continue
    
    return component_resources

def find_manifest_files():
    """Find all YAML manifest files in applications/, common/, and experimental/ directories"""
    manifest_files = []
    base_dirs = ['applications', 'common', 'experimental']
    
    for base_dir in base_dirs:
        if os.path.exists(base_dir):
            # Find all .yaml and .yml files recursively
            yaml_files = glob.glob(f"{base_dir}/**/*.yaml", recursive=True)
            yml_files = glob.glob(f"{base_dir}/**/*.yml", recursive=True)
            manifest_files.extend(yaml_files)
            manifest_files.extend(yml_files)
    
    return manifest_files

def parse_manifest_file(filepath):
    """Parse a YAML manifest file and extract resource requests"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            # Handle multiple YAML documents in one file
            documents = list(yaml.safe_load_all(f))
            
        resources = []
        for doc in documents:
            if not doc or not isinstance(doc, dict):
                continue
                
            # Extract resources from different Kubernetes resource types
            extracted = extract_resources_from_doc(doc, filepath)
            resources.extend(extracted)
                
        return resources
    except Exception as e:
        # Skip files that can't be parsed (templates, non-k8s files, etc.)
        return []

def extract_resources_from_doc(doc, filepath):
    """Extract resource requests from a Kubernetes document"""
    resources = []
    
    if not doc.get('kind') or not doc.get('metadata'):
        return resources
        
    kind = doc.get('kind', '')
    metadata = doc.get('metadata', {})
    name = metadata.get('name', 'unknown')
    namespace = metadata.get('namespace', 'default')
    
    # Get the component category based on file path and resource name
    component = categorize_manifest_resource(filepath, namespace, name, kind)
    if not component:
        return resources
    
    # Extract container resources from various resource types
    containers = []
    
    if kind in ['Deployment', 'StatefulSet', 'DaemonSet']:
        spec = doc.get('spec', {})
        template = spec.get('template', {})
        pod_spec = template.get('spec', {})
        containers = pod_spec.get('containers', [])
        
    elif kind == 'Pod':
        spec = doc.get('spec', {})
        containers = spec.get('containers', [])
        
    elif kind == 'Job':
        spec = doc.get('spec', {})
        template = spec.get('template', {})
        pod_spec = template.get('spec', {})
        containers = pod_spec.get('containers', [])
        
    elif kind == 'CronJob':
        spec = doc.get('spec', {})
        job_template = spec.get('jobTemplate', {})
        job_spec = job_template.get('spec', {})
        template = job_spec.get('template', {})
        pod_spec = template.get('spec', {})
        containers = pod_spec.get('containers', [])
    
    # Parse resource requests from containers
    for container in containers:
        container_resources = container.get('resources', {})
        requests = container_resources.get('requests', {})
        
        cpu_request = requests.get('cpu', '0')
        memory_request = requests.get('memory', '0')
        
        if cpu_request != '0' or memory_request != '0':
            try:
                cpu_millicores = parse_cpu_to_millicores(cpu_request)
                memory_mib = parse_memory_to_mib(memory_request)
                
                resources.append({
                    'component': component,
                    'cpu': cpu_millicores,
                    'memory': memory_mib,
                    'source': f"{filepath}:{name}:{container.get('name', 'unknown')}"
                })
            except (ValueError, TypeError):
                # Skip invalid resource values
                continue
    
    return resources

def categorize_manifest_resource(filepath, namespace, name, kind):
    """Categorize manifest resources into components based on file path and resource details"""
    
    # Normalize filepath for comparison
    filepath = filepath.lower().replace('\\', '/')
    name = name.lower()
    
    # Dex + OAuth2-Proxy
    if any(x in filepath for x in ['dex', 'oauth2-proxy']) or any(x in name for x in ['dex', 'oauth2-proxy']):
        return 'Dex + OAuth2-Proxy'
    
    # Cert Manager
    if 'cert-manager' in filepath or 'cert-manager' in name:
        return 'Cert Manager'
    
    # Istio (includes Knative)
    if any(x in filepath for x in ['istio', 'knative']) or any(x in name for x in ['istio', 'knative', 'cluster-local-gateway', 'cluster-jwks-proxy']):
        return 'Istio'
    
    # Katib
    if 'katib' in filepath or 'katib' in name:
        return 'Katib'
    
    # KServe
    if 'kserve' in filepath or any(x in name for x in ['kserve', 'predictor']):
        return 'KServe'
    
    # Model Registry
    if 'model-registry' in filepath or 'model-registry' in name:
        return 'Model Registry'
    
    # Pipelines
    if 'pipeline' in filepath or any(x in name for x in ['ml-pipeline', 'workflow-controller', 'cache-server', 'minio']):
        return 'Pipelines'
    
    # Metadata (but not model-registry)
    if any(x in name for x in ['metadata', 'mysql']) and 'model-registry' not in name:
        return 'Metadata'
    
    # Spark
    if 'spark' in filepath or 'spark' in name:
        return 'Spark'
    
    # Training
    if 'training-operator' in filepath or 'training' in name:
        return 'Training'
    
    # Kubeflow Core
    if any(x in filepath for x in ['centraldashboard', 'jupyter', 'profiles', 'admission-webhook', 'volumes-web-app', 'tensorboard', 'pvcviewer']) or \
       any(x in name for x in ['centraldashboard', 'jupyter', 'notebook-controller', 'profiles', 'admission-webhook', 'volumes-web-app', 'tensorboard', 'pvcviewer']):
        return 'Kubeflow Core'
    
    # Experimental/Other
    if 'experimental' in filepath or any(x in name for x in ['seaweedfs', 'ray']):
        return 'Other'
    
    # Everything else in applications directory
    if 'applications' in filepath:
        return 'Other'
    
    return None

def get_manifest_resource_requests():
    """Parse all manifest files and get resource requests by component"""
    if not YAML_AVAILABLE:
        print("Skipping manifest parsing (PyYAML not available)")
        return defaultdict(lambda: {'cpu': 0, 'memory': 0})
        
    manifest_files = find_manifest_files()
    component_requests = defaultdict(lambda: {'cpu': 0, 'memory': 0})
    
    print(f"Parsing {len(manifest_files)} manifest files for resource requests...")
    
    for filepath in manifest_files:
        resources = parse_manifest_file(filepath)
        for resource in resources:
            component = resource['component']
            component_requests[component]['cpu'] += resource['cpu']
            component_requests[component]['memory'] += resource['memory']
    
    return component_requests

def get_storage_requirements():
    """Get storage requirements from PVC allocations in manifest files"""
    if not YAML_AVAILABLE:
        print("Skipping storage parsing (PyYAML not available), using fallback values")
        # Fallback to hardcoded values from README
        return {
            'Katib': 3,
            'Metadata': 40,
            'Pipelines': 60,
            'Other': 36,
        }
        
    manifest_files = find_manifest_files()
    component_storage = defaultdict(int)
    
    for filepath in manifest_files:
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                documents = list(yaml.safe_load_all(f))
                
            for doc in documents:
                if not doc or not isinstance(doc, dict):
                    continue
                    
                if doc.get('kind') == 'PersistentVolumeClaim':
                    name = doc.get('metadata', {}).get('name', '')
                    spec = doc.get('spec', {})
                    resources = spec.get('resources', {})
                    requests = resources.get('requests', {})
                    storage = requests.get('storage', '0')
                    
                    # Parse storage size to GB
                    storage_gb = parse_storage_to_gb(storage)
                    
                    # Categorize by PVC name/filepath
                    component = categorize_storage_resource(filepath, name)
                    if component:
                        component_storage[component] += storage_gb
                        
        except Exception:
            continue
    
    return dict(component_storage)

def parse_storage_to_gb(storage_str):
    """Convert storage string to GB"""
    if not storage_str or storage_str == '0':
        return 0
        
    storage_str = storage_str.strip()
    
    if storage_str.endswith('Gi'):
        return int(storage_str[:-2])
    elif storage_str.endswith('Mi'):
        return int(storage_str[:-2]) / 1024
    elif storage_str.endswith('Ki'):
        return int(storage_str[:-2]) / (1024 * 1024)
    elif storage_str.endswith('Ti'):
        return int(storage_str[:-2]) * 1024
    else:
        # Assume GB
        return int(float(storage_str.rstrip('G')))

def categorize_storage_resource(filepath, pvc_name):
    """Categorize storage resources by component"""
    filepath = filepath.lower()
    pvc_name = pvc_name.lower()
    
    if 'katib' in filepath or 'katib' in pvc_name:
        return 'Katib'
    elif 'pipeline' in filepath or any(x in pvc_name for x in ['minio', 'pipeline']):
        return 'Pipelines'
    elif 'metadata' in pvc_name or 'mysql' in pvc_name:
        return 'Metadata'
    elif 'seaweedfs' in filepath or 'seaweedfs' in pvc_name:
        return 'Other'
    else:
        return 'Other'

def calculate_max_resources(actual_usage, manifest_requests):
    """Calculate maximum of actual usage and manifest requests for each component"""
    all_components = set(actual_usage.keys()) | set(manifest_requests.keys())
    max_resources = defaultdict(lambda: {'cpu': 0, 'memory': 0})
    
    for component in all_components:
        actual_cpu = actual_usage.get(component, {}).get('cpu', 0)
        actual_memory = actual_usage.get(component, {}).get('memory', 0)
        
        request_cpu = manifest_requests.get(component, {}).get('cpu', 0)
        request_memory = manifest_requests.get(component, {}).get('memory', 0)
        
        max_resources[component]['cpu'] = max(actual_cpu, request_cpu)
        max_resources[component]['memory'] = max(actual_memory, request_memory)
    
    return max_resources

def generate_markdown_table(component_resources, storage_map, actual_usage, manifest_requests):
    """Generate markdown table from component resources"""
    
    print("## Resource Usage by Components")
    print()
    print("The following table shows the resource requirements for each Kubeflow component, calculated as the maximum of actual usage and configured requests for CPU/memory, plus storage requirements from PVCs:")
    print()
    print("| Component | CPU (cores) | Memory (Mi) | Storage (GB) |")
    print("|-----------|-------------|-------------|--------------|")
    
    total_cpu = 0
    total_memory = 0
    total_storage = 0
    
    # Sort components for consistent output
    for component in sorted(component_resources.keys()):
        cpu_millicores = component_resources[component]['cpu']
        memory_mib = component_resources[component]['memory']
        storage_gb = storage_map.get(component, 0)
        
        total_cpu += cpu_millicores
        total_memory += memory_mib
        total_storage += storage_gb
        
        print(f"| {component} | {cpu_millicores}m | {memory_mib}Mi | {storage_gb}GB |")
    
    print(f"| **Total** | **{total_cpu}m** | **{total_memory}Mi** | **{total_storage}GB** |")
    print()
    print("### Resource Notes")
    print()
    print("- **CPU values** represent the maximum of actual observed usage and configured resource requests")
    print("- **Memory values** represent the maximum of actual observed usage and configured resource requests")  
    print("- **Storage values** represent the total PVC allocations from manifest files")
    print("- Components with high resource requests (like Istio, KServe) may be over-provisioned compared to actual usage")
    print("- Storage requirements are persistent and represent the minimum disk space needed")
    
    # Print detailed breakdown for debugging
    print()
    print("### Detailed Breakdown")
    print()
    for component in sorted(component_resources.keys()):
        actual_cpu = actual_usage.get(component, {}).get('cpu', 0)
        actual_memory = actual_usage.get(component, {}).get('memory', 0)
        request_cpu = manifest_requests.get(component, {}).get('cpu', 0)
        request_memory = manifest_requests.get(component, {}).get('memory', 0)
        
        print(f"**{component}:**")
        print(f"  - Actual usage: {actual_cpu}m CPU, {actual_memory}Mi Memory")
        print(f"  - Manifest requests: {request_cpu}m CPU, {request_memory}Mi Memory")
        print(f"  - Maximum used: {component_resources[component]['cpu']}m CPU, {component_resources[component]['memory']}Mi Memory")
        print()

def main():
    if YAML_AVAILABLE:
        print("Generating resource usage table from live cluster data and manifest requests...")
    else:
        print("Generating resource usage table from live cluster data only...")
    print()
    
    # Get actual usage from kubectl top
    print("Getting actual resource usage from cluster...")
    kubectl_output = run_kubectl_top()
    actual_usage = parse_kubectl_output(kubectl_output)
    
    # Get resource requests from manifest files
    if YAML_AVAILABLE:
        print("Parsing manifest files for resource requests...")
        manifest_requests = get_manifest_resource_requests()
        
        # Calculate maximum of actual usage and requests
        print("Calculating maximum resource requirements...")
        max_resources = calculate_max_resources(actual_usage, manifest_requests)
    else:
        print("Using actual usage only (manifest parsing skipped)...")
        manifest_requests = defaultdict(lambda: {'cpu': 0, 'memory': 0})
        max_resources = actual_usage
    
    # Get storage requirements from PVCs
    if YAML_AVAILABLE:
        print("Calculating storage requirements from PVCs...")
    else:
        print("Using fallback storage values...")
    storage_map = get_storage_requirements()
    
    # Generate table
    print("Generating resource table...")
    print()
    generate_markdown_table(max_resources, storage_map, actual_usage, manifest_requests)

if __name__ == "__main__":
    main() 