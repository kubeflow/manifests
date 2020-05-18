package tests

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/kustomize/v3/pkg/types"
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

// TestNoStatus is a test to try to ensure we don't include status in resources
// as this causes validation issues: https://github.com/kubeflow/manifests/issues/1174
func TestNoStatus(t *testing.T) {
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

		u := &unstructured.Unstructured{}

		if err = yaml.Unmarshal(data, u); err != nil {
			if match, _ := regexp.MatchString(".*Object 'Kind' is missing.*", err.Error()); match {
				t.Logf("Skipping %v; not a K8s resource;", path)
				return nil
			}
			t.Errorf("Error unmarshaling %v; error: %v", path, err)
		}

		if _, ok := u.UnstructuredContent()["status"]; ok {
			t.Errorf("Path %v; has status field", path)
		}
		return nil
	})

	if err != nil {
		t.Errorf("error walking the path %v; error: %v", rootDir, err)

	}
}
