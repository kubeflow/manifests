#!/usr/bin/env python3
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from pathlib import Path


REPOSITORY_ROOT = Path(__file__).resolve().parent.parent
KIND_INSTALL_SCRIPT = REPOSITORY_ROOT / "tests" / "install_KinD_create_KinD_cluster_install_kustomize.sh"


def test_kind_bootstrap_targets_kind_031_and_kubernetes_135():
    script = KIND_INSTALL_SCRIPT.read_text()

    assert 'KIND_VERSION="v0.31.0"' in script
    assert (
        'KIND_NODE_IMAGE="kindest/node:v1.35.0@sha256:'
        '452d707d4862f52530247495d180205e029056831160e22870e37e3f6c1ac31f"'
        in script
    )
    assert "apiVersion: kubeadm.k8s.io/v1beta3" in script
    assert script.count('image: ${KIND_NODE_IMAGE}') == 3
    assert '--name kubeflow' in script
