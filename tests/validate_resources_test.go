package tests

import (
	"bytes"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"strings"
	"testing"
)

const (
	VersionLabel      = "app.kubernetes.io/version"
	InstanceLabel     = "app.kubernetes.io/instance"
	ManagedByLabel    = "app.kubernetes.io/managed-by"
	PartOfLabel       = "app.kubernetes.io/part-of"
	KustomizationFile = "kustomization.yaml"
)

// readKustomization will read a kustomization.yaml and return the kustomize object
func readKustomization(kfDefFile string) (*types.Kustomization, error) {
	data, err := ioutil.ReadFile(kfDefFile)
	if err != nil {
		return nil, err
	}
	def := &types.Kustomization{}
	if err = yaml.Unmarshal(data, def); err != nil {
		return nil, err
	}
	return def, nil
}

// TestCommonLabelsImmutable is a test to try to ensure we don't have mutable labels which will
// cause problems on upgrades per https://github.com/kubeflow/manifests/issues/1131.
func TestCommonLabelsImmutable(t *testing.T) {
	rootDir := ".."

	// Directories to exclude. Thee paths should be relative to rootDir.
	// Subdirectories won't be searched
	excludes := map[string]bool{
		"tests":   true,
		".git":    true,
		".github": true,
	}

	// These labels are likely to be mutable and should not be part of commonLabels
	forbiddenLabels := []string{VersionLabel, ManagedByLabel, InstanceLabel, PartOfLabel}

	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		relPath, err := filepath.Rel(rootDir, path)

		if err != nil {
			t.Fatalf("Could not compute relative path(%v, %v); error: %v", rootDir, path, err)
		}

		if _, ok := excludes[relPath]; ok {
			t.Logf("Skipping directory %v", path)
			return filepath.SkipDir
		}

		// skip directories
		if info.IsDir() {
			return nil
		}

		if info.Name() != KustomizationFile {
			return nil
		}

		k, err := readKustomization(path)

		if err != nil {
			t.Errorf("Error reading file: %v; error: %v", path, err)
			return nil
		}

		if k.CommonLabels == nil {
			return nil
		}

		for _, l := range forbiddenLabels {
			if _, ok := k.CommonLabels[l]; ok {
				t.Errorf("%v has forbidden commonLabel %v", path, l)
			}
		}
		return nil
	})

	if err != nil {
		t.Errorf("error walking the path %v; error: %v", rootDir, err)

	}
}

// TestValidK8sResources reads all the K8s resources and performs a bunch of validation checks.
//
// Currently the following checks are performed:
//  i) ensure we don't include status in resources
//     as this causes validation issues: https://github.com/kubeflow/manifests/issues/1174
//
//  ii) ensure that if annotations are present it is not empty.
//  Having empty annotations https://github.com/GoogleContainerTools/kpt/issues/541 causes problems for kpt and
//  ACM.  Offending YAML looks like
//      metadata:
// 		 name: kf-admin-iap
//       annotations:
//      rules:
//        ...
func TestValidK8sResources(t *testing.T) {
	rootDir := ".."

	// Directories to exclude. Thee paths should be relative to rootDir.
	// Subdirectories won't be searched
	excludes := map[string]bool{
		"tests":                  true,
		".git":                   true,
		".github":                true,
		"profiles/overlays/test": true,
		// Skip cnrm-install. We don't install this with ACM so we don't need to fix it.
		// It seems like if this is an issue it should eventually get fixed in upstream cnrm configs.
		// The CNRM directory has lots of CRDS with non empty status.
		"gcp/v2/management/cnrm-install": true,
	}

	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		relPath, err := filepath.Rel(rootDir, path)

		if err != nil {
			t.Fatalf("Could not compute relative path(%v, %v); error: %v", rootDir, path, err)
		}

		if _, ok := excludes[relPath]; ok {
			return filepath.SkipDir
		}

		// skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non YAML files
		ext := filepath.Ext(info.Name())

		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		data, err := ioutil.ReadFile(path)

		if err != nil {
			t.Errorf("Error reading %v; error: %v", path, err)
		}

		input := bytes.NewReader(data)
		reader := kio.ByteReader{
			Reader: input,
			// We need to disable adding reader annotations because
			// we want to run some checks about whether annotations are set and
			// adding those annotations interferes with that.
			OmitReaderAnnotations: true,
		}

		nodes, err := reader.Read()

		if err != nil {
			t.Errorf("Error unmarshaling %v; error: %v", path, err)
		}

		for _, n := range nodes {
			//root := n

			m, err := n.GetMeta()
			// Skip objects with no metadata
			if err != nil {
				continue
			}

			// Skip Kustomization
			if strings.ToLower(m.Kind) == "kustomization" {
				continue
			}
			if m.Name == "" || m.Kind == "" {
				continue
			}

			// Ensure status isn't set
			f := n.Field("status")

			if !kyaml.IsFieldEmpty(f) {
				t.Errorf("Path %v; resource %v; has status field", path, m.Name)
			}

			metadata := n.Field("metadata")

			checkEmptyAnnotations := func() {
				annotations := metadata.Value.Field("annotations")

				if annotations == nil {
					return
				}

				if kyaml.IsFieldEmpty(annotations) {
					t.Errorf("Path %v; resource %v; has empty annotations; if no annotations are present the field shouldn't be present", path, m.Name)
					return
				}
			}

			checkEmptyAnnotations()
		}
		return nil
	})

	if err != nil {
		t.Errorf("error walking the path %v; error: %v", rootDir, err)
	}
}
