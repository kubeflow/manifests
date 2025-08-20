# Model Catalog Manifests

To deploy the model catalog:

```sh
kubectl apply -k . -n NAMESPACE
```

Replace `NAMESPACE` with your desired Kubernetes namespace.

## sources.yaml Configuration

The `sources.yaml` file configures the model catalog sources. It contains a top-level `catalogs` list, where each entry defines a single catalog source.

### Common Properties

Each catalog source entry supports the following common properties:

- **`name`** (*string*, required): A user-friendly name for the catalog source.
- **`id`** (*string*, required): A unique identifier for the catalog source.
- **`type`** (*string*, required): The type of catalog source. Supported values are `yaml` and `rhec`.
- **`enabled`** (*boolean*, optional): Whether the catalog source is enabled. Defaults to `true` if not specified.

### Catalog Source Types

Below are the supported catalog source types and their specific `properties`.

#### `yaml`

The `yaml` type sources model metadata from a local YAML file.

##### Properties

- **`yamlCatalogPath`** (*string*, required): The path to the YAML file containing the model definitions. This path is relative to the directory where the `sources.yaml` file is located.

##### Example

```yaml
catalogs:
  - name: Sample Catalog
    id: sample_custom_catalog
    type: yaml
    enabled: true
    properties:
      yamlCatalogPath: sample-catalog.yaml
```

#### `rhec`

The `rhec` type sources model metadata from the Red Hat Ecosystem Catalog.

##### Properties

- **`models`** (*list*, required): A list of models to include from the Red Hat Ecosystem Catalog. Each entry in the list must contain a `repository` field.
  - **`repository`** (*string*, required): The name of the model repository in the Red Hat Ecosystem Catalog (e.g., `rhelai1/modelcar-granite-7b-starter`).

##### Example

```yaml
catalogs:
  - name: Red Hat Ecosystem Catalog
    id: sample_rhec_catalog
    type: rhec
    enabled: true
    properties:
      models:
      - repository: rhelai1/modelcar-granite-7b-starter
```
