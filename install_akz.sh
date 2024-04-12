kustomize build example | awk '!/well-defined/' > kubeflow-akz.yaml

while ! kubectl apply -f kubeflow-akz.yaml; do echo "Retrying to apply resources"; sleep 120; done
# while ! kustomize build example | awk '!/well-defined/' | kubectl apply -f -; do echo "Retrying to apply resources"; sleep 10; done


#not working but still keeptrack!
go install github.com/arttor/helmify/cmd/helmify@latest
kustomize build example | awk '!/well-defined/' | ~/go/bin/helmify kbf

TIMEOUT=600s  # 10mins


echo "---"
echo "Waiting for all Kubeflow components to become ready."

echo "Waiting for Cert Manager pods to become ready..."
kubectl wait --timeout=${TIMEOUT} -n cert-manager --all --for=condition=Ready pod

echo "Waiting for istio-system Pods to become ready..."
kubectl wait --timeout=${TIMEOUT} -n istio-system --all --for=condition=Ready pod

echo "Waiting for knative-serving Pods to become ready..."
kubectl wait --timeout=${TIMEOUT} -n knative-serving --all --for=condition=Ready pod

echo "Waiting for kubeflow/ml-pipelines to become ready..."
kubectl wait --timeout=${TIMEOUT} -n kubeflow -l app=ml-pipeline --for=condition=Ready pod

echo "Waiting for kubeflow/kfserving to become ready..."
kubectl wait --timeout=${TIMEOUT} -n kubeflow -l app=kfserving --for=condition=Ready pod

echo "Waiting for kubeflow/katib to become ready..."
kubectl wait --timeout=${TIMEOUT} -n kubeflow -l katib.kubeflow.org/component=controller --for=condition=Ready pod

echo "Waiting for kubeflow/training-operator to become ready..."
kubectl wait --timeout=${TIMEOUT} -n kubeflow -l control-plane=kubeflow-training-operator --for=condition=Ready pod

# cat <<EOF | kubectl apply -f -
# apiVersion: kubeflow.org/v1alpha1
# kind: PodDefault
# metadata:
#   name: access-ml-pipeline
#   namespace: "thanhnm39315-sacombank-com"
# spec:
#   desc: Kubeflow Pipelines
#   selector:
#     matchLabels:
#       access-ml-pipeline: "true"
#   volumes:
#     - name: volume-kf-pipeline-token
#       projected:
#         sources:
#           - serviceAccountToken:
#               path: token
#               expirationSeconds: 99999
#               audience: pipelines.kubeflow.org      
#   volumeMounts:
#     - mountPath: /var/run/secrets/kubeflow/pipelines
#       name: volume-kf-pipeline-token
#       readOnly: true
#   env:
#     - name: KF_PIPELINES_SA_TOKEN_PATH
#       value: /var/run/secrets/kubeflow/pipelines/token
# EOF



# bash install_mlflow.sh
kubectl apply -f dex-configmap.yaml
kubectl delete pod `kubectl get pods -n auth | awk '{print $1}' | grep -iv name` -n auth 


# istioVersion="1.18.2"
# curl -L https://istio.io/downloadIstio | sh -
# mv istio-${istioVersion} /tmp/
# kubectl apply -f /tmp/istio-${istioVersion}/samples/addons
# kubectl rollout status deployment/kiali -n istio-system
# # /tmp/istio-1.18.2/bin/istioctl dashboard kiali
# cd ..

#kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80

#kubectl port-forward -n kubeflow svc/minio-service 9000:9000
#kubectl port-forward -n kubeflow svc/mlflowserver 5000:5000

kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80
# thanhnm39315@sacombank.com QweAsdZxc!23