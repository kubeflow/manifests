#!/bin/bash
set -e

echo "Running Dex login test..."

# Install Python requirements
pip3 install -q requests

# Run the test
python3 tests/gh-actions/test_dex_login.py

echo "Dex login test completed successfully." 