# Development

## Presubmit checks

We use a GitHub Action to run checks on every PR.

```yaml
name: Presubmit Tests
on: pull_request

jobs:
  test:
    runs-on: ubuntu-latest
    steps:

    - name: Step 1 - Checkout Repo
      uses: actions/checkout@v2
    
    - name: Step 2 - Run unit-tests
      run: |
        cd tests
        make generate-changed-only
        make test
```

It's a workflow that replicates the developer workflow on a Ubuntu runner (VM) provided by GitHub.
