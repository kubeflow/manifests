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

# Component categorization rules - centralized mapping
COMPONENT_RULES = {
    'Dex + OAuth2-Proxy': {
        'namespaces': ['auth', 'oauth2-proxy'],
        'keywords': ['dex', 'oauth2-proxy']
    },
    'Cert Manager': {
        'namespaces': ['cert-manager'],
        'keywords': ['cert-manager']
    },
    'Istio': {
        'namespaces': ['istio-system'],
        'keywords': ['istio', 'cluster-local-gateway', 'cluster-jwks-proxy']
    },
    'Knative': {
        'namespaces': ['knative-serving', 'knative-eventing'],
        'keywords': ['knative']
    },
    'Katib': {
        'keywords': ['katib']
    },
    'KServe': {
        'keywords': ['kserve', 'predictor', 'models-web-app']
    },
    'Metadata': {
        'keywords': ['metadata', 'mysql'],
        'exclude': ['model-registry']
    },
    'Model Registry': {
        'keywords': ['model-registry']
    },
    'Pipelines': {
        'keywords': ['ml-pipeline', 'workflow-controller', 'cache-server', 'minio', 'pipeline', 'seaweedfs']
    },
    'Spark': {
        'keywords': ['spark']
    },
    'Training Operator': {
        'keywords': ['training-operator', 'training']
    },
    'Jupyter': {
        'keywords': ['jupyter', 'notebook-controller']
    },
    'Central Dashboard': {
        'keywords': ['centraldashboard']
    },
    'Profiles + KFAM': {
        'keywords': ['profiles', 'profile-controller']
    },
    'Admission Webhook': {
        'keywords': ['admission-webhook', 'poddefaults']
    },
    'Volumes Web App': {
        'keywords': ['volumes-web-app']
    },
    'Tensorboard': {
        'keywords': ['tensorboard']
    },
    'PVC Viewer Controller': {
        'keywords': ['pvcviewer']
    }
}

# Storage fallback values when YAML parsing is unavailable
STORAGE_FALLBACK = {
    'Katib': 3,
    'Metadata': 40,
    'Pipelines': 80,  
}

def run_kubectl_top():
    """Run kubectl top pods commands and return the output"""
    try:
        result = subprocess.run(['kubectl', 'top', 'pods', '--all-namespaces'], 
                              capture_output=True, text=True, check=True)
        return result.stdout
    except subprocess.CalledProcessError as e:
        print(f"Error running kubectl top: {e}")
        print("Make sure metrics-server is installed and running in your cluster")
        sys.exit(1)
    except FileNotFoundError:
        print("Error: kubectl command not found. Please install kubectl")
        sys.exit(1)

def parse_resource_value(value_str, resource_type):
    """Parse CPU (to millicores) or memory (to MiB) resource values"""
    if not value_str or value_str == '0':
        return 0
    
    try:
        if resource_type == 'cpu':
            if value_str.endswith('m'):
                return int(value_str[:-1])
            elif value_str.endswith('n'):
                return int(value_str[:-1]) // 1000000
            else:
                return int(float(value_str) * 1000)
        
        elif resource_type == 'memory':
            if value_str.endswith('Mi'):
                return int(value_str[:-2])
            elif value_str.endswith('Gi'):
                return int(float(value_str[:-2]) * 1024)
            elif value_str.endswith('Ki'):
                return int(value_str[:-2]) // 1024
            elif value_str.endswith(('Bi', 'B')):
                return int(value_str[:-2]) // (1024 * 1024)
            else:
                return int(value_str) // (1024 * 1024)
        
        elif resource_type == 'storage':
            if value_str.endswith('Gi'):
                return int(value_str[:-2])
            elif value_str.endswith('Mi'):
                return int(value_str[:-2]) / 1024
            elif value_str.endswith('Ki'):
                return int(value_str[:-2]) / (1024 * 1024)
            elif value_str.endswith('Ti'):
                return int(value_str[:-2]) * 1024
            else:
                return int(float(value_str.rstrip('G')))
    
    except (ValueError, TypeError):
        return 0
    
    return 0

def categorize_resource(namespace, name, filepath=""):
    """Categorize resources into components based on namespace, name, and filepath
    """
    if namespace in ['kube-system', 'local-path-storage']:
        return None
    
    namespace = namespace.lower()
    name = name.lower()
    filepath = filepath.lower().replace('\\', '/')
    
    for component, rules in COMPONENT_RULES.items():
        if 'namespaces' in rules and namespace in rules['namespaces']:
            return component
        
        if 'keywords' in rules:
            for keyword in rules['keywords']:
                if keyword in name or keyword in filepath:
                    # Check for exclusions
                    if 'exclude' in rules:
                        if any(excl in name for excl in rules['exclude']):
                            continue
                    return component
    
    # Return None for components not in the official list
    return None

def parse_kubectl_output(output):
    """Parse kubectl top output and categorize by component"""
    lines = output.strip().split('\n')[1:]  # Skip header
    component_resources = defaultdict(lambda: {'cpu': 0, 'memory': 0})
    
    for line in lines:
        parts = line.split()
        if len(parts) < 4:
            continue
            
        namespace, pod_name, cpu_str, memory_str = parts[:4]
        component = categorize_resource(namespace, pod_name)
        
        if component:
            cpu_millicores = parse_resource_value(cpu_str, 'cpu')
            memory_mib = parse_resource_value(memory_str, 'memory')
            
            component_resources[component]['cpu'] += cpu_millicores
            component_resources[component]['memory'] += memory_mib
    
    return component_resources

