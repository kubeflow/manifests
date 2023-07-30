#!/bin/sh

set -e

if [ -z "${AWS_BUCKET}" ]; then
  echo >&2 "AWS_BUCKET must be set"
  exit 1
fi

if [ -z "${AWS_ACCESS_KEY_ID}" ]; then
  echo >&2 "AWS_ACCESS_KEY_ID must be set"
  exit 1
fi

if [ -z "${AWS_SECRET_ACCESS_KEY}" ]; then
  echo >&2 "AWS_SECRET_ACCESS_KEY must be set"
  exit 1
fi

mkdir -p "${FILE_DIR}"

# mlflow server \
#     --backend-store-uri "file://$FILE_DIR" \
#     --default-artifact-root "s3://$AWS_BUCKET/mlflow/artifacts" \
#     --host 0.0.0.0 \
#     --port "$PORT"



#cannot run with mysql? strange? so workaround working with postgresql then
mlflow server \
    --backend-store-uri "mysql://mlflow:mlflow@mysql.kubeflow.svc:3306/mlflowdb" \
    --default-artifact-root "s3://$AWS_BUCKET/mlflow/artifacts" \
    --host 0.0.0.0 \
    --port "$PORT"


