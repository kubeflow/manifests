# oidc-authservice for IBM Cloud AppID

## Prerequisites

* Provisioning an AppID instance from IBM Cloud. the free tier works.
* keep the configuration parameters including:
    - client_id
    - OIDC provider URL
    - client_secret

## How to use

After provisioning the statefulset `authservice` in namespace `istio-system`,
update the environment variables via either web console or CLI with parameters
from AppID instance.

Notice that the environment variable `REDIRECT_URL` should be updated with the
actual FQDN of public endpoint of istio ingressgateway, either via ingress or
route. And please keep the path to `/login/oidc`.