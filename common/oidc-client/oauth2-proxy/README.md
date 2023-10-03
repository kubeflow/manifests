# Kubeflow with oauth2-proxy and envoyExtAuthzHttp

For a quick install, see [Example installation](#example-installation).


## Description

Kubeflow authorization currently works based on custom auth headers:
* `kubeflow-userid` with user email
* `kubeflow-groups` with a comma-separated list of groups
    * this is not yet fully functional

This was implemented with custom, minimalistic authorization tool from arrikto
called `oidc-authservice`, which was integrated with Istio using `EnvoyFilter`.

Using `envoyExtAuthzHttp` for authentication is better, because:
* The `envoyExtAuthzHttp` extension simplifies the process of adding external
  authorization to the Envoy proxy within an Istio service mesh. It provides a
  declarative way to configure and apply authorization policies, avoiding the
  complexity of directly modifying EnvoyFilter configurations.
* Istio promotes using the `envoyExtAuthzHttp` extension for adding external
  authorization because it follows a standardized approach. This allows for
  consistency across different Istio deployments and makes it easier to
  understand and maintain the authorization logic within the service mesh.
* By using `envoyExtAuthzHttp`, authorization policies are defined in separate
  configuration resources instead of modifying the EnvoyFilter directly. This
  separation simplifies policy management and makes it easier to update or
  replace authorization logic as needed without modifying EnvoyFilter
  configurations.
* The `envoyExtAuthzHttp` extension integrates well with other Istio features
  such as AuthorizationPolicy and VirtualService. It allows for easier
  integration with Istio's broader ecosystem and leverages the benefits of
  Istio's architecture and capabilities.
* It's highly possible that the internal auth decisions in Kubeflow will
  be based on the JWT directly in future instead of the custom auth headers.

Additional information on why it's worth moving out of `EnvoyFilter` can be found
under these resources:
* https://github.com/istio/istio/issues/27790
* https://istio.io/latest/docs/tasks/security/authorization/authz-custom/

`envoyExtAuthzHttp` could probably be integrated with `oidc-authservice`
but `oauth2-proxy` is a more advanced authentication proxy that has a greater
community support and used across the industry, including official istio documentation on [External Authorization](https://istio.io/latest/docs/tasks/security/authorization/authz-custom).

For more details on the `oauth2-proxy`, see the [official documentation](https://oauth2-proxy.github.io/oauth2-proxy/docs/behaviour).

## OpenShift

This deployment of oauth2-proxy doesn't support OpenShift. To enable integration
with OpenShift:
* use OpenShift distribution of oauth2-proxy available here:
    * https://github.com/openshift/oauth-proxy
* enable RBAC for token reviews by adding the `rbac.tokenreviews.yaml` file
  to the kustomize (see comment in `kustomization.yaml`)

## CloudFlare

CloudFlare will require that some of the static, standard web browser assets are
available without user authentication. This is important because CloudFlare will
want to cache these assets for performance and if these assets are not available
without user authentication, CloudFlare robots will be redirected to the
authentication page. This can result in general issues while accessing the
Kubeflow instance behind CloudFlare.

This can be easily solved by specifying in Istio `AuthorizationPolicy` a set of
assets that don't require authentication. Such `AuthorizationPolicy` definition
is already available under file
`authorizationpolicy.istio-ingressgateway-oauth2-proxy-cloudflare.yaml`.

## Explaining the auth routine

1. Istio is configured with `envoyExtAuthzHttp` extension provider pointing to
   `oauth2-proxy` service. This enables usage of this extension in
   `AuthorizationPolicy` to add external authorization to the service mesh.
2. Istio service mesh is configured with `AuthorizationPolicy`
   `istio-ingressgateway-oauth2-proxy` in `istio-system` (istio root) namespace
   with `action: CUSTOM` with `oauth2-proxy` provider. This means that every
   request to the `istio-ingressgateway` must go through the external
   authorization service, which is `oauth2-proxy`.
3. `oauth2-proxy` will make decisions based on the cookie named
   `oauth2_proxy_kubeflow`. If the cookie is not present, expired or invalid,
   `oauth2-proxy` will redirect to the configured oidc provider, which in
   default setup is `dex`.
    * some of the information passed in the auth redirect are:
        * redirect_uri pointing to the URI that the user should be redirected to
          after successful authentication,
        * code challenge to protect against interception and replay attacks,
        * state parameter to ensure that the authorization response is valid,
          originating from the same request that was initially sent, and that it
          hasn't been tampered with or intercepted by an attacker.

   After this step and successful authentication, cookie is available in user
   browser which is then used by istio and oauth2-proxy to authorize user
   requests.

   Alternatively, `oauth2-proxy` will also look for an `Authorization` header
   with a JWT bearer token. `oauth2-proxy` will automatically trust the JWT
   issued by the configured oidc provider, which in the default setup is dex.
   Additional trusted JWTs can be configured in a way that `oauth2-proxy` will
   skip the authentication and forward the request back to istio including
   provided authorization header.

   The ultimate goal of this step is to provide the `Authorization` header
   with JWT Bearer Token which is later used in `RequestAuthentication` to
   configure Istio to trust this JWT, parse JWT Claims to include user email
   and groups in custom Kubeflow authorization headers, allow routing and
   authorization decision based on JWT Claims.
4. In this step Istio will always have the JWT in the request making it possible
   to make authorization decision based on JWT Claims.

## Using HTTPS
`oauth2-proxy` is configured with `http` endpoint secured by Istio Service Mesh.
Because of that, `oauth2-proxy` will guess that the auth redirect uri is supposed
to be handler with `http` as well. To force using `https`, change variable
`FORCE_HTTPS` to `true` in `kustomization.yaml`. This will use `oauth2-proxy`
configuration option `--cookie-secure` which will redirect with `https`.

## Istio JWT Pub Key Refresh Interval
On initial Kubeflow setup, it's highly possible that `istiod` will become
available before `dex`. Using Istio `RequestAuthentication` configures Istio to call
the Issuer URL to gather the JWT Pub Key. If Issuer is not yet available, placeholder
keys will be configured and the setup will become unusable until Istio has access
to the correct JWT Pub Key. For that reason, `istiod` is configured with env variable
`PILOT_JWT_PUB_KEY_REFRESH_INTERVAL="1m"` so the JWT Pub Key is refreshed every
minute instead of the default 20 minutes.

If this is not configured, user will be presented with following Istio error:
```
Jwks doesn't have key to match kid or alg from Jwt
```

## Issues with this setup
While not directly an Istio or oauth2-proxy issue, currently Kubeflow is configured
to always redirect to the URL provided in logout response body object under key
`afterLogoutURL`. This is custom integration with `oidc-authservice` component.
oauth2-proxy has redirect capabilities and can redirect to the base Kubeflow
page but because of this custom integration, currently logging out will remove
the authentication cookie but not redirect to the Kubeflow Home Page.

The details of this custom integration can be found here:
* https://github.com/arrikto/oidc-authservice/blob/0c4ea9a/server.go#L509
* https://github.com/kubeflow/kubeflow/blob/c6c4492/components/centraldashboard/public/components/logout-button.js#L50

To login again, user have to refresh the page.

## Example installation
To use `oauth2-proxy` with `istio` `envoyExtAuthzHttp`, following changes has to
be done to the `example/kustomization.yaml` file:
* change `OIDC Authservice` to `oauth2-proxy for OIDC`
  ```
  # from
  - ../common/oidc-client/oidc-authservice/base
  # to
  - ../common/oidc-client/oauth2-proxy/base
  ```
* change Dex overlay
  ```
  # from
  - ../common/dex/overlays/istio
  # to
  - ../common/dex/overlays/oauth2-proxy
* add Kustomize Components to configure Istio, oauth2-proxy, Kubernetes and
  Central Dashboard with `oauth2-proxy` using `envoyExtAuthzHttp`
  ```
  components:
  # oauth2-proxy components to configure Istio, oauth2-proxy, Kubernetes and Central Dashboard
  - ../common/oidc-client/oauth2-proxy/components/istio-external-auth
  - ../common/oidc-client/oauth2-proxy/components/istio-use-kubernetes-oidc-issuer
  - ../common/oidc-client/oauth2-proxy/components/allow-unauthenticated-issuer-discovery
  - ../common/oidc-client/oauth2-proxy/components/configure-self-signed-kubernetes-oidc-issuer
  - ../common/oidc-client/oauth2-proxy/components/central-dashboard
  ```

All those changes combined can be done with this single command:
```diff
$ git apply <<EOF
diff --git a/example/kustomization.yaml b/example/kustomization.yaml
index c1a85789..cad8bb8b 100644
--- a/example/kustomization.yaml
+++ b/example/kustomization.yaml
@@ -39,10 +39,10 @@ resources:
 - ../common/istio-1-17/istio-crds/base
 - ../common/istio-1-17/istio-namespace/base
 - ../common/istio-1-17/istio-install/base
-# OIDC Authservice
-- ../common/oidc-client/oidc-authservice/base
+# oauth2-proxy for OIDC
+- ../common/oidc-client/oauth2-proxy/base
 # Dex
-- ../common/dex/overlays/istio
+- ../common/dex/overlays/oauth2-proxy
 # KNative
 - ../common/knative/knative-serving/overlays/gateways
 - ../common/knative/knative-eventing/base
@@ -85,3 +85,11 @@ resources:
 # KServe
 - ../contrib/kserve/kserve
 - ../contrib/kserve/models-web-app/overlays/kubeflow
+
+components:
+# oauth2-proxy components to configure Istio, oauth2-proxy, Kubernetes and Central Dashboard
+- ../common/oidc-client/oauth2-proxy/components/istio-external-auth
+- ../common/oidc-client/oauth2-proxy/components/istio-use-kubernetes-oidc-issuer
+- ../common/oidc-client/oauth2-proxy/components/allow-unauthenticated-issuer-discovery
+- ../common/oidc-client/oauth2-proxy/components/configure-self-signed-kubernetes-oidc-issuer
+- ../common/oidc-client/oauth2-proxy/components/central-dashboard
EOF
```
