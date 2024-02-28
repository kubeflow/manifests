# oauth2-proxy

## `oauth2-proxy` Deployment

This deployment of `oauth2-proxy` has been configured to align closely with the official
`oauth2-proxy` Helm installation. This approach facilitates easier integration with any
existing `oauth2-proxy` deployments that may already be present on the cluster.

### Upgrading `oauth2-proxy`

The `oauth2-proxy` component is designed for easy upgrading, thanks to its foundation on the
official `oauth2-proxy` Helm chart. The use of the standard Helm chart simplifies the upgrade
process, closely following the upgrades of the official `oauth2-proxy` releases.

### Stateless Nature of `oauth2-proxy`

`oauth2-proxy` operates as a stateless application. This statelessness simplifies many
aspects of its operation, particularly upgrades, as there are no concerns about complex state
management or data migration. Additionally, while `oauth2-proxy` is integrated into the
Kubernetes environment, this integration is limited to running the application, thereby
minimizing the impact on Kubernetes infrastructure during upgrades.

These characteristics make the upgrade process for `oauth2-proxy` more predictable and
manageable in Kubernetes environments.
