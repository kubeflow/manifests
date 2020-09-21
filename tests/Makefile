# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
GOLANG_VERSION ?= 1.12.4
GOPATH ?= $(HOME)/go
PYTHON_BIN ?= python
export KUSTOMIZE_BIN ?= kustomize

# Comma seperated items within {} for more than one file
# EXCLUDE ?= istio-install-base_test.go

export GO111MODULE = on
export GO = go

all: test

# Reset various kpt values to default values
# TODO(jlewi): We should add a test to make sure changed values don't get checked in
# We don't run it in generate because we don't want to force all developers to install kpt
# TODO(jlewi): This ends up setting "isSet" to true for the setters in the KptFile.
# Does this prevent gcloud from automatically filling in the values zone and
# region? If so we may need to use yq to set that back to false.
gcp-reset:
	kpt cfg set ../gcp/ gcloud.core.project project-id
	kpt cfg set ../gcp/ gcloud.project.projectNumber projectNumber
	kpt cfg set ../gcp/ gcloud.compute.zone ZONE
	kpt cfg set ../gcp/ gcloud.compute.region REGION
	kpt cfg set ../gcp/ name name
	kpt cfg set ../gcp/ location location
	yq d -i ../gcp/Kptfile  openAPI.definitions["io.k8s.cli.setters.gcloud.core.project"].["x-k8s-cli"].["setter"].["isSet"]
	yq d -i ../gcp/Kptfile  openAPI.definitions["io.k8s.cli.setters.gcloud.project.projectNumber"].["x-k8s-cli"].["setter"].["isSet"]
	yq d -i ../gcp/Kptfile  openAPI.definitions["io.k8s.cli.setters.gcloud.compute.zone"].["x-k8s-cli"].["setter"].["isSet"]
	yq d -i ../gcp/Kptfile  openAPI.definitions["io.k8s.cli.setters.gcloud.compute.region"].["x-k8s-cli"].["setter"].["isSet"]
	yq d -i ../gcp/Kptfile  openAPI.definitions["io.k8s.cli.setters.location"].["x-k8s-cli"].["setter"].["isSet"]
	yq d -i ../gcp/Kptfile  openAPI.definitions["io.k8s.cli.setters.name"].["x-k8s-cli"].["setter"].["isSet"]

generate:
	$(PYTHON_BIN) ../hack/generate_tests.py --all
	$(GO) fmt ./...

generate-changed-only:
	$(PYTHON_BIN) ../hack/generate_tests.py
	$(GO) fmt ./...

modules:
	@GO111MODULE=on $(GO) mod download

test: modules
	@GO111MODULE=on $(GO) test -v github.com/kubeflow/manifests/tests/...

run-unittest-plugin:
	cd .. && XDG_CONFIG_HOME=$$(pwd)/plugins kustomize build profiles/overlays/test --enable_alpha_plugins

run-application-plugin:
	cd .. && XDG_CONFIG_HOME=$$(pwd)/plugins kustomize build jupyter/jupyter-web-app/overlays/application --enable_alpha_plugins
