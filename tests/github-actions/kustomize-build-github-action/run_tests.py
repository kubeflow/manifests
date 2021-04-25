#!/usr/bin/env python3

import os
import sys
import pathlib
import logging
import argparse
from fnmatch import fnmatch
from contextlib import suppress
from subprocess import check_output as yld, DEVNULL

from common import YAML

log = logging.getLogger(__name__)


def build_kustomization(path):
    # XXX: We use the following extra flags for kustomize:
    # `load_restrictor`: Allow loading files outside of the root folder of the
    # kustomization.
    return yld(["kustomize", "build", "--load_restrictor", "none", path],
               stderr=DEVNULL)


def find_kustomizations(path, overlay=None, include_patterns=["*"],
                        exclude_patterns=[]):
    """Find all kustomization files in given directory."""
    kustomizations = []
    for root, _, files in os.walk(path):
        if "kustomization.yaml" in files:
            kname = os.path.basename(root)
            if overlay and kname != overlay:
                continue
            included = False
            for include_pattern in include_patterns:
                if fnmatch(pathlib.PurePath(root).relative_to(path), include_pattern):
                    included = True
            if not included:
                continue
            for exclude_pattern in exclude_patterns:
                if fnmatch(pathlib.PurePath(root).relative_to(path), exclude_pattern):
                    continue
            kustomizations.append(root)

    return kustomizations


def parse_args():
    parser = argparse.ArgumentParser()
    default_root_path = "."
    with suppress(Exception):
        repo_path = yld(["git", "rev-parse", "--show-toplevel"])
        default_root_path = repo_path.decode("utf-8").strip()

    parser.add_argument("--root-path", type=str, default=default_root_path,
                        help="Root path to search for kustomizations.")
    parser.add_argument("--include-patterns", default="*",
                        help="Comma-separate list of path patterns to include.")
    parser.add_argument("--exclude-patterns", default="",
                        help="Comma-separate list of path patterns to exclude.")
    return parser.parse_args()


def main():
    # Detect kustomizations
    logging.basicConfig(level=logging.INFO)
    args = parse_args()
    include_patterns = [p for p in args.include_patterns.split(",") if p]
    exclude_patterns = [p for p in args.exclude_patterns.split(",") if p]
    kustomizations = find_kustomizations(
        args.root_path, include_patterns=include_patterns,
        exclude_patterns=exclude_patterns)

    errors = []
    for kust in kustomizations:
        try:
            build_kustomization(kust)
            log.info("Successfully built kustomization `%s`",
                     pathlib.PurePath(kust).relative_to(args.root_path))
        except Exception as e:
            errors.append(e)
            log.error("Failed to build kustomization `%s`: %s",
                      pathlib.PurePath(kust).relative_to(args.root_path), e)
            continue
    if errors:
        sys.exit(1)


if __name__ == "__main__":
    sys.exit(main())
