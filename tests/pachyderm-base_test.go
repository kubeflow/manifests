package tests_test

import (
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
	"testing"
)

func writePachydermBase(th *KustTestHarness) {
	th.writeF("/manifests/pachyderm/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: service-account`)
	th.writeF("/manifests/pachyderm/base/etcd-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: etcd
  name: etcd
spec:
  ports:
    - name: client-port
      port: 2379
      protocol: TCP
      targetPort: 2379
  selector:
    app: etcd
  type: NodePort`)
	th.writeF("/manifests/pachyderm/base/etcd-headless-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: etcd
  name: etcd-headless
spec:
  clusterIP: None
  sessionAffinity: None
  ports:
    - name: peer-port
      port: 2380
      protocol: TCP
      targetPort: 2380
  selector:
    app: etcd`)
	th.writeF("/manifests/pachyderm/base/etcd-stateful-set.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    app: etcd
  name: etcd
spec:
  replicas: 3
  podManagementPolicy: OrderedReady
  revisionHistoryLimit: 10
  serviceName: $(etcdHeadlessServiceName)
  selector:
    matchLabels:
      app: etcd
  template:
    metadata:
      labels:
        app: etcd
    spec:
      containers:
      - command:
          - "/bin/sh"
          - "-ec"
          - |
            HOSTNAME=$(hostname)
            echo "etcd api version is ${ETCDAPI_VERSION}"

            # store member id into PVC for later member replacement
            collect_member() {
                while ! etcdctl member list &>/dev/null; do sleep 1; done
                etcdctl member list | grep http://${HOSTNAME}.${SERVICE_NAME}:2380 | cut -d':' -f1 | cut -d'[' -f1 > /var/run/etcd/member_id
                exit 0
            }

            eps() {
                EPS=""
                for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                    EPS="${EPS}${EPS:+,}http://${CLUSTER_NAME}-${i}.${SERVICE_NAME}:2379"
                done
                echo ${EPS}
            }

            member_hash() {
                etcdctl member list | grep http://${HOSTNAME}.${SERVICE_NAME}:2380 | cut -d':' -f1 | cut -d'[' -f1
            }

            initial_peers() {
                PEERS=""
                for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                  PEERS="${PEERS}${PEERS:+,}${CLUSTER_NAME}-${i}=http://${CLUSTER_NAME}-${i}.${SERVICE_NAME}:2380"
                done
                echo ${PEERS}
            }

            # re-joining after failure?
            if [ -e /var/run/etcd/default.etcd ]; then
                echo "Re-joining etcd member"
                member_id=$(cat /var/run/etcd/member_id)

                # re-join member
                ETCDCTL_ENDPOINT=$(eps) etcdctl member update ${member_id} http://${HOSTNAME}.${SERVICE_NAME}:2380
                exec etcd --name ${HOSTNAME} \
                    --listen-peer-urls http://${HOSTNAME}.${SERVICE_NAME}:2380 \
                    --listen-client-urls http://${HOSTNAME}.${SERVICE_NAME}:2379,http://127.0.0.1:2379 \
                    --advertise-client-urls http://${HOSTNAME}.${SERVICE_NAME}:2379 \
                    --data-dir /var/run/etcd/default.etcd
            fi

            # etcd-SET_ID
            # SET_ID="${HOSTNAME: -1}"
            SET_ID=${HOSTNAME##*-}

            # adding a new member to existing cluster (assuming all initial pods are available)
            if [ "${SET_ID}" -ge ${INITIAL_CLUSTER_SIZE} ]; then
                export ETCDCTL_ENDPOINT=$(eps)

                # member already added?
                MEMBER_HASH=$(member_hash)
                if [ -n "${MEMBER_HASH}" ]; then
                    # the member hash exists but for some reason etcd failed
                    # as the datadir has not be created, we can remove the member
                    # and retrieve new hash
                    # etcdctl member remove ${MEMBER_HASH}
                    if [ "${ETCDAPI_VERSION}" -eq 3 ]; then
                        ETCDCTL_API=3 etcdctl --user=root:${ROOT_PASSWORD} member remove ${MEMBER_HASH}
                    else
                        etcdctl --username=root:${ROOT_PASSWORD} member remove ${MEMBER_HASH}
                    fi
                fi

                echo "Adding new member"
                etcdctl member add ${HOSTNAME} http://${HOSTNAME}.${SERVICE_NAME}:2380 | grep "^ETCD_" > /var/run/etcd/new_member_envs

                if [ $? -ne 0 ]; then
                    echo "Exiting"
                    rm -f /var/run/etcd/new_member_envs
                    exit 1
                fi

                cat /var/run/etcd/new_member_envs
                source /var/run/etcd/new_member_envs

                collect_member &

                exec etcd --name ${HOSTNAME} \
                    --listen-peer-urls http://0.0.0.0:2380 \
                    --listen-client-urls http://0.0.0.0:2379 \
                    --advertise-client-urls http://${HOSTNAME}.${SERVICE_NAME}:2379 \
                    --data-dir /var/run/etcd/default.etcd \
                    --initial-advertise-peer-urls http://${HOSTNAME}.${SERVICE_NAME}:2380 \
                    --initial-cluster ${ETCD_INITIAL_CLUSTER} \
                    --initial-cluster-state ${ETCD_INITIAL_CLUSTER_STATE}
            fi

            for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                while true; do
                    echo "Waiting for ${CLUSTER_NAME}-${i}.${SERVICE_NAME} to come up"
                    ping -W 1 -c 1 ${CLUSTER_NAME}-${i}.${SERVICE_NAME} > /dev/null && break
                    sleep 1s
                done
            done

            PEERS=""
            for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                PEERS="${PEERS}${PEERS:+,}${CLUSTER_NAME}-${i}=http://${CLUSTER_NAME}-${i}.${SERVICE_NAME}:2380"
            done

            collect_member &

            # join member
            exec etcd --name ${HOSTNAME} \
                --initial-advertise-peer-urls http://${HOSTNAME}.${SERVICE_NAME}:2380 \
                --listen-peer-urls http://0.0.0.0:2380 \
                --listen-client-urls http://0.0.0.0:2379 \
                --advertise-client-urls http://${HOSTNAME}.${SERVICE_NAME}:2379 \
                --initial-cluster-token etcd-cluster-1 \
                --initial-cluster $(initial_peers) \
                --initial-cluster-state new \
                --data-dir /var/run/etcd/default.etcd
        image: pachyderm/etcd:v3.2.7
        imagePullPolicy: IfNotPresent
        name: etcd
        env:
          - name: INITIAL_CLUSTER_SIZE
            value: "3"
          - name: SERVICE_NAME
            value: $(etcdHeadlessServiceName)
          - name: CLUSTER_NAME
            value: pachyderm-etcd
        lifecycle:
          preStop:
            exec:
              command:
                - "/bin/sh"
                - "-ec"
                - |
                  EPS=""
                  for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
                      EPS="${EPS}${EPS:+,}http://${CLUSTER_NAME}-${i}.${SERVICE_NAME}:2379"
                  done

                  HOSTNAME=$(hostname)

                  member_hash() {
                      etcdctl member list | grep http://${HOSTNAME}.${SERVICE_NAME}:2380 | cut -d':' -f1 | cut -d'[' -f1
                  }

                  echo "Removing ${HOSTNAME} from etcd cluster"

                  ETCDCTL_ENDPOINT=${EPS} etcdctl member remove $(member_hash)
                  if [ $? -eq 0 ]; then
                    # Remove everything otherwise the cluster will no longer scale-up
                    rm -rf /var/run/etcd/*
        ports:
        - name: client-port
          containerPort: 2379
        - name: peer-port
          containerPort: 2380
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/data/etcd
          name: etcdvol
      terminationGracePeriodSeconds: 30
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      volumes:
      - name: etcdvol
        hostPath:
          path: /var/pachyderm/etcd
  updateStrategy:
    type: OnDelete
  volumeClaimTemplates: []
`)
	th.writeF("/manifests/pachyderm/base/pachd-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: pachd
  name: service
spec:
  ports:
    - name: api-grpc-port
      nodePort: 30650
      port: 650
      protocol: TCP
      targetPort: 650
    - name: trace-port
      nodePort: 30651
      port: 651
      protocol: TCP
      targetPort: 651
    - name: api-http-port
      nodePort: 30652
      port: 652
      protocol: TCP
      targetPort: 652
    - name: api-git-port
      nodePort: 30999
      port: 999
      protocol: TCP
      targetPort: 999
  selector:
    app: pachd
  type: NodePort`)
	th.writeF("/manifests/pachyderm/base/pachd-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  replicas: 1
  strategy: {}
  selector:
    matchLabels:
      app: pachd
  template:
    metadata:
      labels:
        app: pachd
    spec:
      containers:
      - name: myapp
        image: pachyderm/pachd:1.8.6
        imagePullPolicy: Always
        envFrom:
        - configMapRef:
            name: parameters
        - secretRef:
            name: secrets
        env:
        - name: ETCD_SERVICE_HOST
          value: $(etcdServiceName)
        - name: PACHD_POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        resources:
          limits:
            memory: 512Mi
            cpu: 250m
        ports:
        - name: api-grpc-port
          containerPort: 650
          protocol: TCP
        - name: trace-port
          containerPort: 651
          protocol: TCP
        - name: api-http-port
          containerPort: 652
          protocol: TCP
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /pach
          name: pachdvol
      serviceAccountName: pachyderm-service-account
      volumes:
      - name: pachdvol
        emptyDir: {}

`)
	th.writeF("/manifests/pachyderm/base/params.yaml", `
varReference:
- path: spec/template/spec/containers/env/value
  kind: Deployment
- path: spec/serviceName
  kind: StatefulSet
- path: spec/template/spec/containers/env/value
  kind: StatefulSet`)
	th.writeF("/manifests/pachyderm/base/params.env", `
PACH_ROOT=/pach
NUM_SHARDS=16
EXPORT_OBJECT_API=false
WORKER_IMAGE=pachyderm/worker:1.8.6
WORKER_SIDERCAR_IMAGE=pachyderm/pachd:1.8.6
WORKER_IMAGE_PULL_POLICY=IfNotPresent
PACHD_VERSION=1.8.6
METRICS=true
LOG_LEVEL=info
BLOCK_CACHE_BYTES=0G
PACHYDERM_AUTHENTICATION_DISABLED_FOR_TESTING=false
ETCD_SERVICE_PORT=2379
`)
	th.writeF("/manifests/pachyderm/base/secrets.env", `
`)
	th.writeK("/manifests/pachyderm/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service-account.yaml
- etcd-service.yaml
- etcd-headless-service.yaml
- etcd-stateful-set.yaml
- pachd-service.yaml
- pachd-deployment.yaml
namespace: kubeflow
namePrefix: pachyderm-
commonLabels:
  kustomize.component: pachyderm
configMapGenerator:
- name: parameters
  env: params.env
secretGenerator:
- name: secrets
  env: secrets.env
vars:
- name: etcdHeadlessServiceName
  objref:
    kind: Service
    name: etcd-headless
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: etcdServiceName
  objref:
    kind: Service
    name: etcd
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: pachydermServiceAccountName
  objref:
    kind: ServiceAccount
    name: service-account
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: pachydermServiceAccountNamespace
  objref:
    kind: ServiceAccount
    name: service-account
    apiVersion: v1
  fieldref:
    fieldpath: metadata.namespace
configurations:
- params.yaml
`)
}

func TestPachydermBase(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/pachyderm/base")
	writePachydermBase(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../pachyderm/base"
	fsys := fs.MakeRealFS()
	lrc := loader.RestrictionRootOnly
	_loader, loaderErr := loader.NewLoader(lrc, validators.MakeFakeValidator(), targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pc := plugins.DefaultPluginConfig()
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl(), plugins.NewLoader(pc, rf))
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
