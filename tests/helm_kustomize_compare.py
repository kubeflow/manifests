#!/usr/bin/env python3

import yaml
import sys
import json
from typing import Dict, List, Tuple, Any
import re

def load_manifests(file_path: str) -> List[Dict]:
    """Load YAML manifests from file."""
    with open(file_path, 'r') as f:
        content = f.read()
    
    docs = []
    try:
        for doc in yaml.safe_load_all(content):
            if doc:  
                docs.append(doc)
    except yaml.YAMLError:
        for doc_str in content.split('---'):
            doc_str = doc_str.strip()
            if doc_str:
                try:
                    doc = yaml.safe_load(doc_str)
                    if doc:  
                        docs.append(doc)
                except yaml.YAMLError:
                    continue
    
    return docs

def clean_helm_metadata(obj: Any, component: str = "katib") -> Any:
    """Remove Helm-specific metadata that should be ignored in comparison."""
    if isinstance(obj, dict):
        cleaned = {}
        for key, value in obj.items():
            if key == "metadata" and isinstance(value, dict):
                # Clean metadata section
                cleaned_metadata = {}
                for meta_key, meta_value in value.items():
                    if meta_key == "labels" and isinstance(meta_value, dict):
                        # Remove Helm-specific labels (component-specific logic)
                        cleaned_labels = {}
                        for label_key, label_value in meta_value.items():
                            if component == "kserve-models-web-app":
                                # More restrictive filtering for KServe
                                if not label_key.startswith(('helm.sh/', 'app.kubernetes.io/managed-by')):
                                    cleaned_labels[label_key] = label_value
                            elif component in ["pipeline", "pipelines"]:
                                helm_specific_labels = [
                                    'app.kubernetes.io/managed-by',
                                    'app.kubernetes.io/version', 
                                    'app.kubernetes.io/name',
                                    'app.kubernetes.io/instance',
                                    'app.kubernetes.io/component'
                                ]
                                if not label_key.startswith('helm.sh/') and label_key not in helm_specific_labels:
                                    cleaned_labels[label_key] = label_value
                            else:
                                # Standard filtering for Katib and Model Registry
                                helm_labels = [
                                    'app.kubernetes.io/managed-by',
                                    'app.kubernetes.io/version', 
                                    'app.kubernetes.io/name',
                                    'app.kubernetes.io/instance',
                                    'app.kubernetes.io/component'
                                ]
                                if not label_key.startswith('helm.sh/') and label_key not in helm_labels:
                                    cleaned_labels[label_key] = label_value
                        if cleaned_labels:  # Only add if there are remaining labels
                            cleaned_metadata[meta_key] = cleaned_labels
                    elif meta_key == "annotations" and isinstance(meta_value, dict):
                        # Remove only Helm-specific annotations
                        cleaned_annotations = {}
                        for ann_key, ann_value in meta_value.items():
                            if not ann_key.startswith(('helm.sh/', 'meta.helm.sh/')):
                                # For pipelines, ignore GCP icon annotation
                                if component in ["pipeline", "pipelines"] and ann_key == 'kubernetes-engine.cloud.google.com/icon':
                                    continue
                                cleaned_annotations[ann_key] = ann_value
                        if cleaned_annotations:  # Only add if there are remaining annotations
                            cleaned_metadata[meta_key] = cleaned_annotations
                    else:
                        # Keep all other metadata fields as-is
                        cleaned_metadata[meta_key] = meta_value
                cleaned[key] = cleaned_metadata
            else:
                cleaned[key] = clean_helm_metadata(value, component)
        return cleaned
    elif isinstance(obj, list):
        return [clean_helm_metadata(item, component) for item in obj]
    else:
        return obj

def normalize_kustomize_refs(obj: Any, path: str = "") -> Any:
    """Normalize Kustomize hash suffixes in secret/configmap references throughout the manifest."""
    if isinstance(obj, dict):
        normalized = {}
        for key, value in obj.items():
            current_path = f"{path}.{key}" if path else key
            
            # Normalize secret/configmap references in common locations
            if key == "name" and isinstance(value, str):
                if any(ref_pattern in path for ref_pattern in [
                    'secretKeyRef', 'configMapKeyRef', 'configMapRef', 'secret', 'volumes'
                ]):
                    value = re.sub(r'-[a-z0-9]{10}$', '', value)
                elif 'volumes' in path and key == 'secretName':
                    value = re.sub(r'-[a-z0-9]{10}$', '', value)
                elif 'volumes' in path and 'configMap' in path:
                    value = re.sub(r'-[a-z0-9]{10}$', '', value)
            
            normalized[key] = normalize_kustomize_refs(value, current_path)
        return normalized
    elif isinstance(obj, list):
        return [normalize_kustomize_refs(item, path) for item in obj]
    else:
        return obj

