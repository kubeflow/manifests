#!/usr/bin/env python3
"""Exercise the shared UI fixture guard without a Kubernetes cluster."""

import copy
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPT = Path(
    os.environ.get(
        "MODEL_REGISTRY_EGRESS_SCRIPT", ROOT / "tests/model_registry_egress_test.sh"
    )
)
BASH = os.environ.get("BASH_BINARY", "bash")
IMAGE = "ghcr.io/kubeflow/hub/ui:v0.3.18"
DEPLOYMENT = {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {"name": "model-registry-ui", "namespace": "kubeflow"},
    "spec": {
        "template": {
            "spec": {
                "serviceAccountName": "model-registry-ui",
                "containers": [
                    {
                        "name": "model-registry-ui",
                        "image": IMAGE,
                        "args": ["--deployment-mode=kubeflow", "--port=8080"],
                    }
                ],
            }
        }
    },
}


class SharedRegistryImageTest(unittest.TestCase):
    def run_guard(
        self, expected_image=IMAGE, actual_image=None, change=None, fixture_change=None
    ):
        with tempfile.TemporaryDirectory(prefix="registry-image-test-") as directory:
            directory = Path(directory)
            intended = copy.deepcopy(DEPLOYMENT)
            intended["spec"]["template"]["spec"]["containers"][0][
                "image"
            ] = expected_image
            deployed = copy.deepcopy(intended)
            deployed["spec"]["template"]["spec"]["containers"][0]["image"] = (
                actual_image or expected_image
            )
            if change:
                change(deployed["spec"]["template"]["spec"])
            # A List with an unrelated first object prevents accidental reliance
            # on the renderer's resource ordering.
            fixture = {
                "apiVersion": "v1",
                "kind": "List",
                "items": [
                    {
                        "apiVersion": "v1",
                        "kind": "ServiceAccount",
                        "metadata": {"name": "model-registry-ui"},
                    },
                    intended,
                ],
            }
            if fixture_change:
                fixture_change(fixture)
            (directory / "fixture.json").write_text(json.dumps(fixture))
            (directory / "deployed.json").write_text(json.dumps(deployed))
            programs = {
                "kustomize": """import os
from pathlib import Path
print((Path(os.environ['IMAGE_TEST_DIRECTORY']) / 'fixture.json').read_text())
""",
                "kubectl": """import json, os, sys
from pathlib import Path
directory = Path(os.environ['IMAGE_TEST_DIRECTORY'])
arguments = sys.argv[1:]
if arguments == ['config', 'current-context']:
    print('kind-registry-image-test')
elif arguments[0] == 'create':
    assert '--dry-run=client' in arguments
    print(Path(arguments[arguments.index('-f') + 1]).read_text())
elif arguments[0] == 'apply':
    if arguments[arguments.index('-f') + 1] == '-':
        try:
            json.loads(sys.stdin.read())
        except json.JSONDecodeError:
            (directory / 'control-reached').touch()
            sys.exit(77)
elif 'get' in arguments:
    if 'deployment' in arguments:
        print((directory / 'deployed.json').read_text())
    else:
        print(json.dumps({'metadata': {'name': 'backup'}, 'spec': {}}))
elif 'rollout' not in arguments and 'delete' not in arguments:
    raise AssertionError(arguments)
""",
            }
            for name, source in programs.items():
                executable = directory / name
                executable.write_text(
                    "#!" + Path(sys.executable).as_posix() + "\n" + source, newline="\n"
                )
                executable.chmod(0o755)
            environment = dict(
                os.environ,
                PATH=str(directory) + os.pathsep + os.environ["PATH"],
                IMAGE_TEST_DIRECTORY=str(directory),
            )
            result = subprocess.run(
                [BASH, str(SCRIPT)],
                env=environment,
                capture_output=True,
                text=True,
                timeout=15,
            )
            return result, (directory / "control-reached").exists()

    def assert_guard_accepts(self, image):
        result, reached = self.run_guard(image)
        self.assertEqual(result.returncode, 77, result.stdout + result.stderr)
        self.assertTrue(reached)

    def test_source_version_updates_are_accepted(self):
        for tag in ("v0.3.17", "v0.3.18"):
            with self.subTest(tag=tag):
                self.assert_guard_accepts("ghcr.io/kubeflow/hub/ui:" + tag)

    def test_digest_and_tag_with_digest_are_preserved(self):
        for image in (
            "ghcr.io/kubeflow/hub/ui@sha256:" + "a" * 64,
            "ghcr.io/kubeflow/hub/ui:v0.3.18@sha256:" + "b" * 64,
        ):
            with self.subTest(image=image):
                self.assert_guard_accepts(image)

    def test_genuine_image_mismatches_are_rejected(self):
        for intended, deployed in (
            (IMAGE, "ghcr.io/kubeflow/hub/ui:v0.3.17"),
            (
                "ghcr.io/kubeflow/hub/ui@sha256:" + "a" * 64,
                "ghcr.io/kubeflow/hub/ui@sha256:" + "b" * 64,
            ),
        ):
            with self.subTest(intended=intended, deployed=deployed):
                result, reached = self.run_guard(intended, deployed)
                self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
                self.assertFalse(reached)

    def test_invalid_intended_fixture_is_rejected(self):
        result, reached = self.run_guard("example.invalid/mock-ui:v0.3.18")
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        self.assertFalse(reached)
        for change in (
            lambda fixture: fixture["items"].pop(),
            lambda fixture: fixture["items"].append(copy.deepcopy(fixture["items"][1])),
            lambda fixture: fixture["items"][1]["metadata"].update(namespace="other"),
            lambda fixture: fixture["items"][1]["spec"]["template"]["spec"][
                "containers"
            ].clear(),
            lambda fixture: fixture["items"][1]["spec"]["template"]["spec"][
                "containers"
            ].append(
                copy.deepcopy(
                    fixture["items"][1]["spec"]["template"]["spec"]["containers"][0]
                )
            ),
        ):
            with self.subTest(change=change):
                result, reached = self.run_guard(fixture_change=change)
                self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
                self.assertFalse(reached)

    def test_identity_and_default_authentication_guards_remain(self):
        for change in (
            lambda pod: pod.update(serviceAccountName="default"),
            lambda pod: pod["containers"][0]["args"].append("--auth-method=user_token"),
            lambda pod: pod["containers"][0].update(
                env=[{"name": "MOCK_MR_CLIENT", "value": "true"}]
            ),
        ):
            with self.subTest(change=change):
                result, reached = self.run_guard(change=change)
                self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
                self.assertFalse(reached)


if __name__ == "__main__":
    unittest.main()
