package analyzer

import (
	"fleet-evaluator/pkg/models"
	"testing"
)

func TestAnalyzeGitRepo(t *testing.T) {
	gr := models.GitRepo{
		Kind: "GitRepo",
		Spec: struct {
			Repo                  string   `yaml:"repo"`
			Branch                string   `yaml:"branch"`
			Revision              string   `yaml:"revision"`
			Paths                 []string `yaml:"paths"`
			PollingInterval       string   `yaml:"pollingInterval"`
			InsecureSkipTLSVerify bool     `yaml:"insecureSkipTLSVerify"`
			ClientSecretName      string   `yaml:"clientSecretName"`
			ServiceAccount        string   `yaml:"serviceAccount"`
			Targets               []struct {
				ClusterSelector interface{} `yaml:"clusterSelector"`
				ClusterGroup    string      `yaml:"clusterGroup"`
				ClusterName     string      `yaml:"clusterName"`
			} `yaml:"targets"`
			CorrectDrift struct {
				Enabled bool `yaml:"enabled"`
				Force   bool `yaml:"force"`
			} `yaml:"correctDrift"`
		}{
			Paths: []string{"/"},
		},
	}

	findings := analyzeGitRepo("test.yaml", gr)
	foundC1 := false
	for _, f := range findings {
		if f.ID == "C1" {
			foundC1 = true
			break
		}
	}

	if !foundC1 {
		t.Errorf("Expected to find C1 (deploying from repo root), but didn't")
	}
}