def normalize_manifest(manifest: Dict, component: str = "katib") -> Dict:
    """Normalize manifest by removing/standardizing certain fields."""
    normalized = manifest.copy()
    
    # Clean Helm-specific metadata
    normalized = clean_helm_metadata(normalized, component)
    
    # Normalize Kustomize hash references
    normalized = normalize_kustomize_refs(normalized)
    
    # Handle ConfigMap data normalization (only for Katib)
    if component == "katib" and normalized.get('kind') == 'ConfigMap' and 'data' in normalized:
        data = normalized['data']
        normalized_data = {}
        for key, value in data.items():
            if isinstance(value, str):
                # Remove leading/trailing whitespace and normalize YAML content
                normalized_value = value.strip()
                # Remove leading --- if present (common in ConfigMap YAML data)
                if normalized_value.startswith('---'):
                    normalized_value = normalized_value[3:].strip()
                normalized_data[key] = normalized_value
            else:
                normalized_data[key] = value
        normalized['data'] = normalized_data
    
    if 'metadata' in normalized and 'name' in normalized['metadata']:
        kind = normalized.get('kind', '')
        if kind in ['Secret', 'ConfigMap']:
            name = normalized['metadata']['name']
            normalized['metadata']['name'] = re.sub(r'-[a-z0-9]{10}$', '', name)
    
    if 'metadata' in normalized:
        metadata = normalized['metadata']
        
        metadata.pop('generation', None)
        metadata.pop('resourceVersion', None)
        metadata.pop('uid', None)
        metadata.pop('creationTimestamp', None)
        metadata.pop('managedFields', None)
        
        # For pipelines component, remove labels from Argo CRDs
        if component in ["pipeline", "pipelines"] and normalized.get('kind') == 'CustomResourceDefinition':
            crd_name = normalized.get('metadata', {}).get('name', '')
            if 'argoproj.io' in crd_name:
                metadata.pop('labels', None)
    
    normalized.pop('status', None)
    
    # For pipelines Argo CRDs, normalize schema differences between versions
    if component in ["pipeline", "pipelines"] and normalized.get('kind') == 'CustomResourceDefinition':
        crd_name = normalized.get('metadata', {}).get('name', '')
        if 'argoproj.io' in crd_name and 'spec' in normalized:
            def normalize_crd_schema(obj, path=""):
                if isinstance(obj, dict):
                    cleaned = {}
                    for k, v in obj.items():
                        if k.startswith('x-kubernetes-'):
                            continue
                        if k == 'default' and 'properties' in path:
                            continue
                        if k == 'required' and 'openAPIV3Schema.properties' in path:
                            continue
                        cleaned[k] = normalize_crd_schema(v, f"{path}.{k}" if path else k)
                    return cleaned
                elif isinstance(obj, list):
                    return [normalize_crd_schema(item, path) for item in obj]
                else:
                    return obj
            
            normalized['spec'] = normalize_crd_schema(normalized['spec'])
    
    def remove_empty_values(obj):
        if isinstance(obj, dict):
            return {k: remove_empty_values(v) for k, v in obj.items() 
                   if v is not None and v != {} and v != []}
        elif isinstance(obj, list):
            return [remove_empty_values(item) for item in obj if item is not None]
        else:
            return obj
    
    return remove_empty_values(normalized)

def get_resource_key(manifest: Dict, component: str = "katib") -> str:
    """Generate a unique key for the resource."""
    kind = manifest.get('kind', 'Unknown')
    name = manifest.get('metadata', {}).get('name', 'unknown')
    namespace = manifest.get('metadata', {}).get('namespace', '')
    
    if kind in ['Secret', 'ConfigMap']:
        name = re.sub(r'-[a-z0-9]{10}$', '', name)
    
    # Include namespace in key only for Katib
    if component == "katib" and namespace:
        return f"{kind}/{namespace}/{name}"
    else:
        return f"{kind}/{name}"

