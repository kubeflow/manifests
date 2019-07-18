# Editing Overlay files

Following instructions will help you configure your setup with values for each overlay file. This is a guideline instruction. Please edit these files appropriately if you need to change other values than what is being mentioned here.

## dex.yaml

### Insert certificate in `ca` configMap

This certificate should be the signing authority ca certificate file for dex.example.com.

### Edit tls secret in `dex` Deployment

Replace `dex.example.com.tls` with your own secret as created during certificate setup stage.

### Edit `dex` configMap

#### In data:config.yaml

- `issuer` needs to be set to your own domain for dex server
- `connectors`: Within the element with *type: ldap* change the paramter *config:host* to the domain or IP of the LDAP server. For eg. ldap.auth.svc.cluster.local:389
- `staticClients`: *id* and *secret* can be modified to a different value, this should be reflected in the client application configuration. *redirectURIs* needs to be set to your own domain for client application for dex.

## dex_k8s_authenticator.yaml

### Edit `dex-k8s-authenticator-cm` configMap data:config.yaml section

- `cluster` has these following values to be set
  - *name* to an appropriate name for your Kubernetes cluster
  - *redirect_uri, client_id, client_secret and issuer* need to be set as the same as mentioned in dex configMap staticClients.
  - *k8s_master_uri* needs to be set to the Kubernetes API server URI.
  - *k8s_ca_pem* needs to be populated with the Kubernetes CA cert file contents obtained by  
  `kubectl config view --raw -o json | jq -r '.clusters[0].cluster."certificate-authority-data"' | tr -d '"' | base64 --decode`  
  if your kube config file exists.
- *idp_ca_pem* needs to be populated with dex's domain's SSL ca cert file contents.
- *listen* needs to be set to the appropriate domain and port.

## authentication_policy.yaml

### Edit `pipelines-auth-policy`

- `spec:origins:[jwt]` has the following values to be set
  - *audiences* To be set as the same client id mentioned in dex configMap staticClients
  - *issuer* Needs to be set to your own dex server's domain, same as dex configMap staticClients
  - *jwksUri* Needs to be set to your own dex server's jwks keys location
