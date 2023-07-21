for i in `kubectl get ns | awk '{print $1}' | grep -ivE "name|kube-node|kube-system|kube-public"`
do 
echo $i
# kubectl scale deploy -n $i --replicas=0 --all 
# kubectl scale statefulset -n $i --replicas=0 --all 
kubectl delete ns $i &

done

istioVersion="1.18.1"

curl -L https://istio.io/downloadIstio | sh -
mv istio-${istioVersion} /tmp/
kubectl delete -f /tmp/istio-${istioVersion}/samples/addons
