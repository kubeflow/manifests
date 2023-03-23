#!/bin/bash

set -eux
sudo apt update

# Install prereq packages
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common tar

# Add GPG key for the official Docker repository
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

# Add the Docker repository APT sources
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu focal stable"

# Update to install from Docker repo
apt-cache policy docker-ce

# Install Docker
sudo apt install -y docker-ce

# Verify Docker is running
sudo systemctl status docker

# Add user to the docker group
sudo usermod -a -G docker ubuntu