#!/bin/bash

set -eux

# Install python
sudo apt install python3
python3 --version

# Install pip
sudo apt install -y python3-pip