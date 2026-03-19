# Kustomize v4 → v5

This project requires Kustomize v5. The following migration patterns must be followed when maintaining or creating new components to avoid deprecation warnings.

## `commonLabels` to `labels`
The `commonLabels` field is deprecated. Use `labels` instead. To prevent breaking immutable label selectors (like `matchLabels` in a Deployment), you should configure `includeSelectors: false` and `includeTemplates: true`.

**Before (v4):**
```yaml
commonLabels:
  app: my-app
```

**After (v5):**
```yaml
labels:
- includeSelectors: false
  includeTemplates: true
  pairs:
    app: my-app
```

## `bases` to `resources`
The `bases` field is deprecated. Simply rename it to `resources`.

**Before (v4):**
```yaml
bases:
- ../../base
```

**After (v5):**
```yaml
resources:
- ../../base
```

## `patchesStrategicMerge` / `patchesJson6902` to `patches`
Both of the legacy patch formats have been consolidated under the generic `patches` field. 

**Before (v4):**
```yaml
patchesStrategicMerge:
- deployment-patch.yaml
```

**After (v5):**
```yaml
patches:
- path: deployment-patch.yaml
```

## `vars` to `replacements`
The `vars` and `varReference` mechanisms are fundamentally deprecated. Use `replacements` to directly substitute string values dynamically.

**Before (v4):**
```yaml
vars:
- name: MY_VAR
  objref:
    kind: Service
    name: service
  fieldref:
    fieldpath: metadata.name
```

**After (v5):**
```yaml
replacements:
- source:
    kind: Service
    name: service
    fieldPath: metadata.name
  targets:
  - select:
      kind: Certificate
      name: cert
    fieldPaths:
    - spec.dnsNames.0
```
