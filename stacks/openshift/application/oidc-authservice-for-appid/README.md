# oidc-authservice for IBM Cloud AppID

## Prerequisites

* Provisioning an [AppID](https://cloud.ibm.com/catalog/services/app-id) 
instance from IBM Cloud. It can start with the _Lite_ plan, but will need the
_Graduated tier_ once you need more than 1000 authentication events per month.
* FQDN of OpenShift Route of istio ingress gateway.
* Create an application with type _reguarwebapp_ under the provioned AppID
instance. Make sure the caope contains `email` and retrieve the following
configuration parameters from your AppID. They will be used to configure the
OIDC auth service:
    - `clientId`
    - `secret`
    - `oAuthServerUrl`

## How to use

1. Create the namespace `istio-system` if not exist:
```SHELL
kubectl create namespace istio-system
```
2. Create a secret prior to kubeflow deployment by filling parameters accordingly:
```SHELL
kubectl create secret generic appid-application-configuration -n istio-system \
  --from-literal=clientId=<clientId> \
  --from-literal=secret=<secret> \
  --from-literal=oAuthServerUrl=<oAuthServerUrl> \
  --from-literal=oidcRedirectUrl=https://<routeFQDN>/login/oidc
```
* `<oAuthServerUrl>` - fill in the value of `oAuthServerUrl`
* `<clientId>` - fill in the value of `clientId`
* `<secret>` - fill in the value of `secret`
* `<routeFQDN>` - fill in the FQDN of OpenShift Route of istio ingress gateway

Notice that it recommend using HTTPS for the value of `oidcRedirectUrl`, which
requires additional setup:
1. enable [TLS passthrough](https://docs.openshift.com/enterprise/3.0/architecture/core_concepts/routes.html#passthrough-termination) mode for the route.
2. expose kubeflow dashboard over HTTPS by following steps of [this section](https://www.kubeflow.org/docs/ibm/deploy/authentication/#exposing-the-kubeflow-dashboard-with-dns-and-tls-termination).

After deploying Kubeflow successfully, you will need to add the value of
`https://<routeFQDN>/login/oidc` to the IBM Cloud AppID instance under the
_Authentication Settings_ of the _Manage Authentication_ menu. Or AppID won't
redirect authenticated requests back to Kubeflow.