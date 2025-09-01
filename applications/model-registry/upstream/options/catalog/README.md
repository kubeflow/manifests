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
- **`excludedModels`** (*string list*, optional): A list of models to exclude from the catalog. These can be an exact name with a tag (e.g., `model-a:1.0`) or a pattern ending with `*` to exclude all tags for a repository (e.g., `model-b:*`).

##### Example

```yaml
catalogs:
  - name: Sample Catalog
    id: sample_custom_catalog
    type: yaml
    enabled: true
    properties:
      yamlCatalogPath: sample-catalog.yaml
      excludedModels:
      - model-a:1.0
      - model-b:*
```

#### `rhec`

The `rhec` type sources model metadata from the Red Hat Ecosystem Catalog.

##### Properties

- **`models`** (*string list*, required): A list of models to include from the Red Hat Ecosystem Catalog. Each entry contains the full name of the model repository in the Red Hat Ecosystem Catalog (e.g., `rhelai1/modelcar-granite-7b-starter`).
- **`excludedModels`** (*string list*, optional): A list of models to exclude from the catalog. These can be an exact name with a tag (e.g., `rhelai1/modelcar-granite-7b-starter:b9514c3`) or a pattern ending with `*` to exclude all tags for a repository (e.g., `rhelai1/modelcar-granite-7b-starter:*`).

##### Example

```yaml
catalogs:
  - name: Red Hat Ecosystem Catalog
    id: sample_rhec_catalog
    type: rhec
    enabled: true
    properties:
      models:
      - rhelai1/modelcar-granite-7b-starter
      excludedModels:
      - rhelai1/modelcar-granite-7b-starter:v0
      - rhelai1/modelcar-granite-*
```