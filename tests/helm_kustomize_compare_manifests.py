#!/usr/bin/env python3
import yaml
import sys
import json

def clean_helm_metadata(obj):
    """Remove ONLY helm-specific metadata from objects"""
    if isinstance(obj, dict):
        if 'metadata' in obj:
            metadata = obj['metadata']
            
            if 'labels' in metadata:
                labels = metadata['labels']
                helm_specific_labels = [
                    'helm.sh/chart', 
                    'app.kubernetes.io/managed-by',
                    'app.kubernetes.io/instance',
                    'app.kubernetes.io/version'
                ]
                for label in helm_specific_labels:
                    if label in labels:
                        del labels[label]
            
            if 'annotations' in metadata:
                annotations = metadata['annotations']
                helm_specific_annotations = [
                    'meta.helm.sh/release-name', 
                    'meta.helm.sh/release-namespace'
                ]
                for annotation in helm_specific_annotations:
                    if annotation in annotations:
                        del annotations[annotation]
        
        for key, value in list(obj.items()):
            if isinstance(value, (dict, list)):
                clean_helm_metadata(value)
    
    elif isinstance(obj, list):
        for item in obj:
            clean_helm_metadata(item)
    
    return obj

def json_diff(obj1, obj2, path=""):
    """Simple recursive comparison showing differences"""
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
                differences.append(f"{key_path}: missing in helm")
            else:
                differences.extend(json_diff(obj1[key], obj2[key], key_path))
    
    elif isinstance(obj1, list):
        if len(obj1) != len(obj2):
            differences.append(f"{path}: list length mismatch ({len(obj1)} vs {len(obj2)})")
        else:
            for i, (item1, item2) in enumerate(zip(obj1, obj2)):
                differences.extend(json_diff(item1, item2, f"{path}[{i}]"))
    
    elif obj1 != obj2:
        differences.append(f"{path}: '{obj1}' != '{obj2}'")
    
    return differences

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

    expected_extra_helm = set()
    if component == "spark-operator" and kubeflow_rbac_enabled:
        expected_extra_helm = {
            'ClusterRole/kubeflow-spark-admin',
            'ClusterRole/kubeflow-spark-edit', 
            'ClusterRole/kubeflow-spark-view'
        }

    extra_kustomize = kustomize_keys - helm_keys
    extra_helm = helm_keys - kustomize_keys
    unexpected_extra_helm = extra_helm - expected_extra_helm

    success = True
    
    if extra_kustomize:
        print(f"CRITICAL: Resources missing in Helm output: {', '.join(sorted(extra_kustomize))}")
        success = False

    if unexpected_extra_helm:
        print(f"WARNING: Unexpected extra resources in Helm: {', '.join(sorted(unexpected_extra_helm))}")
        
    print(f"\nCOMPARISON SUMMARY:")
    print(f"  Kustomize resources: {len(kustomize_keys)}")
    print(f"  Helm resources: {len(helm_keys)}")
    print(f"  Common resources: {len(kustomize_keys & helm_keys)}")
    print(f"  Resources only in Kustomize: {len(extra_kustomize)}")
    print(f"  Resources only in Helm: {len(extra_helm)}")
    print(f"  Expected extra Helm resources: {len(expected_extra_helm)}")

    common = kustomize_keys & helm_keys
    resources_with_differences = []
    
    for key in sorted(common):
        kustomize_doc = json.loads(json.dumps(kustomize[key]))
        helm_doc = json.loads(json.dumps(helm[key]))
        
        clean_helm_metadata(kustomize_doc)
        clean_helm_metadata(helm_doc)
        
        differences = json_diff(kustomize_doc, helm_doc)
        
        if differences:
            print(f"\nDifferences in {key}:")
            for diff in differences[:10]:  
                print(f"   {diff}")
            if len(differences) > 10:
                print(f"   ... and {len(differences) - 10} more differences")
            resources_with_differences.append(key)

    if not success or resources_with_differences:
        if resources_with_differences:
            print(f"\nFAILED: Found differences in {len(resources_with_differences)} resources:")
            for resource in resources_with_differences:
                print(f"   - {resource}")
        if extra_kustomize:
            print(f"\nFAILED: {len(extra_kustomize)} resources are missing in Helm output.")
            print("Helm templates need to be implemented for:")
            for resource in sorted(extra_kustomize):
                print(f"   - {resource}")
        print("\nREASON: Manifests are NOT equivalent!")
        sys.exit(1)
    else:
        print(f"\nSUCCESS: All {len(common)} common resources match!")
        print("Manifests are equivalent (ignoring only helm-specific metadata)!")
        sys.exit(0)

if __name__ == "__main__":
    main()