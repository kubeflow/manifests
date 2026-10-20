#!/usr/bin/env python3

import os
import subprocess
import tempfile
import unittest
from pathlib import Path

REPOSITORY_ROOT = Path(__file__).resolve().parents[1]
LIBRARY_PATH = REPOSITORY_ROOT / "scripts/library.sh"


class HelmSynchronizationLibraryTest(unittest.TestCase):
    def test_helpers_use_explicit_temporary_file_suffixes(self):
        library_text = LIBRARY_PATH.read_text()

        self.assertNotIn(".tmp.XXXXXX", library_text)
        self.assertIn(".temporary.XXXXXX", library_text)

    def update_application_version(self, application_version, chart_text=None):
        if chart_text is None:
            chart_text = "apiVersion: v2\nappVersion: v1.0.0\n"

        with tempfile.TemporaryDirectory() as temporary_directory:
            chart_path = Path(temporary_directory) / "Chart.yaml"
            chart_path.write_text(chart_text)
            result = subprocess.run(
                [
                    "bash",
                    "-c",
                    'source "$1"; update_helm_chart_application_version "$2" "$3"',
                    "bash",
                    str(LIBRARY_PATH),
                    str(chart_path),
                    application_version,
                ],
                capture_output=True,
                text=True,
            )
            return result, chart_path.read_text()

    def test_supported_versions_and_commit_shas_are_quoted_exactly(self):
        original_chart = (
            "apiVersion: v2\n"
            "name: test-chart\n"
            "appVersion: v1.0.0\n"
            "annotations:\n"
            "  category: Test\n"
        )
        supported_versions = [
            "v7.15.2",
            "1.2.3-rc.1+build.5",
            "1234567",
            "a" * 40,
        ]

        for application_version in supported_versions:
            with self.subTest(application_version=application_version):
                result, chart_text = self.update_application_version(
                    application_version,
                    chart_text=original_chart,
                )
                expected_chart = original_chart.replace(
                    "appVersion: v1.0.0",
                    f'appVersion: "{application_version}"',
                )

                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(chart_text, expected_chart)

    def test_invalid_references_are_rejected_without_modifying_chart(self):
        original_chart = "apiVersion: v2\nappVersion: v1.0.0\n"
        invalid_references = [
            "main",
            "123456",
            "a" * 41,
            "12345g7",
            "01.2.3",
            "v1.2.3\nname: injected",
        ]

        for application_version in invalid_references:
            with self.subTest(application_version=application_version):
                result, chart_text = self.update_application_version(
                    application_version,
                    chart_text=original_chart,
                )

                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(chart_text, original_chart)

    def test_duplicate_application_version_fields_fail_atomically(self):
        original_chart = "apiVersion: v2\nappVersion: v1.0.0\nappVersion: v1.0.1\n"

        result, chart_text = self.update_application_version(
            "v7.15.2",
            chart_text=original_chart,
        )

        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(chart_text, original_chart)

    def run_atomic_writer(self, producer_body, original_content="original\n"):
        with tempfile.TemporaryDirectory() as temporary_directory:
            output_path = Path(temporary_directory) / "generated.yaml"
            output_path.write_text(original_content)
            output_path.chmod(0o640)
            result = subprocess.run(
                [
                    "bash",
                    "-c",
                    (
                        'source "$1"; '
                        "produce_output() { "
                        f"{producer_body}"
                        "; }; "
                        'write_generated_file_atomically "$2" produce_output'
                    ),
                    "bash",
                    str(LIBRARY_PATH),
                    str(output_path),
                ],
                capture_output=True,
                text=True,
            )
            temporary_files = list(
                Path(temporary_directory).glob("generated.yaml.temporary.*")
            )
            return (
                result,
                output_path.read_text(),
                output_path.stat().st_mode & 0o777,
                temporary_files,
            )

    def test_atomic_writer_replaces_file_after_successful_generation(self):
        result, output, mode, temporary_files = self.run_atomic_writer(
            "printf 'updated\\n'"
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(output, "updated\n")
        self.assertEqual(mode, 0o640)
        self.assertEqual(temporary_files, [])

    def test_atomic_writer_preserves_file_when_generation_fails(self):
        result, output, mode, temporary_files = self.run_atomic_writer(
            "printf 'partial\\n'; return 42"
        )

        self.assertEqual(result.returncode, 42)
        self.assertEqual(output, "original\n")
        self.assertEqual(mode, 0o640)
        self.assertEqual(temporary_files, [])


class SynchronizationBranchTest(unittest.TestCase):
    def setUp(self):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        self.directory = Path(directory.name)
        self.git("init", "--initial-branch=original")
        for key, value in (
            ("user.name", "Synchronization Test"),
            ("user.email", "synchronization@example.invalid"),
            ("commit.gpgSign", "false"),
            ("core.hooksPath", "/dev/null"),
        ):
            self.git("config", key, value)
        self.component = self.directory / "component.txt"
        self.component.write_text("original\n")
        self.git("add", "component.txt")
        self.git("commit", "-m", "Initial component")
        self.original_commit = self.git("rev-parse", "HEAD")

    def git(self, *arguments):
        return subprocess.run(
            ["git", *arguments],
            cwd=self.directory,
            capture_output=True,
            text=True,
            check=True,
        ).stdout.strip()

    def synchronize(self, branch, *, commit=False, no_commit=False):
        environment = os.environ.copy()
        environment.pop("KUBEFLOW_SYNCHRONIZE_NO_COMMIT", None)
        if no_commit:
            environment["KUBEFLOW_SYNCHRONIZE_NO_COMMIT"] = "true"
        command = 'source "$1"; setup_error_handling; create_branch "$2"'
        if commit:
            command += (
                "; printf 'updated\\n' > component.txt; "
                'commit_changes "$PWD" "Synchronize component" component.txt'
            )
        return subprocess.run(
            ["bash", "-c", command, "bash", str(LIBRARY_PATH), branch],
            cwd=self.directory,
            env=environment,
            capture_output=True,
            text=True,
        )

    def test_new_branch_receives_synchronization_commit(self):
        result = self.synchronize("synchronization/new-target", commit=True)

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(
            self.git("branch", "--show-current"), "synchronization/new-target"
        )
        self.assertEqual(self.git("rev-parse", "HEAD^"), self.original_commit)
        self.assertEqual(self.git("rev-parse", "original"), self.original_commit)

    def test_existing_branch_receives_commit_and_preserves_its_history(self):
        self.git("checkout", "-b", "synchronization-target")
        self.component.write_text("target history\n")
        self.git("commit", "-am", "Existing target history")
        target_commit = self.git("rev-parse", "HEAD")
        self.git("checkout", "original")

        result = self.synchronize("synchronization-target", commit=True)

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.git("branch", "--show-current"), "synchronization-target")
        self.assertEqual(self.git("rev-parse", "HEAD^"), target_commit)
        self.assertEqual(self.git("rev-parse", "original"), self.original_commit)

    def test_current_target_branch_is_preserved(self):
        result = self.synchronize("original")

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.git("branch", "--show-current"), "original")
        self.assertEqual(self.git("rev-parse", "HEAD"), self.original_commit)

    def test_no_commit_mode_preserves_branch_and_references(self):
        self.git("branch", "existing-target")
        for branch in ("existing-target", "new-target"):
            with self.subTest(branch=branch):
                references = self.git("show-ref")

                result = self.synchronize(branch, no_commit=True)

                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(self.git("branch", "--show-current"), "original")
                self.assertEqual(self.git("show-ref"), references)

    def test_checkout_conflict_stops_before_writing_or_committing(self):
        self.git("checkout", "-b", "synchronization-target")
        self.component.write_text("target history\n")
        self.git("commit", "-am", "Existing target history")
        target_commit = self.git("rev-parse", "HEAD")
        self.git("checkout", "original")
        self.component.write_text("local changes\n")

        result = self.synchronize("synchronization-target", commit=True)

        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.git("branch", "--show-current"), "original")
        self.assertEqual(self.git("rev-parse", "original"), self.original_commit)
        self.assertEqual(self.git("rev-parse", "synchronization-target"), target_commit)
        self.assertEqual(self.component.read_text(), "local changes\n")


if __name__ == "__main__":
    unittest.main()
