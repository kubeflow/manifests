package tests_test

import (
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/target"
	"testing"
)

func writeE2eTasksOverlaysE2e(th *KustTestHarness) {
	th.writeF("/manifests/e2e/e2e-tasks/overlays/e2e/task.yaml", `
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: kfctl-e2e
spec:
  inputs:
    resources:
    - name: image
      type: image
    params:
    - name: image
      type: string
      description: the image name
    - name: project
      type: string
      description: the gcp project to run the e2e tests
    - name: cluster
      type: string
      description: name of k8 cluster
    - name: bucket
      type: string
      description: name of gcp bucket to store test results
    - name: repos_dir
      type: string
      description: path to where the repo is downloaded
    - name: zone
      type: string
      description: k8 zone where e2e tests run
    - name: configPath
      type: string
      description: the location of the config file for kfctl init
    - name: email
      type: string
      description: email of project owner
    - name: platform
      type: string
      description: all | k8s
    - name: REPO_OWNER
      type: string
      description: git repository org
    - name: REPO_NAME
      type: string
      description: git repository name
    - name: config_file
      type: string
      description: the location of the prow_config file
    - name: JOB_NAME
      type: string
      description: prow job name
    - name: JOB_TYPE
      type: string
      description: presubmit | postsubmit | periodic
    - name: PULL_NUMBER
      type: string
      description: PR #
    - name: PULL_BASE_REF
      type: string
      description: master | pull id
    - name: PULL_PULL_SHA
      type: string
      description: sha of pull id
    - name: BUILD_NUMBER
      type: string
      description: build #
    image: "${image}"
    command: ["python"]
    args:
    - "-m"
    - "kubeflow.testing.run_tests"
    - "--project"
    - "${inputs.params.project}"
    - "--zone"
    - "${inputs.params.zone}"
    - "--cluster"
    - "${inputs.params.cluster}"
    - "--bucket"
    - "${inputs.params.bucket}"
    - "--config_file"
    - "${inputs.params.config_file}"
    - "--repos_dir"
    - "${inputs.params.repos_dir}"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_SECRET
    - name: JOB_NAME
      value: ${inputs.params.JOB_NAME}
    - name: JOB_TYPE
      value: ${inputs.params.JOB_TYPE}
    - name: PULL_NUMBER
      value: "${inputs.params.PULL_NUMBER}"
    - name: REPO_NAME
      value: "${inputs.params.REPO_NAME}"
    - name: REPO_OWNER
      value: "${inputs.params.REPO_OWNER}"
    - name: BUILD_NUMBER
      value: "${inputs.params.BUILD_NUMBER}"
    volumeMounts:
    - name: kaniko-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  volumes:
  - name: kaniko-secret
    secret:
      secretName: kaniko-secret
  - name: kubeflow
    persistentVolumeClaim:
      claimName: kubeflow-pvc
`)
	th.writeK("/manifests/e2e/e2e-tasks/overlays/e2e", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- task.yaml
`)
	th.writeF("/manifests/e2e/e2e-tasks/base/persistent-volume-claim.yaml", `
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: kubeflow-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Gi
`)
	th.writeF("/manifests/e2e/e2e-tasks/base/secret.yaml", `
---
apiVersion: v1
kind: Secret
metadata:
  name: kaniko-secret
type: Opaque
data:
  kaniko-secret.json: ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsCiAgInByb2plY3RfaWQiOiAiY29uc3RhbnQtY3ViaXN0LTE3MzEyMyIsCiAgInByaXZhdGVfa2V5X2lkIjogImQxMjkzODQ1NWI2NGEyZDlhZWE1MDVjNjZkNzIyMjJmNmUyNDg0MzYiLAogICJwcml2YXRlX2tleSI6ICItLS0tLUJFR0lOIFBSSVZBVEUgS0VZLS0tLS1cbk1JSUV2UUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktjd2dnU2pBZ0VBQW9JQkFRQ2dJRHptZmFqMkdKSnVcbjlvOUxsa3dDYXIvVXYyalN3MEJTRXpNeWozQlNCc3BSdXVhOG1qU3MyZTdNRjVnYmMxVGdoL21HSkNqMitZS3JcbkZDOHdyMEt5bWx6NDdRWE5aeUMxb2I2TitxUnJncEJyeXNZWE1iSkFnQUFQck1rS2lhY2VjRzBHN3JkVGhjQVNcblh2Y2hBZG9RdTFTTWhNQys4aEtHMytpM1ZETXE2K3o4ZStwL1AzeEhhdUl3UTBYcWp4RGpONE5VRlMrbTF3WG9cbmpQM2xkbE15cWJUZHFvdFNHanIrTUxJZ1U3OEowdlp1NDF5WDl1ZjY2QWJyOE91N2E4bFdpVnpJUzZubHVqSTJcbm00T3NWYXlZaUN6dHhOam5iazVCVUhaYXZpekNaLzE1YzRlWFRuYXlsQXI0S2IrS3ZHSk5XbXVjSUdjQlNkZnlcbjhGZE1aNEdSQWdNQkFBRUNnZ0VBQWhXaFdpQ1Y4d3d4dmMvMHh1UmdQVzgzQlM4OGdveXVjcmw1OWpIOGgvd0lcbkpSN3VHWWJ3bE9nUXg3Ukg3UldPVWx0YXkzYVl2b1h0ZitPOW9JYXJqTnVRclBubUlUQW1QMDhQMzRpV3RnOWdcbjZRV0UvdkROSU10VFlhMm9NdStyVHRMNzhzTG5zQ2ExMWxkaTE0QUNPS3R3Ymk0UWIwaTJ3Q0VJK1Y1a0ZsVElcbmlkZkFiWGdiakt6dDR0dFg5YmR3OVBpUEcrTkRxQkVpQ0FiaXNMYmxyWlVOWUl1WVVsVWpoTmNKS29hRWNXQkdcbkMzcS9BMk1oTElTZ2FVQlYyZnludXEyaVdVZFZQTGwxNTRNNWtMREFSNk8rS0J6UWVZVnUraGxzdVcrSE1qcGFcbjI5U05FSWhuWmYySERPTi82bUw3R1VEcVlOblhISDVCK1AvT3NnSVQ5UUtCZ1FEV3R4SWNXSks3NFdadmhYeGNcbktOR1pLdHlxQWxielhSVFI3M1NGTVpSQjZNUjdEOFloenIwN3dmRXI5UEk5ekl3QjB3V3BzR2VGRVFQUFpTYXlcbnlKaDQxcTkvYlZ2bUtiM2svRndrMmVFSy9JRFBNRzdqTHdqbVBBbG5UYVpVRUZncHVKYUw5VEVvYlBSWXZtbGVcbmpyeXNmNVRpWXBIY3RkVlhkL3IwZzRqbDNRS0JnUUMrNmg4a28yNVkrS3BPRitxaGU3UnA1Z1VZdFhwQjh6RnpcbjB1UnZacE5DQXFMUmw3Vjk5RW9YcnBZWFF5TE5CY3BSUkpDaEpVenZDMkRtWVhhRXhLUzlWbmZYSzRncG5uVzZcbnAzSitPQ0NtOUJOdUMrYStxYXp0WDluK1JrRXpaem0waFdjTFRTejVpUlhrOWkxTkRKZnY4aXBKd1ptMnhPNDBcbkNXdU5pZ2R4UlFLQmdEbllhRkNxckIxaHhDOFhUMEdrM1pMZU1VUzhES0RUMnVBVUd0Z25XMEhHYStpYmYwMXNcblhSN1VTUjBHaUp5TmxzcUhCMmVIMXR2S2tiUTJGQTdtYSsxaUtUV3pTS2JoWi85ZzNaSXdBS2p0RGViRHJad1dcbjk5YlBKZGxtMmdDYnhxUzJ6aGcybmwrOXVyYU4xZVZibndqNTlpcG5VOVNhU0RlZ1kwT3NqQjBoQW9HQkFKUzRcbnN5d0tlRktzMjVaY1FUWXN0TDF1SjNnNUh4VXpDc29NZGxGbDJiOHBhSWJYcE5XS3NSRkR1cjVDV1dEWGF1VG1cbkFialc0dGl3eDNxUVlCQkxVMzMvVnZueWVtN1pkeUxCZ0lwYzFPclo1aXpxN29TR2p5U1hiNjBLTTQ2RWtrcFRcblJaTmpPbTdsWUgzdFhCclNmYVc0dzBLVG8xZmlqeUZRV1UxNFFoWDFBb0dBSnJrcHg5TjE5UExnQ3MrWDhKa1ZcblRRZXppeHpUd0VLekdRQVU1SnJVai9Cd1g5RDBMa2tkSVdXWXBTMzZqdnpqQUM2Nk93SHBOQlF2dXRRRC9CU3FcbjZUd0ovbVhQN1p0U1hsUFd3MGExVkhNNG5oTmcrbDRXR3BHNCtmdTduZUM5bHlJZkZ2am1jZHM1d0RNNXRDOVFcbjcwOUlYSDEzamdXbzlNYzQyTExkdUxFPVxuLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLVxuIiwKICAiY2xpZW50X2VtYWlsIjogImRvY2tlckBjb25zdGFudC1jdWJpc3QtMTczMTIzLmlhbS5nc2VydmljZWFjY291bnQuY29tIiwKICAiY2xpZW50X2lkIjogIjExMzA4NDc2Nzg5NTE3MzQ1MTIwOCIsCiAgImF1dGhfdXJpIjogImh0dHBzOi8vYWNjb3VudHMuZ29vZ2xlLmNvbS9vL29hdXRoMi9hdXRoIiwKICAidG9rZW5fdXJpIjogImh0dHBzOi8vb2F1dGgyLmdvb2dsZWFwaXMuY29tL3Rva2VuIiwKICAiYXV0aF9wcm92aWRlcl94NTA5X2NlcnRfdXJsIjogImh0dHBzOi8vd3d3Lmdvb2dsZWFwaXMuY29tL29hdXRoMi92MS9jZXJ0cyIsCiAgImNsaWVudF94NTA5X2NlcnRfdXJsIjogImh0dHBzOi8vd3d3Lmdvb2dsZWFwaXMuY29tL3JvYm90L3YxL21ldGFkYXRhL3g1MDkvZG9ja2VyJTQwY29uc3RhbnQtY3ViaXN0LTE3MzEyMy5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbSIKfQo=
---
apiVersion: v1
kind: Secret
metadata:
  name: docker-secret
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: eyJhdXRocyI6eyJodHRwczovL2djci5pbyI6eyJ1c2VybmFtZSI6Il9qc29uX2tleSIsInBhc3N3b3JkIjoie1xuICBcInR5cGVcIjogXCJzZXJ2aWNlX2FjY291bnRcIixcbiAgXCJwcm9qZWN0X2lkXCI6IFwiY29uc3RhbnQtY3ViaXN0LTE3MzEyM1wiLFxuICBcInByaXZhdGVfa2V5X2lkXCI6IFwiZTMzZDhhZDQ4OWZkYTEzMTg0ZmQxZDYzZmVjMDhjY2RhZTlkZTE5MFwiLFxuICBcInByaXZhdGVfa2V5XCI6IFwiLS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tXFxuTUlJRXZnSUJBREFOQmdrcWhraUc5dzBCQVFFRkFBU0NCS2d3Z2dTa0FnRUFBb0lCQVFDV0xaWW5iU3BuZk1Uc1xcblRsSktZU2lUNWxvd1pveWRPZndMNEQxM0dCUnlneml2aWpDOUtValE4a2t2Y05NeEJweHRSUUtNYVd2UjNoaHZcXG5XejE4eC9xZW1DdlhuV1FMMngxNzJzeTVwcTRsVnlaUzJKbi9jNHhJQVNGNHhXb3lvQVpsNXpFZk8wem01M3BkXFxuNCtqd0NMcDdDRSs1UVJTclBjdXhIdkNvZUhKeXhxWGtmOGJoTVpXN1BJQUVNMWlZSEdNb01ua1VOdVFiU0VvTlxcbkw0SHJTcUgwZnV5RkVLNjFRYUJKR3RSdloyZVZTUmhkUEFaRTFEQW5OWFVldXI0NTE3bkhXQTl5WVRTQVdnRTZcXG41U01TSEVxWVVMSmdMMzd2SktiUUFwSjAwZ25nNnBNTFJEanB0aFdBa0lBL3BYamxNR2lmMnZsdnY5REVRN3p1XFxuSFRwU2dEZWJBZ01CQUFFQ2dnRUFTTTN0MHN4SjkrU1ZaUWZ0UGZEUEpyQlFQZEdoVHFHckxxaTFzNVJCYVdoelxcbkpTcWx5VGFIL2YvUGVnZkU0cW9WVUtYWmYrK2xuUmNDR280TmgzNDlZZ0JjbE1sUkZLeFRwVlVqMWNiWCt2TStcXG5lWUJYVysrTTdPVmJjRHlvYU1XS2hJRnBuMzMwb0taTWZOTDkvTXdHZDVuR2FJV0QreVpZcHRQY2tKZmZ5QU1HXFxuT1dBTTd6cXdrRHg3RDV4Umg2T0Nzb1kydlNwUGh2cmJzbUxOdTVlYm5TWUxPWFJlV0ZvR1JjN1JOQi9LOHRVUFxcbkxueTVNR25jVGFvRXhOUWR5ZkM0Y3B3M3prZHhqM2NUdXVGVFJzb2tBdHFSRElSUVM2MFl6bTZQMi9vdFNWKytcXG5QN3NMeVdMVUdRZ3ZqTkZNdkZQRGJzSVE5Rm14bUpHb25MVjRpQUFaOVFLQmdRRFFQckpsc2U5U1hZT1ZrQzdmXFxuWEtnK2xmb1luNjBUbTI2YmJKNWV5WC84TlF0Y1BlSFZTcXk4cGRZVnpiR25Xckd2Qzh3dzREdFJFdVVZcytYMFxcblFRZU5TNHNJUFBDZHNqNXBMRjl6d2lGVERReWl0S016RXg2NDlSRnFteVRmV3Z2Z0RueUFqWlJJeU9nS3NTWnJcXG5NOGc1eVY1NVlDdlRPTlovUWYvR2g4NmJkUUtCZ1FDNG5mOFBWWHFRRlI2YWczNHhqNURqc2hXZlhJT3ZaWTg3XFxuclNzK3BUSDJxdWxYYzhudEhZQVBPZktnVGJmZWs1cGRSNStnRzN5WWcrbUcrWE5WQlRheDJXbDVWT2tZU0VxSlxcbkhVZDFrQ25PbVlUZVp0Ri8vdTdzektwRmE4L0hkcUtMNFdETkZaeU9iZ1NYcWdXbGFVM2FWWkhkK3V2ZzM3SEVcXG56M05oUlpIMHp3S0JnUUNzRG1GUGJNaVRnUGdySnNuVGVyYjNudXJZVlhXbThaRmRrVXo0ZS92bTRkelZCYndGXFxuZ29GZURKYnB4TjIzckZPS2tYRFFJVFJoTS85ZGZhWE5QYjJEbkpydTM0cmVnRnJZZ3ZVS3E2YmsrNjhvNzU2M1xcbm9HQ042TTNQQ3doWUV0Qnd1d2RiSDU4WTFBWUViNEdTcVdJUmZMTTJEYU9vRFJvTVl2ZDFqTmZEMFFLQmdDQjNcXG4rUkd6VU5qaVBmMml2cURzeE9pbXUxTEpySWMrYjFCcGhqK0FRaWRGcThBcnB3bkN0SEQ1R2dqRFltRU15SXM3XFxuTzRHbkUrU20zbjFVaGNvZ0hweHN4allHanZBc1ZwK0N2THlhWEIvdnRBU0JSTHNrRk5Va3NaVi8vb3p2K21wclxcbmV1RFd1aS82ZldoSENMTXNyL3FFTGlGQ0xoWGdnWjFCZHVOV252TFZBb0dCQUxLcTZLcmdhNk5oWWpEY3hkN3FcXG5tU2FzRm9qdjN3SDUralBwVmU0akIyeW0xeDJGN2JqQmVWMWdpZDNPVWJqQkRSTllWSldZUkZNNzFxc0xyRUY5XFxuTEhidVNVRnorSVZNOXFnOUhzT3NqWDh3ekI1QVh2bG9Vclo0a1Y3RmJLOUp5WnRXU2RYMnpNNllmY1lJSkdiRlxcbjVHUU9HM3hKM1hZZEJKTm52T1YwcXIvS1xcbi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS1cXG5cIixcbiAgXCJjbGllbnRfZW1haWxcIjogXCJkb2NrZXJAY29uc3RhbnQtY3ViaXN0LTE3MzEyMy5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbVwiLFxuICBcImNsaWVudF9pZFwiOiBcIjExMzA4NDc2Nzg5NTE3MzQ1MTIwOFwiLFxuICBcImF1dGhfdXJpXCI6IFwiaHR0cHM6Ly9hY2NvdW50cy5nb29nbGUuY29tL28vb2F1dGgyL2F1dGhcIixcbiAgXCJ0b2tlbl91cmlcIjogXCJodHRwczovL29hdXRoMi5nb29nbGVhcGlzLmNvbS90b2tlblwiLFxuICBcImF1dGhfcHJvdmlkZXJfeDUwOV9jZXJ0X3VybFwiOiBcImh0dHBzOi8vd3d3Lmdvb2dsZWFwaXMuY29tL29hdXRoMi92MS9jZXJ0c1wiLFxuICBcImNsaWVudF94NTA5X2NlcnRfdXJsXCI6IFwiaHR0cHM6Ly93d3cuZ29vZ2xlYXBpcy5jb20vcm9ib3QvdjEvbWV0YWRhdGEveDUwOS9kb2NrZXIlNDBjb25zdGFudC1jdWJpc3QtMTczMTIzLmlhbS5nc2VydmljZWFjY291bnQuY29tXCJcbn0iLCJlbWFpbCI6InVzZXJAZXhhbXBsZS5jb20iLCJhdXRoIjoiWDJwemIyNWZhMlY1T25zS0lDQWlkSGx3WlNJNklDSnpaWEoyYVdObFgyRmpZMjkxYm5RaUxBb2dJQ0p3Y205cVpXTjBYMmxrSWpvZ0ltTnZibk4wWVc1MExXTjFZbWx6ZEMweE56TXhNak1pTEFvZ0lDSndjbWwyWVhSbFgydGxlVjlwWkNJNklDSmxNek5rT0dGa05EZzVabVJoTVRNeE9EUm1aREZrTmpObVpXTXdPR05qWkdGbE9XUmxNVGt3SWl3S0lDQWljSEpwZG1GMFpWOXJaWGtpT2lBaUxTMHRMUzFDUlVkSlRpQlFVa2xXUVZSRklFdEZXUzB0TFMwdFhHNU5TVWxGZG1kSlFrRkVRVTVDWjJ0eGFHdHBSemwzTUVKQlVVVkdRVUZUUTBKTFozZG5aMU5yUVdkRlFVRnZTVUpCVVVOWFRGcFpibUpUY0c1bVRWUnpYRzVVYkVwTFdWTnBWRFZzYjNkYWIzbGtUMlozVERSRU1UTkhRbEo1WjNwcGRtbHFRemxMVldwUk9HdHJkbU5PVFhoQ2NIaDBVbEZMVFdGWGRsSXphR2gyWEc1WGVqRTRlQzl4WlcxRGRsaHVWMUZNTW5neE56SnplVFZ3Y1RSc1ZubGFVekpLYmk5ak5IaEpRVk5HTkhoWGIzbHZRVnBzTlhwRlprOHdlbTAxTTNCa1hHNDBLMnAzUTB4d04wTkZLelZSVWxOeVVHTjFlRWgyUTI5bFNFcDVlSEZZYTJZNFltaE5XbGMzVUVsQlJVMHhhVmxJUjAxdlRXNXJWVTUxVVdKVFJXOU9YRzVNTkVoeVUzRklNR1oxZVVaRlN6WXhVV0ZDU2tkMFVuWmFNbVZXVTFKb1pGQkJXa1V4UkVGdVRsaFZaWFZ5TkRVeE4yNUlWMEU1ZVZsVVUwRlhaMFUyWEc0MVUwMVRTRVZ4V1ZWTVNtZE1NemQyU2t0aVVVRndTakF3WjI1bk5uQk5URkpFYW5CMGFGZEJhMGxCTDNCWWFteE5SMmxtTW5ac2RuWTVSRVZSTjNwMVhHNUlWSEJUWjBSbFlrRm5UVUpCUVVWRFoyZEZRVk5OTTNRd2MzaEtPU3RUVmxwUlpuUlFaa1JRU25KQ1VWQmtSMmhVY1VkeVRIRnBNWE0xVWtKaFYyaDZYRzVLVTNGc2VWUmhTQzltTDFCbFoyWkZOSEZ2VmxWTFdGcG1LeXRzYmxKalEwZHZORTVvTXpRNVdXZENZMnhOYkZKR1MzaFVjRlpWYWpGallsZ3JkazByWEc1bFdVSllWeXNyVFRkUFZtSmpSSGx2WVUxWFMyaEpSbkJ1TXpNd2IwdGFUV1pPVERrdlRYZEhaRFZ1UjJGSlYwUXJlVnBaY0hSUVkydEtabVo1UVUxSFhHNVBWMEZOTjNweGQydEVlRGRFTlhoU2FEWlBRM052V1RKMlUzQlFhSFp5WW5OdFRFNTFOV1ZpYmxOWlRFOVlVbVZYUm05SFVtTTNVazVDTDBzNGRGVlFYRzVNYm5rMVRVZHVZMVJoYjBWNFRsRmtlV1pETkdOd2R6TjZhMlI0YWpOalZIVjFSbFJTYzI5clFYUnhVa1JKVWxGVE5qQlplbTAyVURJdmIzUlRWaXNyWEc1UU4zTk1lVmRNVlVkUlozWnFUa1pOZGtaUVJHSnpTVkU1Um0xNGJVcEhiMjVNVmpScFFVRmFPVkZMUW1kUlJGRlFja3BzYzJVNVUxaFpUMVpyUXpkbVhHNVlTMmNyYkdadldXNDJNRlJ0TWpaaVlrbzFaWGxZTHpoT1VYUmpVR1ZJVmxOeGVUaHdaRmxXZW1KSGJsZHlSM1pET0hkM05FUjBVa1YxVlZseksxZ3dYRzVSVVdWT1V6UnpTVkJRUTJSemFqVndURVk1ZW5kcFJsUkVVWGxwZEV0TmVrVjROalE1VWtaeGJYbFVabGQyZG1kRWJubEJhbHBTU1hsUFowdHpVMXB5WEc1Tk9HYzFlVlkxTlZsRGRsUlBUbG92VVdZdlIyZzRObUprVVV0Q1oxRkRORzVtT0ZCV1dIRlJSbEkyWVdjek5IaHFOVVJxYzJoWFpsaEpUM1phV1RnM1hHNXlVM01yY0ZSSU1uRjFiRmhqT0c1MFNGbEJVRTltUzJkVVltWmxhelZ3WkZJMUsyZEhNM2xaWnl0dFJ5dFlUbFpDVkdGNE1sZHNOVlpQYTFsVFJYRktYRzVJVldReGEwTnVUMjFaVkdWYWRFWXZMM1UzYzNwTGNFWmhPQzlJWkhGTFREUlhSRTVHV25sUFltZFRXSEZuVjJ4aFZUTmhWbHBJWkN0MWRtY3pOMGhGWEc1Nk0wNW9VbHBJTUhwM1MwSm5VVU56UkcxR1VHSk5hVlJuVUdkeVNuTnVWR1Z5WWpOdWRYSlpWbGhYYlRoYVJtUnJWWG8wWlM5MmJUUmtlbFpDWW5kR1hHNW5iMFpsUkVwaWNIaE9Nak55Ums5TGExaEVVVWxVVW1oTkx6bGtabUZZVGxCaU1rUnVTbkoxTXpSeVpXZEdjbGxuZGxWTGNUWmlheXMyT0c4M05UWXpYRzV2UjBOT05rMHpVRU4zYUZsRmRFSjNkWGRrWWtnMU9Ga3hRVmxGWWpSSFUzRlhTVkptVEUweVJHRlBiMFJTYjAxWmRtUXhhazVtUkRCUlMwSm5RMEl6WEc0clVrZDZWVTVxYVZCbU1tbDJjVVJ6ZUU5cGJYVXhURXB5U1dNcllqRkNjR2hxSzBGUmFXUkdjVGhCY25CM2JrTjBTRVExUjJkcVJGbHRSVTE1U1hNM1hHNVBORWR1UlN0VGJUTnVNVlZvWTI5blNIQjRjM2hxV1VkcWRrRnpWbkFyUTNaTWVXRllRaTkyZEVGVFFsSk1jMnRHVGxWcmMxcFdMeTl2ZW5ZcmJYQnlYRzVsZFVSWGRXa3ZObVpYYUVoRFRFMXpjaTl4UlV4cFJrTk1hRmhuWjFveFFtUjFUbGR1ZGt4V1FXOUhRa0ZNUzNFMlMzSm5ZVFpPYUZscVJHTjRaRGR4WEc1dFUyRnpSbTlxZGpOM1NEVXJhbEJ3Vm1VMGFrSXllVzB4ZURKR04ySnFRbVZXTVdkcFpETlBWV0pxUWtSU1RsbFdTbGRaVWtaTk56RnhjMHh5UlVZNVhHNU1TR0oxVTFWR2VpdEpWazA1Y1djNVNITlBjMnBZT0hkNlFqVkJXSFpzYjFWeVdqUnJWamRHWWtzNVNubGFkRmRUWkZneWVrMDJXV1pqV1VsS1IySkdYRzQxUjFGUFJ6TjRTak5ZV1dSQ1NrNXVkazlXTUhGeUwwdGNiaTB0TFMwdFJVNUVJRkJTU1ZaQlZFVWdTMFZaTFMwdExTMWNiaUlzQ2lBZ0ltTnNhV1Z1ZEY5bGJXRnBiQ0k2SUNKa2IyTnJaWEpBWTI5dWMzUmhiblF0WTNWaWFYTjBMVEUzTXpFeU15NXBZVzB1WjNObGNuWnBZMlZoWTJOdmRXNTBMbU52YlNJc0NpQWdJbU5zYVdWdWRGOXBaQ0k2SUNJeE1UTXdPRFEzTmpjNE9UVXhOek0wTlRFeU1EZ2lMQW9nSUNKaGRYUm9YM1Z5YVNJNklDSm9kSFJ3Y3pvdkwyRmpZMjkxYm5SekxtZHZiMmRzWlM1amIyMHZieTl2WVhWMGFESXZZWFYwYUNJc0NpQWdJblJ2YTJWdVgzVnlhU0k2SUNKb2RIUndjem92TDI5aGRYUm9NaTVuYjI5bmJHVmhjR2x6TG1OdmJTOTBiMnRsYmlJc0NpQWdJbUYxZEdoZmNISnZkbWxrWlhKZmVEVXdPVjlqWlhKMFgzVnliQ0k2SUNKb2RIUndjem92TDNkM2R5NW5iMjluYkdWaGNHbHpMbU52YlM5dllYVjBhREl2ZGpFdlkyVnlkSE1pTEFvZ0lDSmpiR2xsYm5SZmVEVXdPVjlqWlhKMFgzVnliQ0k2SUNKb2RIUndjem92TDNkM2R5NW5iMjluYkdWaGNHbHpMbU52YlM5eWIySnZkQzkyTVM5dFpYUmhaR0YwWVM5NE5UQTVMMlJ2WTJ0bGNpVTBNR052Ym5OMFlXNTBMV04xWW1semRDMHhOek14TWpNdWFXRnRMbWR6WlhKMmFXTmxZV05qYjNWdWRDNWpiMjBpQ24wPSJ9fX0=
---
apiVersion: v1
kind: Secret
metadata:
  name: client-secret
type: Opaque
data:
  CLIENT_ID: MzM2MzM1NTQxOTkzLWdzZTFyMnZvc3Q1Z2JiMTN0ZWpjYmk0M3UyY3NjYTRpLmFwcHMuZ29vZ2xldXNlcmNvbnRlbnQuY29t
  CLIENT_SECRET: ZFBIbmFvbUc3dUNodjFVTWY0bVFuX0tk
---
apiVersion: v1
kind: Secret
metadata:
  name: kfctl-e2e-secret
type: Opaque
data:
  kfctl-e2e.json: ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsCiAgInByb2plY3RfaWQiOiAiY29uc3RhbnQtY3ViaXN0LTE3MzEyMyIsCiAgInByaXZhdGVfa2V5X2lkIjogIjRhZGJmYmI0M2I1ZWJjMTMyZjRlOGM4Y2NmZjBkZDJjMjg2ZjhhY2QiLAogICJwcml2YXRlX2tleSI6ICItLS0tLUJFR0lOIFBSSVZBVEUgS0VZLS0tLS1cbk1JSUV2d0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktrd2dnU2xBZ0VBQW9JQkFRQ2hQTzVlTkJVcFJ6YWpcbnJpSlltMHpPUG4vUWpEZGlhODdwc3RtMmwvOC92MjM5anFHWjJhV1RIZ2lsQ3hESTRmZFRhTmY4VE1zc0lGT0Vcbjh1YmgxU3hvVW11T0hHU2RDSTgxaFVDMWVobjJNV2owRTAySk5LTHVWRStOR0lLcW9WZUduN3RCRzEwUTIwalBcbllCN1dzNGlkcmVzUnNqeVVkUnhSelo3TEJSc2NxRkIzYSsvVUcxdERlMGszZjgxVGEzYi85TTdoRUsyU3pxL2FcblA3RElRK3NzenFsQzVUdzNlM1lGZHl1SHAzTVAxY09aMjFrNE56TzMvdHNGUTdZVFBKLytlemxxaWJ5azdVRkRcbjF5YzJRSU1UMUdlMS92SmkvQ1BGa2JUL2dNM252cjZlZ0RQb2lzbFF0WGxoaks4ZVJJQ0cwbkI5WFIvN3ZCY1BcbnBKQVBKK0xGQWdNQkFBRUNnZ0VBRHA2alpSVkhOSmFleEs2ZHNmZTY1TVIrcGJQaENHS0t3TFBxZTdKQ2Z2bVFcbjNvOFg1Z2I1dFdJMnhITHNFd1dIZmNwTU16WHFBaXZ0aHI0dU9ISGN5ZTlzamRaOTVQaUpZUC96d3hZTzArZUhcbnJheUdZNWQxNG9oRjE1NzQvTUc1QkVnOWNWb3EwUmwyVVBTcWpJQjdpd00zQk83QUFYSUdTYXMyNTdwUElBVHZcbllvdFl4Tzc5NHFFSFpSWTY3OWRLRS9NVVNUM1JvZUxoazNxN3EzZXhYenRTZklmM1ZJL2Q2aTVOSHJUWDc0b3RcbmFhQnFIYmE4VVNYK21HVzNWQnZ4WWpTeDE0ZlBjRnBFYkRzaWJ0NC9zOW11VnBXVFpBTlljZ3NwSHA4OWFiR0xcblVXcEFMbEY1WU9lUER3aUR4WE55S2piZ1d4MTVUZy9Sc1U5RU40ZDViUUtCZ1FEVW16MzlYQ3lzaGc3OHNiTm9cbmJUcDU0WkJDZkVyS2NLTCsrUnZ0OVNuNGdEM1l5K2FaL0VQTjBvNHJDQVdjSG1SUDE0TThNV2tHUElsRklJOEhcbjVGQmhJUVlhdjdNaUV4UUVodCtuRy9NbEtqVzA3K2x1Z2w2dXpjVUhGa2lveU81QlZTa1VWM1ZkMklIZG5ublZcbjFMeEg5aVVYTFRSMFQzNmcvc21xQzY2T0J3S0JnUURDSmF3aHBPZVJWUWNkbWs1L05ENXU5TmVCbjc3NGxnVjFcbld1YTFqN3g0UGU3bWYrK2F3VzhmM0ZaYlNKNmphcHdwWGdDVm5reEdDeE8zbUNiOFFNb2x3Y2d0WjlXRGc3ZG1cblhJT0c5NEcyekZaVU51cUpPUWN2bXdqMDQ0QzRXOWpuYkw0VzJySGI1TmJxS0dPZWd5TFRVMkxId2hKV01TT0ZcbjE1cEJOYnZWMHdLQmdRQys1TlQ0RkRjWWdSWWIxZ0pjbFJhWU1PdXlocDh3dlluY25oZHh4VnQyQU0rSFJTMDhcbnZjQ3pvWVo5SktyRXpwVUxDMXFPUlY4amRsOWFiaTEreklWUGNMTm1lUkdDV3RieWFaZVBHQTF0SlVVcnZPNC9cbmgxYzBaUld2azFhU0ZqZTIrWnYwNDhKQ0RSQXR5UWxqOGF0TWdib3o0U2JqK0N5ZXFhYXd3K1JyS1FLQmdRQ1BcblR5K1lSaG1JOWJLaTcxd3lHV1pja083akIvLzNqd1hJY3FrS0xHZDZlbnoyT0VtdGVrdUV3U2dkaWFWUUMwbnFcblh6RWZRQklkUWQvMERhUDVYL25YbFFzbU9SY3FWUGZ6M3laWFlpdWx1MytkK2t0MXIxcldrU0l6WWh1SitvSkpcbmtjTmZLMTlPYWNVYVkyWGxnL2NZOXR4Ymg2M3hZYVJQRDY4Vm80eGJjUUtCZ1FDbmc1REFsbDUrKzI4S2hSUitcbk8zVzVhYTFlTWtidUV3Z1l4Uk1Ocm5qSU1zMTEwOEkrc3NNOEpjQ2RqLzI4dnNDczgvUGUvOVA2cWJVbGViNlNcbmdJU1ZjTzV0U1FtQVY4M3h4d0dnTkJYQy9wRU5xT2NMU3haN01UaVdZWTVtOVBxZEM2U090Y2lzZXdsNFRWZ1lcblNzQ0ptZVN1d1NBaDFzd25hd2p4MkY3eU53PT1cbi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS1cbiIsCiAgImNsaWVudF9lbWFpbCI6ICJrZmN0bC1lMmVAY29uc3RhbnQtY3ViaXN0LTE3MzEyMy5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbSIsCiAgImNsaWVudF9pZCI6ICIxMDMxMTQ1Njg2Njg2Nzc5MzY4NTEiLAogICJhdXRoX3VyaSI6ICJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20vby9vYXV0aDIvYXV0aCIsCiAgInRva2VuX3VyaSI6ICJodHRwczovL29hdXRoMi5nb29nbGVhcGlzLmNvbS90b2tlbiIsCiAgImF1dGhfcHJvdmlkZXJfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9vYXV0aDIvdjEvY2VydHMiLAogICJjbGllbnRfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9yb2JvdC92MS9tZXRhZGF0YS94NTA5L2tmY3RsLWUyZSU0MGNvbnN0YW50LWN1YmlzdC0xNzMxMjMuaWFtLmdzZXJ2aWNlYWNjb3VudC5jb20iCn0K
`)
	th.writeF("/manifests/e2e/e2e-tasks/base/service-account.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: e2e-pipelines
imagePullSecrets:
- name: docker-secret
`)
	th.writeF("/manifests/e2e/e2e-tasks/base/cluster-role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: e2e-pipelines-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: e2e-pipelines
`)
	th.writeF("/manifests/e2e/e2e-tasks/base/task.yaml", `
---
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: kfctl-build-push
spec:
  inputs:
    resources:
    - name: docker-source
      type: git
    params:
    - name: pathToDockerFile
      type: string
      description: The path to the dockerfile to build
      default: /workspace/docker-source/Dockerfile
    - name: pathToContext
      type: string
      description:
        The build context used by Kaniko
        (https://github.com/GoogleContainerTools/kaniko#kaniko-build-contexts)
      default: /workspace/docker-source
  outputs:
    resources:
    - name: builtImage
      type: image
      outputImageDir: /workspace/builtImage
  steps:
  - name: build-and-push
    image: gcr.io/kaniko-project/executor:v0.10.0
    command:
    - /kaniko/executor
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    args: ["--dockerfile=${inputs.params.pathToDockerFile}",
           "--destination=${outputs.resources.builtImage.url}",
           "--context=${inputs.params.pathToContext}",
           "--target=kfctl_base"]
    volumeMounts:
    - name: kaniko-secret
      mountPath: /secret
  volumes:
  - name: kaniko-secret
    secret:
      secretName: kaniko-secret
---
apiVersion: tekton.dev/v1alpha1
kind: Task
metadata:
  name: kfctl-init-generate-apply
spec:
  inputs:
    resources:
    - name: image
      type: image
    params:
    - name: namespace
      type: string
      description: the namespace to deploy kf 
    - name: app_dir
      type: string
      description: where to create the kf app
    - name: configPath
      type: string
      description: url for config arg
    - name: project
      type: string
      description: name of project
    - name: zone
      type: string
      description: zone of project
    - name: platform
      type: string
      description: all | k8s
    - name: email
      type: string
      description: email for gcp
  outputs:
    resources:
    - name: builtImage
      type: image
      outputImageDir: /workspace/builtImage
  steps:
  - name: kfctl-init
    image: "${inputs.resources.image.url}"
    command: ["/usr/local/bin/kfctl"]
    args:
    - "init"
    - "--config"
    - "${inputs.params.configPath}"
    - "--project"
    - "${inputs.params.project}"
    - "--namespace"
    - "${inputs.params.namespace}"
    - "${inputs.params.app_dir}"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    volumeMounts:
    - name: kaniko-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
    imagePullPolicy: Always
  - name: kfctl-generate
    image: "${inputs.resources.image.url}"
    imagePullPolicy: Always
    workingDir: "${inputs.params.app_dir}"
    command: ["/usr/local/bin/kfctl"]
    args:
    - "generate"
    - "${inputs.params.platform}"
    - "--zone"
    - "${inputs.params.zone}"
    - "--email"
    - "${inputs.params.email}"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kaniko-secret.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: kaniko-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  - name: kfctl-apply
    image: "${inputs.resources.image.url}"
    imagePullPolicy: Always
    workingDir: "${inputs.params.app_dir}"
    command: ["/usr/local/bin/kfctl"]
    args:
    - "apply"
    - "${inputs.params.platform}"
    - "--verbose"
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /secret/kfctl-e2e.json
    - name: CLIENT_ID
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_ID
    - name: CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: client-secret
          key: CLIENT_SECRET
    volumeMounts:
    - name: kfctl-e2e-secret
      mountPath: /secret
    - name: kubeflow
      mountPath: /kubeflow
  volumes:
  - name: kaniko-secret
    secret:
      secretName: kaniko-secret
  - name: docker-secret
    secret:
      secretName: docker-secret
  - name: kfctl-e2e-secret
    secret:
      secretName: kfctl-e2e-secret
  - name: kubeflow
    persistentVolumeClaim:
      claimName: kubeflow-pvc
`)
	th.writeK("/manifests/e2e/e2e-tasks/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- persistent-volume-claim.yaml
- secret.yaml
- service-account.yaml
- cluster-role-binding.yaml
- task.yaml
namespace: tekton-pipelines
`)
}

func TestE2eTasksOverlaysE2e(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/e2e/e2e-tasks/overlays/e2e")
	writeE2eTasksOverlaysE2e(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../e2e/e2e-tasks/overlays/e2e"
	fsys := fs.MakeRealFS()
	_loader, loaderErr := loader.NewLoader(targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl())
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	n, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := n.EncodeAsYaml()
	th.assertActualEqualsExpected(m, string(expected))
}
