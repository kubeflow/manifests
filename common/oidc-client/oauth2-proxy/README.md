* This deployment of oauth2-proxy doesn't support OpenShift. To enable
  integration with OpenShift, use openshift distribution of oauth2-proxy
  available here: https://github.com/openshift/oauth-proxy
  * also enable RBAC for token reviews:
    ```yaml

    ```

* Login routine will use /oauth2 to manage actions
    * When oauth2-proxy detects need to will redirect to dex and provide
        * oidc client id
        * redirect url to kubeflow with path to `/oauth2/callback`
    * Login starts with /oauth2/callback
    * if using oauth2-proxy for multiple domains, must ensure path based routing to `/oauth2` routes to oauth2-proxy...???
* Include VirtualService for logging out (did axel already do that?)