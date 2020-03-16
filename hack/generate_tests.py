#!/usr/bin/python
"""Regenerate tests for only the files that have changed."""

import argparse
import jinja2
import logging
import os
import shutil
import subprocess

TOP_LEVEL_EXCLUDES = ["docs", "hack", "kfdef", "stacks", "tests"]

# The subdirectory to story the expected manifests in
# We use a subdirectory of test_data because we could potentially
# have more than one version of a manifest.
KUSTOMIZE_OUTPUT_DIR = "test_data/expected"

def generate_test_name(repo_root, package_dir):
  """Generate the name of the go file to write the test to.

  Args:
    repo_root: Root of the repository
    package_dir: Full path to the kustomize directory.
  """

  rpath = os.path.relpath(package_dir, repo_root)

  pieces = rpath.split(os.path.sep)
  name = "-".join(pieces) + "_test.go"

  return name

def run_kustomize_build(repo_root, package_dir):
  """Run kustomize build and store the output in the test directory."""

  rpath = os.path.relpath(package_dir, repo_root)

  output_dir = os.path.join(repo_root, "tests", rpath, KUSTOMIZE_OUTPUT_DIR)

  if os.path.exists(output_dir):
    # Remove any previous version of the directory so that we ensure
    # that all files in that directory are from the new run
    # of kustomize build -o
    logging.info("Removing directory %s", output_dir)
    shutil.rmtree(output_dir)

  logging.info("Creating directory %s", output_dir)
  os.makedirs(output_dir)

  subprocess.check_call(["kustomize", "build", "--load_restrictor", "none",
                         "-o", output_dir], cwd=os.path.join(repo_root,
                                                             package_dir))

def get_changed_dirs():
  """Return a list of directories of changed kustomization packages."""
  # Generate a list of the files which have changed with respect to the upstream
  # branch

  # TODO(jlewi): Upstream doesn't seem to work in some cases. I think
  # upstream might end up referring to the
  origin = os.getenv("REMOTE_ORIGIN", "@{upstream}")
  logging.info("Using %s as remote origin; you can override using environment "
               "variable REMOTE_ORIGIN", origin)
  modified_files = subprocess.check_output(
    ["git", "diff", "--name-only", origin])

  repo_root = subprocess.check_output(["git", "rev-parse", "--show-toplevel"])
  repo_root = repo_root.decode()
  repo_root = repo_root.strip()

  modified_files = modified_files.decode().split()
  changed_dirs = set()
  for f in modified_files:
    changed_dir = os.path.dirname(os.path.join(repo_root, f))
    # Check if its a kustomization directory by looking for a kustomization
    # file
    if not os.path.exists(os.path.join(repo_root, changed_dir,
                                       "kustomization.yaml")):
      continue

    changed_dirs.add(changed_dir)

    # If the base package changes then we need to regenerate all the overlays
    # We assume that the base package is named "base".

    if os.path.basename(changed_dir) == "base":
      parent_dir = os.path.dirname(changed_dir)
      for root, _, files in os.walk(parent_dir):
        for f in files:
          if f == "kustomization.yaml":
            changed_dirs.add(root)

  return changed_dirs

def find_kustomize_dirs(root):
  """Find all kustomization directories in root"""

  changed_dirs = set()
  for top in os.listdir(root):
    if top.startswith("."):
      logging.info("Skipping directory %s", os.path.join(root, top))
      continue

    if top in TOP_LEVEL_EXCLUDES:
      continue

    for child, _, files in os.walk(os.path.join(root, top)):
      for f in files:
        if f == "kustomization.yaml":
          changed_dirs.add(child)

  return changed_dirs

def remove_unmatched_tests(repo_root, package_dirs):
  """Remove any tests that don't map to a kustomization.yaml file.

  This ensures tests don't linger if a package is deleted.
  """

  # Create a set of all the expected test names
  expected_tests = set()

  for d in package_dirs:
    expected_tests.add(generate_test_name(repo_root, d))

  tests_dir = os.path.join(repo_root, "tests")
  for name in os.listdir(tests_dir):
    if not name.endswith("_test.go"):
      continue

    # TODO(jlewi): we should rename and remove "_test.go" from the filename.
    if name in ["kusttestharness_test.go"]:
      logging.info("Ignoring %s", name)
      continue

    if not name in expected_tests:
      test_file = os.path.join(tests_dir, name)
      logging.info("Deleting test %s; doesn't map to kustomization file",
                   test_file)
      os.unlink(test_file)

if __name__ == "__main__":

  logging.basicConfig(
      level=logging.INFO,
      format=('%(levelname)s|%(asctime)s'
              '|%(pathname)s|%(lineno)d| %(message)s'),
      datefmt='%Y-%m-%dT%H:%M:%S',
  )
  logging.getLogger().setLevel(logging.INFO)

  parser = argparse.ArgumentParser()

  parser.add_argument(
      "--all",
      dest = "all_tests",
      action = "store_true",
      help="Regenerate all tests")

  parser.set_defaults(all_tests=False)

  args = parser.parse_args()

  repo_root = subprocess.check_output(["git", "rev-parse", "--show-toplevel"])
  repo_root = repo_root.decode()
  repo_root = repo_root.strip()

  # Get a list of package directories
  package_dirs = find_kustomize_dirs(repo_root)

  remove_unmatched_tests(repo_root, package_dirs)

  if args.all_tests:
    changed_dirs = package_dirs
  else:
    changed_dirs = get_changed_dirs()

  this_dir = os.path.dirname(__file__)
  loader = jinja2.FileSystemLoader(searchpath=os.path.join(
    this_dir, "templates"))
  env = jinja2.Environment(loader=loader)
  template = env.get_template("kustomize_test.go.template")
  for full_dir in changed_dirs:
    test_name = generate_test_name(repo_root, full_dir)
    logging.info("Regenerating test %s for %s ", test_name, full_dir)

    rpath = os.path.relpath(full_dir, repo_root)

    # Generate the kustomize output
    run_kustomize_build(repo_root, full_dir)
    test_path = os.path.join(repo_root, "tests", rpath, "kustomize_test.go")

    # Create the go test file.
    # TODO(jlewi): We really shouldn't need to redo this if it already
    # exists.

    # The go package name will be the final directory in the path
    package_name = os.path.basename(full_dir)
    # Go package names replace hyphens with underscores
    package_name = package_name.replace("-", "_")

    # We need to construct the path relative to the _test.go file of
    # the kustomize package. This path with consist of ".." entries repeated
    # enough times to get to the root of the repo. We then add the relative
    # path to the kustomize package.
    pieces = rpath.split(os.path.sep)

    p = [".."] * len(pieces)
    p.append("..")
    p.append(rpath)
    package_dir = os.path.join(*p)
    # The package dir is the relative path to the kustomize directory
    test_contents = template.render({"package": package_name,
                                     "package_dir":package_dir})


    logging.info("Writing file: %s", test_path)
    with open(test_path, "w") as test_file:
      test_file.write(test_contents)
