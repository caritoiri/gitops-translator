package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gitops-values-translator-go/translator"
)

func main() {
	workspaceDir := "/home/laborant/workspace"
	srcDir := filepath.Join(workspaceDir, "gitops/infra")

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Look for values-*.yaml files in gitops/infra/
		filename := info.Name()
		if !strings.HasPrefix(filename, "values-") || (!strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml")) {
			return nil
		}

		// Extract environment name from filename (e.g. values-develop.yaml -> develop)
		envName := strings.TrimSuffix(strings.TrimPrefix(filename, "values-"), ".yaml")
		envName = strings.TrimSuffix(envName, ".yml")

		// Standardize envName to match destination structures
		envLower := strings.ToLower(envName)
		if strings.Contains(envLower, "dev") {
			envName = "develop"
		} else if strings.Contains(envLower, "stg") || strings.Contains(envLower, "staging") {
			envName = "staging"
		} else if strings.Contains(envLower, "release") || strings.Contains(envLower, "prep") {
			envName = "release"
		} else if strings.Contains(envLower, "prod") || strings.Contains(envLower, "master") {
			envName = "production"
		}

		// Read legacy yaml
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Determine team and app from path
		// e.g. gitops/infra/bg-crm/api-rest-integracion-crm/values-develop.yaml
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) < 3 {
			// Skip files not in deep nested directories (like base folders)
			return nil
		}
		team := parts[0]

		// Set default cluster based on team/env rules matching the pipeline
		cluster := "on-premise"
		if team == "ganamovil" {
			cluster = "on-prem-ngm"
		} else if envName != "production" && team == "middleware" {
			cluster = "on-prem-middl"
		}

		// Translate values and get Argo CD app manifest
		translatedValues, targetValuesPath, argoApp, targetArgoPath, _, err := translator.TranslateValuesWithArgo(
			string(data),
			cluster,
			envName,
			team,
			"",
			false,
		)
		if err != nil {
			fmt.Printf("⚠️ Skip %s due to error: %v\n", path, err)
			return nil
		}

		// 1. Write the modern values file to charts repo
		fullValuesDestPath := filepath.Join(workspaceDir, targetValuesPath)
		valuesDir := filepath.Dir(fullValuesDestPath)
		if err := os.MkdirAll(valuesDir, 0755); err != nil {
			return fmt.Errorf("failed to create values dir %s: %w", valuesDir, err)
		}
		if err := ioutil.WriteFile(fullValuesDestPath, []byte(translatedValues), 0644); err != nil {
			return fmt.Errorf("failed to write values file %s: %w", fullValuesDestPath, err)
		}

		// 2. Write the modern ArgoCD Application manifest to argocd repo
		fullArgoDestPath := filepath.Join(workspaceDir, "argocd", targetArgoPath)
		argoDir := filepath.Dir(fullArgoDestPath)
		if err := os.MkdirAll(argoDir, 0755); err != nil {
			return fmt.Errorf("failed to create argo dir %s: %w", argoDir, err)
		}
		if err := ioutil.WriteFile(fullArgoDestPath, []byte(argoApp), 0644); err != nil {
			return fmt.Errorf("failed to write argo file %s: %w", fullArgoDestPath, err)
		}

		fmt.Printf("✅ Migrated: %s\n   -> Values: %s\n   -> ArgoCD: %s\n", path, fullValuesDestPath, fullArgoDestPath)
		return nil
	})

	if err != nil {
		fmt.Printf("❌ Migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\n🎉 Bulk migration applied successfully to all legacy configurations!")
}
