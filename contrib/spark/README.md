# Kubeflow Spark Operator

[![Integration Test](https://github.com/kubeflow/spark-operator/actions/workflows/integration.yaml/badge.svg)](https://github.com/kubeflow/spark-operator/actions/workflows/integration.yaml)[![Go Report Card](https://goreportcard.com/badge/github.com/kubeflow/spark-operator)](https://goreportcard.com/report/github.com/kubeflow/spark-operator)

## What is Spark Operator?

The Kubernetes Operator for Apache Spark aims to make specifying and running [Spark](https://github.com/apache/spark) applications as easy and idiomatic as running other workloads on Kubernetes. It uses
[Kubernetes custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) for specifying, running, and surfacing status of Spark applications.

## Overview

For a complete reference of the custom resource definitions, please refer to the [API Definition](docs/api-docs.md). For details on its design, please refer to the [Architecture](https://www.kubeflow.org/docs/components/spark-operator/overview/#architecture). It requires Spark 2.3 and above that supports Kubernetes as a native scheduler backend.

The Kubernetes Operator for Apache Spark currently supports the following list of features:

* Supports Spark 2.3 and up.
* Enables declarative application specification and management of applications through custom resources.
* Automatically runs `spark-submit` on behalf of users for each `SparkApplication` eligible for submission.
* Provides native [cron](https://en.wikipedia.org/wiki/Cron) support for running scheduled applications.
* Supports customization of Spark pods beyond what Spark natively is able to do through the mutating admission webhook, e.g., mounting ConfigMaps and volumes, and setting pod affinity/anti-affinity.
* Supports automatic application re-submission for updated `SparkApplication` objects with updated specification.
* Supports automatic application restart with a configurable restart policy.
* Supports automatic retries of failed submissions with optional linear back-off.
* Supports mounting local Hadoop configuration as a Kubernetes ConfigMap automatically via `sparkctl`.
* Supports automatically staging local application dependencies to Google Cloud Storage (GCS) via `sparkctl`.
* Supports collecting and exporting application-level metrics and driver/executor metrics to Prometheus.
