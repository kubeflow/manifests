# Configuration for installing KCC in the management cluster.

Configs are a copy of the CNRM install (see [docs](https://cloud.google.com/config-connector/docs/how-to/install-upgrade-uninstall#namespaced-mode))

To update:

1. Download the the latest GCS install bundle listed on (https://cloud.google.com/config-connector/docs/how-to/install-upgrade-uninstall#namespaced-mode)

1. Copy the system components for the namespaced install bundle to `install-system`
1. Copy the per namespace components to the template stored in the blueprint repo.

   * You will need to add kpt setters to the per namespace components.