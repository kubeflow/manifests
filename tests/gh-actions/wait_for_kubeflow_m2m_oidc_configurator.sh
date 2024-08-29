#!/bin/bash

CRONJOB_NAME=kubeflow-m2m-oidc-configurator
NAMESPACE=istio-system
RETRY_INTERVAL=5    # Each loop iterates after 'RETRY_INTERVAL' seconds
MAX_RETRIES=20    # Each loop iterates for a total number of 'MAX_RETRIES'

# Function to get the latest Job created by the CronJob
get_latest_job() {
  kubectl get jobs -n "${NAMESPACE}" \
    --sort-by=.metadata.creationTimestamp -o json \
    | jq --arg cronjob_name "${CRONJOB_NAME}" -r '.items[] | select(.metadata.ownerReferences[] | select(.name==$cronjob_name)) | .metadata.name' \
    | tail -n 1
}

# Wait until a Job is created
echo "Waiting for a Job to be created by the ${CRONJOB_NAME} CronJob..."
for ((i=1; i<=MAX_RETRIES; i++)); do
  JOB_NAME=$(get_latest_job)
  if [[ -n "${JOB_NAME}" ]]; then
    echo "Job ${JOB_NAME} created."
    break
  fi
  if [[ $i -eq $MAX_RETRIES ]]; then
    echo "Job creation timed out."
    exit 1
  fi
  sleep "${RETRY_INTERVAL}"
  echo "Waiting..."
done

# Wait for the Job to complete successfully
echo "Waiting for the Job ${JOB_NAME} to complete..."
for ((i=1; i<=MAX_RETRIES; i++)); do
  STATUS=$(kubectl get job "${JOB_NAME}" -n "${NAMESPACE}" -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}')
  if [[ "${STATUS}" == "True" ]]; then
    echo "Job ${JOB_NAME} completed successfully."
    break
  fi

  FAILED=$(kubectl get job "${JOB_NAME}" -n "${NAMESPACE}" -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}')
  if [[ "${FAILED}" == "True" ]]; then
    echo "Job ${JOB_NAME} failed."
    exit 1
  fi
  if [[ $i -eq $MAX_RETRIES ]]; then
    echo "Job completion timed out."
    exit 1
  fi
  sleep "${RETRY_INTERVAL}"
done