def deep_diff(obj1: Any, obj2: Any, path: str = "", resource_key: str = "") -> List[str]:
    """Compare two objects and return list of differences."""
    differences = []
    
    if type(obj1) != type(obj2):
        differences.append(f"{path}: type mismatch ({type(obj1).__name__} vs {type(obj2).__name__})")
        return differences
    
    if isinstance(obj1, dict):
        all_keys = set(obj1.keys()) | set(obj2.keys())
        for key in sorted(all_keys):
            key_path = f"{path}.{key}" if path else key
            if key not in obj1:
                differences.append(f"{key_path}: missing in kustomize")
            elif key not in obj2:
                if 'CustomResourceDefinition' in resource_key and 'argoproj.io' in resource_key:
                    if 'openAPIV3Schema.properties' in path and key == 'properties':
                        continue
                    elif '.properties.' in path:
                        continue
                differences.append(f"{key_path}: missing in helm")
            else:
                differences.extend(deep_diff(obj1[key], obj2[key], key_path, resource_key))
    
    elif isinstance(obj1, list):
        if len(obj1) != len(obj2):
            differences.append(f"{path}: list length mismatch ({len(obj1)} vs {len(obj2)})")
        else:
            for i, (item1, item2) in enumerate(zip(obj1, obj2)):
                differences.extend(deep_diff(item1, item2, f"{path}[{i}]", resource_key))
    
    elif obj1 != obj2:
        differences.append(f"{path}: '{obj1}' != '{obj2}'")
    
    return differences

def get_expected_helm_extras(component: str, scenario: str) -> set:
    """Get expected extra resources in Helm for different components and scenarios."""
    if component == "katib":
        return {
            'Secret/kubeflow/katib-webhook-cert',  # Webhook certificates
        }
    elif component == "model-registry":
        return set()  # No extra resources in Helm for Model Registry
    elif component == "kserve-models-web-app":
        if scenario == "base":
            return {"AuthorizationPolicy/kserve-models-web-app"}
        return set()
    elif component == "pipelines":
        extras = {
            # Built-in third-party Argo resources (from thirdParty.argo templates)
            'ConfigMap/workflow-controller-configmap',
            'Deployment/workflow-controller',
            'PriorityClass/workflow-controller',
            'Role/argo-role',
            'RoleBinding/argo-binding',
            'ServiceAccount/argo',
        }
        
        # These get extra resources from the subchart dependency
        argo_subchart_scenarios = [
            'dev',  
            'platform-agnostic-multi-user-emissary',  
            'platform-agnostic-multi-user',
            'platform-agnostic-multi-user-legacy', 
        ]
        
        if any(sc in scenario for sc in argo_subchart_scenarios):
            # Argo Workflows subchart creates these extra resources
            extras.update({
                # Argo Workflows subchart RBAC resources
                'ClusterRole/pipeline-argo-workflows-admin',
                'ClusterRole/pipeline-argo-workflows-edit',
                'ClusterRole/pipeline-argo-workflows-view',
                'ClusterRole/pipeline-argo-workflows-server',
                'ClusterRole/pipeline-argo-workflows-server-cluster-template',
                'ClusterRole/pipeline-argo-workflows-workflow-controller',
                'ClusterRole/pipeline-argo-workflows-workflow-controller-cluster-template',
                'ClusterRoleBinding/pipeline-argo-workflows-server',
                'ClusterRoleBinding/pipeline-argo-workflows-server-cluster-template',
                'ClusterRoleBinding/pipeline-argo-workflows-workflow-controller',
                'ClusterRoleBinding/pipeline-argo-workflows-workflow-controller-cluster-template',
                'Role/pipeline-argo-workflows-workflow',
                'RoleBinding/pipeline-argo-workflows-workflow',
                # Argo Workflows subchart deployments and services
                'Deployment/pipeline-argo-workflows-server',
                'Deployment/pipeline-argo-workflows-workflow-controller',
                'Service/pipeline-argo-workflows-server',
                'ServiceAccount/pipeline-argo-workflows-server',
                'ServiceAccount/pipeline-argo-workflows-workflow-controller',
                'ConfigMap/pipeline-argo-workflows-workflow-controller-configmap',
                # Argo Workflows CRDs from subchart
                'CustomResourceDefinition/clusterworkflowtemplates.argoproj.io',
                'CustomResourceDefinition/cronworkflows.argoproj.io',
                'CustomResourceDefinition/workflowartifactgctasks.argoproj.io',
                'CustomResourceDefinition/workfloweventbindings.argoproj.io',
                'CustomResourceDefinition/workflows.argoproj.io',
                'CustomResourceDefinition/workflowtaskresults.argoproj.io',
                'CustomResourceDefinition/workflowtasksets.argoproj.io',
                'CustomResourceDefinition/workflowtemplates.argoproj.io',
            })
        
        if 'multi-user' in scenario:
            extras.update({
                'ConfigMap/kubeflow-pipelines-profile-controller-code',
                'ConfigMap/kubeflow-pipelines-profile-controller-env',
                'ConfigMap/pipeline-api-server-config',
            })
        
        return extras
    else:
        return set()

