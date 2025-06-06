#!/usr/bin/env python3
import yaml
import sys
import os
from pathlib import Path

def patch_yaml_file(file_path):
    """Patch a YAML file to add namespace and Istio labels"""
    with open(file_path, 'r') as f:
        content = f.read()
    
    if '{{- if .Values.sparkOperator.enabled }}' in content:
        return
    
    try:
        docs = list(yaml.safe_load_all(content))
        patched_docs = []
        
        for doc in docs:
            if doc and isinstance(doc, dict):
                if 'metadata' in doc and doc.get('kind') in [
                    'Deployment', 'Service', 'ServiceAccount', 'Role', 'RoleBinding'
                ]:
                    doc['metadata']['namespace'] = '{{ include "kubeflow.namespace" . }}'
                
                if doc.get('kind') == 'Deployment' and 'spec' in doc:
                    if 'template' in doc['spec'] and 'metadata' in doc['spec']['template']:
                        template_meta = doc['spec']['template']['metadata']
                        if 'labels' not in template_meta:
                            template_meta['labels'] = {}
                        template_meta['labels']['sidecar.istio.io/inject'] = 'false'
                
                patched_docs.append(doc)
        
        with open(file_path, 'w') as f:
            f.write('{{- if .Values.sparkOperator.enabled }}\n')
            for doc in patched_docs:
                f.write('---\n')
                yaml.dump(doc, f, default_flow_style=False, sort_keys=False)
            f.write('{{- end }}\n')
    
    except Exception as e:
        print(f"Warning: Could not patch {file_path}: {e}")
        with open(file_path, 'w') as f:
            f.write('{{- if .Values.sparkOperator.enabled }}\n')
            f.write(content)
            f.write('{{- end }}\n')

if __name__ == "__main__":
    templates_dir = sys.argv[1]
    for yaml_file in Path(templates_dir).rglob("*.yaml"):
        patch_yaml_file(str(yaml_file)) 