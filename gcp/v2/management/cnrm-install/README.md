# Configuration for installing KCC in the management cluster.

Configs are a copy of the CNRM operator install (see [docs](https://cloud.google.com/config-connector/docs/how-to/advanced-install#manual))

To update:

1. Download the the latest GCS install bundle listed on (https://cloud.google.com/config-connector/docs/how-to/advanced-install#manual)

1. Copy the operator-system/configconnector-operator.yaml to `install-system`
1. Edit "0-cnrm-system.yaml" to add the kpt setter; change

     ```
  apiVersion: v1
  kind: ServiceAccount
  metadata:
    annotations:
      cnrm.cloud.google.com/version: 1.15.1
      iam.gke.io/gcp-service-account: cnrm-system@${PROJECT_ID?}.iam.gserviceaccount.com
    labels:
      cnrm.cloud.google.com/system: "true"
    name: cnrm-controller-manager
    namespace: cnrm-system
     ```

   to

     ```
     annotations:
      ...
      iam.gke.io/gcp-service-account: cnrm-system@${PROJECT_ID?}.iam.gserviceaccount.com # {"$kpt-set":"cnrm-system"}
     ```
