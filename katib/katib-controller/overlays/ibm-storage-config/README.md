# ibm-storage overlay
This overlay contains the `../db` overlay as well. This duplication is necessary
since kustomize prevents basing this overlay ontop of another overlay that isn't
within this path.
