#!/bin/bash

if [ -s changed_files_in_PR.txt ]; then
  while IFS= read -r file; do
    echo "Running yamllint on $file"
    yamllint "$file"
  done < changed_files_in_PR.txt
else
  echo "No YAML files changed in this PR."
fi
