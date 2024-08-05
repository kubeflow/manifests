# SeaweedFS

- [Official documentation](https://github.com/seaweedfs/seaweedfs/wiki)
- [Official repository](https://github.com/seaweedfs/seaweedfs)

SeaweedFS is a simple and highly scalable distributed file system. It has an S3 interface which makes it usable as an object store for kubeflow.

## Prerequisites

- Kubernetes (any recent Version should work)
- You should have `kubectl` available and configured to talk to the desired cluster.
- `kustomize`.

## Compile manifests

```bash
kubectl kustomize ./base/
```

## Install SeaweedFS

**WARNING**
This replaces the service `minio-service` and will redirect the traffic to seaweedfs.

```bash
kubectl kustomize ./base/ | kubectl apply -f -
```

## Verify deployment

Run
```bash
./test.sh
```
With the ready check on the container it already verifies that the S3 starts correctly.
You can then use it with the endpoint at http://localhost:8333.
To activate authentication open a shell on the pod and use `weed shell` to configure your instance.
Create a user with the command `s3.configure -user <username> -access_key <access-key> -secret-key <secret-key> -actions Read:<my-bucket>/<my-prefix>,Write::<my-bucket>/<my-prefix>`
Documentation for this can also be found [here](https://github.com/seaweedfs/seaweedfs/wiki/Amazon-S3-API).

## Uninstall SeaweedFS

```bash
kubectl kustomize ./base/ | kubectl delete -f -
```
