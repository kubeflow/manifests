#!/bin/bash

namespace="kubeflow"
error_flag=0

# Function to check if 'id' command is available in a container
has_id_command() {
  local pod_name="$1"
  local container_name="$2"

  # Execute 'id' command and capture the output
  if kubectl exec -it -n "$namespace" "$pod_name" -c "$container_name" -- id -u >/dev/null 2>&1; then
    return 0  # 'id' command is available
  else
    return 1  # 'id' command is not available
  fi
}

# Function to check 'securityContext' and 'runAsNonRoot' at the pod or container level
has_securityContext_and_runAsNonRoot() {
  local pod_name="$1"
  local container_name="$2"

  # Use jq to check if 'securityContext' is defined at the pod level
  local securityContextPod=$(kubectl get pod -n "$namespace" "$pod_name" -o json | jq -r '.spec.securityContext')

  if [ "$securityContextPod" = "null" ]; then
    : # 'securityContext' is missing at the pod level, continue checking at the container level
  else
    # Check 'runAsNonRoot' at the pod level
    local runAsNonRootPod=$(kubectl get pod -n "$namespace" "$pod_name" -o json | jq -r '.spec.securityContext.runAsNonRoot // "Missing"')

    if [ "$runAsNonRootPod" = "Missing" ]; then
      : # 'runAsNonRoot' is missing at the pod level, continue checking at the container level
    else
      return 0  # 'runAsNonRoot' is present at the pod level (success)
    fi
  fi
  
  # Use jq to check 'securityContext' at the container level
  local securityContextContainer=$(kubectl get pod -n "$namespace" "$pod_name" -o json | jq -r '.spec.containers[] | select(.name == "'"$container_name"'").securityContext')

  if [ "$securityContextContainer" = "null" ]; then
    if [ "$securityContextPod" = "null" ]; then
      echo "Error: 'securityContext' is missing at the pod and container level in container $container_name of pod $pod_name"
      return 1
    else
      echo "Error: There is no runasnonroot on pod level and 'securityContext' is missing at container level in container $container_name of pod $pod_name"
      return 1
    fi
  fi

  # Check 'runAsNonRoot' at the container level
  local runAsNonRootContainer=$(kubectl get pod -n "$namespace" "$pod_name" -o json | jq -r '.spec.containers[] | select(.name == "'"$container_name"'").securityContext.runAsNonRoot // "Missing"')

  if [ "$runAsNonRootContainer" = "Missing" ]; then
    echo "Error: There is no runasnonroot on pod level and'runAsNonRoot' is missing in container $container_name of pod $pod_name"
    return 1  # 'runAsNonRoot' is missing at the container level (fail)
  fi

  return 0  # 'securityContext' and 'runAsNonRoot' are defined at the container level
}

# Get a list of pod names in the specified namespace that are not in the "Completed" state
pod_names=$(kubectl get pods -n "$namespace" --field-selector=status.phase!=Succeeded,status.phase!=Failed -o json | jq -r '.items[].metadata.name')

# Loop through the pod names and execute checks
for pod_name in $pod_names; do
  echo "Entering pod $pod_name in namespace $namespace..."

  container_names=$(kubectl get pod -n "$namespace" "$pod_name" -o json | jq -r '.spec.containers[].name')

  for container_name in $container_names; do
    if has_securityContext_and_runAsNonRoot "$pod_name" "$container_name"; then
      error_flag=1
    fi

    if has_id_command "$pod_name" "$container_name"; then
      user_id=$(kubectl exec -it -n "$namespace" "$pod_name" -c "$container_name" -- id -u)
      
      # Clean up whitespace in the user_id using tr
      user_id_cleaned=$(echo -n "$user_id" | tr -d '[:space:]')

      if [ "$user_id_cleaned" = "0" ]; then
        echo "Error: Pod $pod_name contains user ID 0 in container $container_name"
        error_flag=1
      else
        echo "Container: $container_name - User ID: $user_id_cleaned"
      fi
    else
      echo "Warning: 'id' command not available in container $container_name"
    fi
  done
done

# Exit with an error if any pod contains an error condition
if [ $error_flag -eq 1 ]; then
  exit 1
fi

# Exit successfully
exit 0
