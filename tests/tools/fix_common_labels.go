// fix_common_labels.go is a run once program to fix the commonLabels in Kubeflow
// kustomization files. Per https://github.com/kubeflow/manifests/issues/1131 commonLabels should be immutable.
// Our commonLabels violated this because we were including version.
// This is a simple program to remove those from commonLabels.
package main

import (
	"github.com/ghodss/yaml"
	"github.com/google/martian/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

const (
	VersionLabel      = "app.kubernetes.io/version"
	ManagedByLabel    = "app.kubernetes.io/managed-by"
	InstanceLabel     = "app.kubernetes.io/instance"
	PartOfLabel       = "app.kubernetes.io/part-of"
	KustomizationFile = "kustomization.yaml"
)

// readKustomization will read a kustomization.yaml and return the kustomize object
func readKustomization(path string) (*types.Kustomization, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	def := &types.Kustomization{}
	if err = yaml.Unmarshal(data, def); err != nil {
		return nil, err
	}
	return def, nil
}

func writeKustomization(path string, k *types.Kustomization) error {
	yaml, err := yaml.Marshal(k)

	if err != nil {
		return errors.WithStack(errors.Wrapf(err, "Error trying to marshal kustomization for path: %v", path))
	}

	kustomizationFileErr := ioutil.WriteFile(path, yaml, 0644)
	if kustomizationFileErr != nil {
		return errors.WithStack(errors.Wrapf(kustomizationFileErr, "Error writing file: %v", path))
	}
	return nil
}

func main() {
	rootDir := ".."

	// Directories to exclude. Thee paths should be relative to rootDir.
	// Subdirectories won't be searched
	excludes := map[string]bool{
		"tests":   true,
		".git":    true,
		".github": true,
	}

	// These labels are likely to be mutable and should not be part of commonLabels
	labelsToRemove := []string{VersionLabel, ManagedByLabel, InstanceLabel, PartOfLabel}

	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		relPath, err := filepath.Rel(rootDir, path)

		if err != nil {
			log.Errorf("Could not compute relative path(%v, %v); error: %v", rootDir, path, err)
			return err
		}

		if _, ok := excludes[relPath]; ok {
			log.Infof("Skipping directory %v", path)
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
			log.Errorf("Error reading file: %v; error: %v", path, err)
			return nil
		}

		if k.CommonLabels == nil {
			return nil
		}

		modified := false

		for _, l := range labelsToRemove {
			if _, ok := k.CommonLabels[l]; ok {
				delete(k.CommonLabels, l)
				modified = true
			}
		}

		if !modified {
			return nil
		}
		log.Infof("Writing updated kustomization: %v", path)

		if err := writeKustomization(path, k); err != nil {
			log.Errorf("Error writing %v; error: %v", path, err)
		}
		return nil
	})

	if err != nil {
		log.Errorf("error walking the path %v; error: %v", rootDir, err)

	}
}