def compare_manifests(kustomize_file: str, helm_file: str, component: str, scenario: str, namespace: str = "", verbose: bool = False) -> bool:
    """Compare Kustomize and Helm manifests."""
    kustomize_manifests = load_manifests(kustomize_file)
    helm_manifests = load_manifests(helm_file)
    
    kustomize_resources = {}
    helm_resources = {}
    
    for manifest in kustomize_manifests:
        normalized = normalize_manifest(manifest, component)
        key = get_resource_key(normalized, component)
        kustomize_resources[key] = normalized
    
    for manifest in helm_manifests:
        normalized = normalize_manifest(manifest, component)
        key = get_resource_key(normalized, component)
        helm_resources[key] = normalized
    
    kustomize_keys = set(kustomize_resources.keys())
    helm_keys = set(helm_resources.keys())
    
    common_keys = kustomize_keys & helm_keys
    only_in_kustomize = kustomize_keys - helm_keys
    only_in_helm = helm_keys - kustomize_keys
    
    expected_helm_extras = get_expected_helm_extras(component, scenario)
    unexpected_helm_extras = only_in_helm - expected_helm_extras
    
    differences_found = []
    success = True
    
    if only_in_kustomize:
        print(f"Resources only in Kustomize: {len(only_in_kustomize)}")
        if verbose:
            for resource in sorted(only_in_kustomize):
                print(f"  - {resource}")        
        success = False
        differences_found.extend(only_in_kustomize)
    
    if unexpected_helm_extras:
        print(f"Unexpected resources only in Helm: {len(unexpected_helm_extras)}")
        if verbose:
            for resource in sorted(unexpected_helm_extras):
                print(f"  - {resource}")
        success = False
        differences_found.extend(unexpected_helm_extras)
    
    # Compare common resources
    for key in sorted(common_keys):
        kustomize_resource = kustomize_resources[key]
        helm_resource = helm_resources[key]
        
        differences = deep_diff(kustomize_resource, helm_resource, resource_key=key)
        
        if differences:
            print(f"Differences in {key}: {len(differences)} fields")
            if verbose:
                print(f"  Detailed differences for {key}:")
                for diff in differences:
                    print(f"    {diff}")
                print()
            differences_found.append(key)
            success = False
    
    if not success:
        print(f"Found differences in {len(differences_found)} resources")
        return False
    
    return True

if __name__ == "__main__":
    if len(sys.argv) < 5:
        print("Usage: python compare.py <kustomize_file> <helm_file> <component> <scenario> [namespace] [--verbose]")
        print("Components: katib, model-registry, kserve-models-web-app, pipeline")
        sys.exit(1)
    
    kustomize_file = sys.argv[1]
    helm_file = sys.argv[2]
    component = sys.argv[3]
    scenario = sys.argv[4]
    namespace = sys.argv[5] if len(sys.argv) > 5 and not sys.argv[5].startswith('--') else ""
    verbose = '--verbose' in sys.argv

    if component not in ["katib", "model-registry", "kserve-models-web-app", "notebook-controller", "pipelines"]:
        print(f"ERROR: Unknown component: {component}")
        print("Supported components: katib, model-registry, kserve-models-web-app, notebook-controller, pipelines")
        sys.exit(1)
    if component == "pipeline":
        component = "pipelines"

    success = compare_manifests(kustomize_file, helm_file, component, scenario, namespace, verbose)
    sys.exit(0 if success else 1) 