def find_manifest_files():
    """Find all YAML manifest files in relevant directories"""
    manifest_files = []
    for base_dir in ['applications', 'common']:
        if os.path.exists(base_dir):
            manifest_files.extend(glob.glob(f"{base_dir}/**/*.yaml", recursive=True))
            manifest_files.extend(glob.glob(f"{base_dir}/**/*.yml", recursive=True))
    return manifest_files

def extract_containers_from_k8s_resource(doc):
    """Extract containers from various Kubernetes resource types"""
    kind = doc.get('kind', '')
    
    container_paths = {
        'Deployment': ['spec', 'template', 'spec', 'containers'],
        'StatefulSet': ['spec', 'template', 'spec', 'containers'],
        'DaemonSet': ['spec', 'template', 'spec', 'containers'],
        'Pod': ['spec', 'containers'],
        'Job': ['spec', 'template', 'spec', 'containers'],
        'CronJob': ['spec', 'jobTemplate', 'spec', 'template', 'spec', 'containers']
    }
    
    if kind not in container_paths:
        return []
    
    current = doc
    for key in container_paths[kind]:
        current = current.get(key, {})
        if not current:
            return []
    
    return current if isinstance(current, list) else []

def parse_manifest_resources(manifest_files):
    """Parse all manifest files and extract resource requests and storage"""
    component_requests = defaultdict(lambda: {'cpu': 0, 'memory': 0})
    component_storage = defaultdict(int)
    
    for filepath in manifest_files:
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                documents = list(yaml.safe_load_all(f))
            
            for doc in documents:
                if not doc or not isinstance(doc, dict):
                    continue
                
                metadata = doc.get('metadata', {})
                name = metadata.get('name', 'unknown')
                namespace = metadata.get('namespace', 'default')
                kind = doc.get('kind', '')
                
                component = categorize_resource(namespace, name, filepath)
                if not component:
                    continue
                
                if kind == 'PersistentVolumeClaim':
                    storage_str = doc.get('spec', {}).get('resources', {}).get('requests', {}).get('storage', '0')
                    storage_gb = parse_resource_value(storage_str, 'storage')
                    component_storage[component] += storage_gb
                
                containers = extract_containers_from_k8s_resource(doc)
                for container in containers:
                    requests = container.get('resources', {}).get('requests', {})
                    cpu_request = requests.get('cpu', '0')
                    memory_request = requests.get('memory', '0')
                    
                    if cpu_request != '0' or memory_request != '0':
                        cpu_millicores = parse_resource_value(cpu_request, 'cpu')
                        memory_mib = parse_resource_value(memory_request, 'memory')
                        
                        component_requests[component]['cpu'] += cpu_millicores
                        component_requests[component]['memory'] += memory_mib
        
        except Exception:
            continue  
    
    return component_requests, dict(component_storage)

def calculate_max_resources(actual_usage, manifest_requests):
    """Calculate maximum of actual usage and manifest requests"""
    all_components = set(actual_usage.keys()) | set(manifest_requests.keys())
    max_resources = defaultdict(lambda: {'cpu': 0, 'memory': 0})
    
    for component in all_components:
        actual = actual_usage.get(component, {'cpu': 0, 'memory': 0})
        requests = manifest_requests.get(component, {'cpu': 0, 'memory': 0})
        
        max_resources[component]['cpu'] = max(actual['cpu'], requests['cpu'])
        max_resources[component]['memory'] = max(actual['memory'], requests['memory'])
    
    return max_resources

def generate_table(component_resources, storage_map, actual_usage, manifest_requests):
    """Generate markdown table from component resources"""
    print("## Resource Usage by Components")
    print()
    print("The following table shows the resource requirements for each Kubeflow components:")
    print()
    print("| Component | CPU (cores) | Memory (Mi) | Storage (GB) |")
    print("|-----------|-------------|-------------|--------------|")
    
    totals = {'cpu': 0, 'memory': 0, 'storage': 0}
    
    for component in sorted(component_resources.keys()):
        resources = component_resources[component]
        storage = storage_map.get(component, 0)
        
        totals['cpu'] += resources['cpu']
        totals['memory'] += resources['memory']
        totals['storage'] += storage
        
        print(f"| {component} | {resources['cpu']}m | {resources['memory']}Mi | {storage}GB |")
    
    print(f"| **Total** | **{totals['cpu']}m** | **{totals['memory']}Mi** | **{totals['storage']}GB** |")
    print()
    
    print("### Notes")
    print("- CPU/Memory values are maximum of actual usage and configured requests")
    print("- Storage values are total PVC allocations from manifest files")
    print("- Components not matching the official list are excluded from the table")
    print()

def main():
    print("Generating resource usage table...")
    print()
    
    print("Getting actual resource usage from cluster...")
    kubectl_output = run_kubectl_top()
    actual_usage = parse_kubectl_output(kubectl_output)
    
    if YAML_AVAILABLE:
        print("Parsing manifest files...")
        manifest_files = find_manifest_files()
        manifest_requests, storage_map = parse_manifest_resources(manifest_files)
        max_resources = calculate_max_resources(actual_usage, manifest_requests)
    else:
        print("Using actual usage only (manifest parsing skipped)...")
        manifest_requests = defaultdict(lambda: {'cpu': 0, 'memory': 0})
        max_resources = actual_usage
        storage_map = STORAGE_FALLBACK
    
    generate_table(max_resources, storage_map, actual_usage, manifest_requests)

if __name__ == "__main__":
    main()