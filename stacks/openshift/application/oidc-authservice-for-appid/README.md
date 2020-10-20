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

Create a secret prior to kubeflow deployment by filling parameters accordingly:
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
* `<routeFQDN>` - fill in the public endpoint of istio ingressgateway.

Notice that the environment variable `REDIRECT_URL` should be updated with the
actual FQDN of public endpoint of istio ingressgateway, either via ingress or
route. And please keep the path to `/login/oidc`.

After deploying Kubeflow successfully, you will need to add the value of
`oidcRedirectUrl` to the IBM Cloud AppID instance under the _Authentication 
Settings_ of the _Manage Authentication_ menu. Or AppID won't redirect authenticated requests
back to Kubeflow.