# Example with external databases

## Overview

In this example all components necessary for kubeflow to operate will be
created inside the kubernetes cluster of your choosing, except for the `katib`
and `pipelines` MySQL databases

## How to use this example:

### 1. Create your own `kustomization.yaml`

Your `kustomization.yaml` should target a `branch`, `tag` or `release` and
you must specify the credentials and host(s) for the external database(s)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - https://github.com/kubeflow/manifests/examples/external-dbs?ref=v1.4.1
configMapGenerator:
  - envs:
    - pipeline-mysql.env
    name: pipeline-install-config
    namespace: kubeflow
    behavior: merge
secretGenerator:
  - name: katib-mysql-secrets
    behavior: replace
    envs:
      - secrets-envs/katib-mysql.env
  - name: mysql-secret
    behavior: replace
    envs:
      - secrets-envs/pipelines-mysql.env
```

### 2. Create files with database credentials and hosts

File structure would look something like this

```shell
your-kustomize-folder
├── kustomization.yaml
├── pipeline-mysql.env
└── secrets-envs
    ├── katib-mysql.env
    └── pipelines-mysql.env
```

#### `pipeline-mysql.env` contents

```shell
dbHost=MY-PIPELINE-MYSQL-HOST
```

#### `secrets-envs/katib-mysql.env` contents

```shell
DB_USER=MY-KATIB-MYSQL-DB-USERNAME
DB_PASSWORD=MY-KATIB-MYSQL-DB-PASSWORD
KATIB_MYSQL_DB_HOST=MY-KATIB-MYSQL-DB-HOST
KATIB_MYSQL_DB_PORT=MY-KATIB-MYSQL-PORT
```

#### `secrets-envs/pipelines-mysql.env` contents

```shell
username=MY-PIPELINE-MYSQL-DB-USERNAME
password=MY-PIPELINE-MYSQL-DB-PASSWORD
```

### 3. Apply resources

```shell
while ! kustomize build your-kustomize-folder | kubectl apply -f -; do echo "Retrying to apply resources"; sleep 10; done
```
