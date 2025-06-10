#!/usr/bin/env python3
import yaml
import sys
import os
from pathlib import Path

def patch_yaml_file(file_path, component):
    """Patch a YAML file to add conditional rendering and namespace"""
    with open(file_path, 'r') as f:
        content = f.read()
    
    component_map = {
        'spark-operator': 'sparkOperator.enabled',
        'cert-manager': 'certManager.enabled',
    }
    
    condition = component_map.get(component, f'{component}.enabled')
    condition_check = f'{{{{- if .Values.{condition} }}}}'
    
    if condition_check in content:
        return
    
    try:
        docs = list(yaml.safe_load_all(content))
        patched_docs = []
        
        for doc in docs:
            if doc and isinstance(doc, dict):
                if 'metadata' in doc and doc.get('kind') in [
                    'Deployment', 'Service', 'ServiceAccount', 'Role', 'RoleBinding'
                ]:
                    if not isinstance(doc['metadata'].get('namespace'), str) or '{{' not in doc['metadata'].get('namespace', ''):
                        if component == 'cert-manager':
                            if doc.get('kind') in ['Role', 'RoleBinding'] and 'leaderelection' in doc['metadata'].get('name', ''):
                                doc['metadata']['namespace'] = 'kube-system'
                            else:
                                doc['metadata']['namespace'] = '{{ .Values.global.certManagerNamespace }}'
                        else:
                            doc['metadata']['namespace'] = '{{ include "kubeflow.namespace" . }}'
                
                if doc.get('kind') == 'Deployment' and 'spec' in doc and component == 'spark-operator':
                    if 'template' in doc['spec'] and 'metadata' in doc['spec']['template']:
                        template_meta = doc['spec']['template']['metadata']
                        if 'labels' not in template_meta:
                            template_meta['labels'] = {}
                        template_meta['labels']['sidecar.istio.io/inject'] = 'false'
                
                patched_docs.append(doc)
        
        with open(file_path, 'w') as f:
            f.write(f'{condition_check}\n')
            for doc in patched_docs:
                f.write('---\n')
                yaml.dump(doc, f, default_flow_style=False, sort_keys=False)
            f.write('{{- end }}\n')
    
    except Exception as e:
        print(f"Warning: Could not patch {file_path}: {e}")
        with open(file_path, 'w') as f:
            f.write(f'{condition_check}\n')
            f.write(content)
            f.write('{{- end }}\n')

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: patch-templates.py <templates_dir> <component>")
        sys.exit(1)
    
    templates_dir = sys.argv[1]
    component = sys.argv[2]
    for yaml_file in Path(templates_dir).rglob("*.yaml"):
        patch_yaml_file(str(yaml_file), component) 