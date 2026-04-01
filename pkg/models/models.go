package models

import (
	"time"
)

type Severity string

const (
	Critical Severity = "CRITICAL"
	Warning  Severity = "WARNING"
	Info     Severity = "INFO"
)

type Finding struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	File           string   `json:"file"`
	Location       string   `json:"location"`
	Issue          string   `json:"issue"`
	Impact         string   `json:"impact"`
	Recommendation string   `json:"recommendation"`
	Severity       Severity `json:"severity"`
}

type Report struct {
	Date            time.Time `json:"date"`
	FilesEvaluated  []string  `json:"files_evaluated"`
	Environments    []string  `json:"environments"`
	Summary         map[Severity]int `json:"summary"`
	Findings        []Finding `json:"findings"`
	ComparisonTable string    `json:"comparison_table,omitempty"`
}

// Partial Fleet models for parsing
type GitRepo struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Spec struct {
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
	} `yaml:"spec"`
}

type FleetYaml struct {
	DefaultNamespace string `yaml:"defaultNamespace"`
	Namespace        string `yaml:"namespace"`
	Helm             *Helm   `yaml:"helm"`
	TargetCustomizations []TargetCustomization `yaml:"targetCustomizations"`
	DependsOn        []Dependency `yaml:"dependsOn"`
}

type Helm struct {
	ReleaseName  string                 `yaml:"releaseName"`
	Chart        string                 `yaml:"chart"`
	Version      string                 `yaml:"version"`
	Repo         string                 `yaml:"repo"`
	Values       map[string]interface{} `yaml:"values"`
	ValuesFiles  []string               `yaml:"valuesFiles"`
	Atomic       bool                   `yaml:"atomic"`
	WaitForJobs  bool                   `yaml:"waitForJobs"`
}

type TargetCustomization struct {
	Name            string      `yaml:"name"`
	ClusterSelector interface{} `yaml:"clusterSelector"`
	ClusterGroup    string      `yaml:"clusterGroup"`
	ClusterName     string      `yaml:"clusterName"`
	Helm            *Helm       `yaml:"helm"`
}

type Dependency struct {
	Name string `yaml:"name"`
}
