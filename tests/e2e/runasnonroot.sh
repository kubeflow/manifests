#!/bin/bash

namespace="kubeflow"
error_flag=0

# Get a list of pod names in the specified namespace
pod_names=$(kubectl get pods -n $namespace -o json | jq -r '.items[].metadata.name')
echo "Checking for root containers in namespace $namespace"

# Loop through the pod names and execute the 'id' command within each container
for pod_name in $pod_names; do
  echo "Entering pod $pod_name in namespace $namespace..."
  
  container_names=$(kubectl get pod -n $namespace $pod_name -o json | jq -r '.spec.containers[].name')

  for container_name in $container_names; do
    user_id=$(kubectl exec -it -n $namespace $pod_name -c $container_name -- id -u)
    # echo "Container: $container_name - User ID: $user_id"
    
    if [ "$user_id" -eq 0 ]; then
      echo "Error: Pod $pod_name contains user ID 0 in container $container_name"
      error_flag=1
    fi
  done

  echo "-------------------------------------"
done

# Exit with an error if any pod contains user ID 0
if [ $error_flag -eq 1 ]; then
  exit 1
fi

# Exit successfully if no pod contains user ID 0
exit 0