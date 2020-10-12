# oidc-authservice for IBM Cloud AppID

## Prerequisites

* Provisioning an [AppID](https://cloud.ibm.com/catalog/services/app-id) 
instance from IBM Cloud. It can start with the _Lite_ plan, but will need the
_Graduated tier_ once you need more than 1000 authentication events per month.
* Create an application with type _reguarwebapp_ under the provioned AppID
instance. Make sure the caope contains `email` and retrieve the following
configuration parameters from your AppID. They will be used to configure the
OIDC auth service:
    - `clientId`
    - `secret`
    - `oAuthServerUrl`

## How to use

This configuration should be included as part of kubeflow deployment. After a successful Kubeflow deployment, you will find no pod running in the statefulset
`authservice` in namespace `istio-system`. It will need to fill in valid 
configuration from the environment variables via either web console or CLI with parameters from AppID instance:
* `OIDC_PROVIDER` - fill in the value of `oAuthServerUrl`
* `CLIENT_ID` - fill in the value of `clientId`
* `CLIENT_SECRET` - fill in the value of `secret`
* `REDIRECT_URL` - fill in `https://<hostname-of-the-route-istio-ingressgateway>/login/oidc`

Notice that the environment variable `REDIRECT_URL` should be updated with the
actual FQDN of public endpoint of istio ingressgateway, either via ingress or
route. And please keep the path to `/login/oidc`.

After updating the statusfulset `authservice`, you will need to add the value of
`REDIRECT_URL` to the AppID instance under the _Authentication Settings_ of the _Manage Authentication_ menu. Or AppID won't redirect authenticated requests
back to Kubeflow.