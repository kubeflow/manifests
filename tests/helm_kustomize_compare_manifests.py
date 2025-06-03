#!/usr/bin/env python3
import yaml
import sys
import json

def clean_metadata(obj, source='helm'):
    """Remove source-specific metadata from objects"""
    if isinstance(obj, dict):
        if 'metadata' in obj:
            metadata = obj['metadata']
            
            if 'labels' in metadata:
                labels = metadata['labels']
                helm_labels = [
                    'helm.sh/chart', 
                    'app.kubernetes.io/managed-by',
                    'app.kubernetes.io/instance',
                    'app.kubernetes.io/version'
                ]
                for label in helm_labels:
                    labels.pop(label, None)
                if not labels:
                    metadata.pop('labels', None)
            
            if 'annotations' in metadata:
                annotations = metadata['annotations']
                helm_annotations = [
                    'meta.helm.sh/release-name', 
                    'meta.helm.sh/release-namespace',
                    'deployment.kubernetes.io/revision'
                ]
                for annotation in helm_annotations:
                    annotations.pop(annotation, None)
                if not annotations:
                    metadata.pop('annotations', None)
        
        for key, value in list(obj.items()):
            if isinstance(value, (dict, list)):
                clean_metadata(value, source)
    
    elif isinstance(obj, list):
        for item in obj:
            clean_metadata(item, source)
    
    return obj

def normalize_for_comparison(doc, component, expected_namespace):
    """Normalize document for comparison"""
    if not doc:
        return doc
    
    if 'metadata' in doc:
        namespaced_kinds = [
            'Service', 'ServiceAccount', 'Deployment', 'Job', 'ConfigMap', 
            'Secret', 'Role', 'RoleBinding', 'NetworkPolicy', 'PodDisruptionBudget'
        ]
        if doc.get('kind') in namespaced_kinds:
            if 'namespace' not in doc['metadata'] or not doc['metadata']['namespace']:
                doc['metadata']['namespace'] = expected_namespace
    
    return doc

def load_and_index(file_path):
    """Load YAML and index by kind/name"""
    with open(file_path) as f:
        content = f.read()
        if not content.strip():
            return {}
        docs = list(yaml.safe_load_all(content))
    
    indexed = {}
    for doc in docs:
        if doc and isinstance(doc, dict) and 'kind' in doc and 'metadata' in doc:
            kind = doc['kind']
            name = doc['metadata'].get('name', 'unnamed')
            key = f"{kind}/{name}"
            indexed[key] = doc
    
    return indexed

def main():
    if len(sys.argv) < 3:
        print("Usage: compare-manifests.py <kustomize-file> <helm-file> [component] [namespace] [kubeflow-rbac-enabled]")
        sys.exit(1)
    
    kustomize_file = sys.argv[1]
    helm_file = sys.argv[2]
    component = sys.argv[3] if len(sys.argv) > 3 else "component"
    namespace = sys.argv[4] if len(sys.argv) > 4 else "kubeflow"
    kubeflow_rbac_enabled = sys.argv[5].lower() == 'true' if len(sys.argv) > 5 else False

    kustomize = load_and_index(kustomize_file)
    helm = load_and_index(helm_file)

    kustomize_keys = set(kustomize.keys())
    helm_keys = set(helm.keys())

    expected_differences = {
        "spark-operator": {
            'ClusterRole/kubeflow-spark-admin',
            'ClusterRole/kubeflow-spark-edit', 
            'ClusterRole/kubeflow-spark-view'
        }
    }

    expected_extra_kustomize = set()
    if component in expected_differences and not kubeflow_rbac_enabled:
        expected_extra_kustomize = expected_differences[component]

    extra_kustomize = kustomize_keys - helm_keys
    extra_helm = helm_keys - kustomize_keys
    unexpected_extra_kustomize = extra_kustomize - expected_extra_kustomize

    success = True
    
    if unexpected_extra_kustomize:
        print(f"ERROR: Unexpected resources in Kustomize: {', '.join(sorted(unexpected_extra_kustomize))}")
        success = False

    if extra_helm:
        print(f"ERROR: Unexpected resources in Helm: {', '.join(sorted(extra_helm))}")
        success = False

    common = kustomize_keys & helm_keys
    differences = []
    
    for key in sorted(common):
        kustomize_doc = json.loads(json.dumps(kustomize[key]))
        helm_doc = json.loads(json.dumps(helm[key]))
        
        kustomize_clean = clean_metadata(kustomize_doc, 'kustomize')
        helm_clean = clean_metadata(helm_doc, 'helm')
        
        kustomize_normalized = normalize_for_comparison(kustomize_clean, component, namespace)
        helm_normalized = normalize_for_comparison(helm_clean, component, namespace)
        
        kustomize_json = json.dumps(kustomize_normalized, sort_keys=True)
        helm_json = json.dumps(helm_normalized, sort_keys=True)
        
        if kustomize_json != helm_json:
            differences.append(key)

    if differences:
        print(f"ERROR: Content differences in resources: {', '.join(sorted(differences))}")
        success = False

    if not success:
        print("FAILED: Manifests are not equivalent!")
        sys.exit(1)

if __name__ == "__main__":
    main()