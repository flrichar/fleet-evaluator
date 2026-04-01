package analyzer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"fleet-evaluator/pkg/models"
	"gopkg.in/yaml.v3"
)

func AnalyzeFiles(files []string) (*models.Report, error) {
	report := &models.Report{
		Date:           time.Now(),
		FilesEvaluated: files,
		Summary:        make(map[models.Severity]int),
		Findings:       []models.Finding{},
	}

	gitRepos := make(map[string]models.GitRepo)
	
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		// Try to parse as GitRepo first
		var gitRepo models.GitRepo
		err = yaml.Unmarshal(data, &gitRepo)
		if err == nil && gitRepo.Kind == "GitRepo" {
			findings := analyzeGitRepo(file, gitRepo)
			report.Findings = append(report.Findings, findings...)
			gitRepos[file] = gitRepo
		} else {
			// Try to parse as fleet.yaml
			var fleetYaml models.FleetYaml
			err = yaml.Unmarshal(data, &fleetYaml)
			if err == nil && (fleetYaml.Helm != nil || len(fleetYaml.TargetCustomizations) > 0 || fleetYaml.DefaultNamespace != "") {
				findings := analyzeFleetYaml(file, fleetYaml)
				report.Findings = append(report.Findings, findings...)
			}
		}

		// Scenario-specific: detect hardcoded secrets in any YAML
		secrets := detectHardcodedSecrets(file, data)
		report.Findings = append(report.Findings, secrets...)
	}

	// Cross-file analysis
	report.Findings = append(report.Findings, analyzeCrossFileConsistency(gitRepos)...)

	for _, finding := range report.Findings {
		report.Summary[finding.Severity]++
	}

	return report, nil
}

func analyzeGitRepo(file string, gr models.GitRepo) []models.Finding {
	findings := []models.Finding{}

	isProd := strings.Contains(strings.ToLower(file), "prod") || strings.Contains(strings.ToLower(gr.Metadata.Name), "prod")

	// Path check
	for _, p := range gr.Spec.Paths {
		if p == "/" {
			findings = append(findings, models.Finding{
				ID:             "C1",
				Title:          "Deploying from repo root",
				File:           file,
				Location:       "spec.paths",
				Issue:          "GitRepo is configured to deploy from the root directory ('/')",
				Impact:         "Risks deploying unintended manifests like CI configs, docs, or test files.",
				Recommendation: "Specify explicit paths for Kubernetes manifests.",
				Severity:       models.Critical,
			})
		}
	}

	// Polling interval
	if gr.Spec.PollingInterval != "" {
		dur, err := time.ParseDuration(gr.Spec.PollingInterval)
		if err == nil && dur < 60*time.Second {
			findings = append(findings, models.Finding{
				ID:             "W1",
				Title:          "Aggressive polling interval",
				File:           file,
				Location:       "spec.pollingInterval",
				Issue:          fmt.Sprintf("Polling interval is set to %s", gr.Spec.PollingInterval),
				Impact:         "Very low intervals increase API load on Git providers and may lead to rate limiting.",
				Recommendation: "Increase polling interval to 60s or more.",
				Severity:       models.Warning,
			})
		}
	}

	// Insecure skip TLS
	if gr.Spec.InsecureSkipTLSVerify {
		findings = append(findings, models.Finding{
			ID:             "C2",
			Title:          "Insecure TLS verification",
			File:           file,
			Location:       "spec.insecureSkipTLSVerify",
			Issue:          "insecureSkipTLSVerify is set to true",
			Impact:         "Disables TLS certificate validation, making the connection vulnerable to MITM attacks.",
			Recommendation: "Use a valid CA bundle instead of skipping verification.",
			Severity:       models.Critical,
		})
	}

	// Empty targets
	if len(gr.Spec.Targets) == 0 {
		findings = append(findings, models.Finding{
			ID:             "W2",
			Title:          "Empty targets array",
			File:           file,
			Location:       "spec.targets",
			Issue:          "Targets array is empty",
			Impact:         "GitRepo will not deploy to any clusters.",
			Recommendation: "Specify cluster selectors or group names in targets.",
			Severity:       models.Warning,
		})
	}

	// serviceAccount for production
	if isProd && gr.Spec.ServiceAccount == "" {
		findings = append(findings, models.Finding{
			ID:             "W6",
			Title:          "Missing serviceAccount for production",
			File:           file,
			Location:       "spec.serviceAccount",
			Issue:          "Production GitRepo does not have a dedicated serviceAccount",
			Impact:         "Fleet will use its default service account, which may have excessive permissions.",
			Recommendation: "Use a dedicated serviceAccount with minimal RBAC for production deployments.",
			Severity:       models.Warning,
		})
	}

	// correctDrift.force
	if gr.Spec.CorrectDrift.Force {
		findings = append(findings, models.Finding{
			ID:             "W8",
			Title:          "Drift correction with force enabled",
			File:           file,
			Location:       "spec.correctDrift.force",
			Issue:          "correctDrift.force is set to true",
			Impact:         "Force-deletes and recreates resources on drift, which can be disruptive for stateful workloads or PVCs.",
			Recommendation: "Use with caution. Ensure no stateful resources are managed by this GitRepo if force is enabled.",
			Severity:       models.Warning,
		})
	}

	// Positive finding: Revision pinning
	if gr.Spec.Revision != "" {
		findings = append(findings, models.Finding{
			ID:             "I1",
			Title:          "Good practice: Revision pinning",
			File:           file,
			Location:       "spec.revision",
			Issue:          "GitRepo uses a specific revision (tag or SHA)",
			Impact:         "Ensures reproducible deployments.",
			Recommendation: "Continue using revision pinning for production.",
			Severity:       models.Info,
		})
	}

	return findings
}

func analyzeCrossFileConsistency(gitRepos map[string]models.GitRepo) []models.Finding {
	findings := []models.Finding{}
	
	labelKeys := make(map[string]string) // key -> file

	for file, gr := range gitRepos {
		for _, target := range gr.Spec.Targets {
			if target.ClusterSelector == nil {
				continue
			}
			// Extremely simplified label extraction for evaluation purposes
			if m, ok := target.ClusterSelector.(map[string]interface{}); ok {
				if matchLabels, ok := m["matchLabels"].(map[string]interface{}); ok {
					for k := range matchLabels {
						labelKeys[k] = file
					}
				}
			}
		}
	}

	// Check if both 'env' and 'environment' are used
	if _, hasEnv := labelKeys["env"]; hasEnv {
		if file, hasEnvironment := labelKeys["environment"]; hasEnvironment {
			findings = append(findings, models.Finding{
				ID:             "W7",
				Title:          "Inconsistent label keys",
				File:           file,
				Location:       "spec.targets.clusterSelector.matchLabels",
				Issue:          "Both 'env' and 'environment' label keys are used across GitRepos",
				Impact:         "Inconsistent labeling makes targeting confusing and prone to errors.",
				Recommendation: "Standardize on a single label taxonomy (e.g., always use 'env').",
				Severity:       models.Warning,
			})
		}
	}

	return findings
}

func analyzeFleetYaml(file string, fy models.FleetYaml) []models.Finding {
	findings := []models.Finding{}

	// defaultNamespace
	if fy.DefaultNamespace == "" {
		findings = append(findings, models.Finding{
			ID:             "W3",
			Title:          "Missing defaultNamespace",
			File:           file,
			Location:       "defaultNamespace",
			Issue:          "defaultNamespace is not set in fleet.yaml",
			Impact:         "Resources might deploy to the GitRepo's namespace in the management cluster by mistake.",
			Recommendation: "Always set defaultNamespace explicitly.",
			Severity:       models.Warning,
		})
	}

	// helm.releaseName
	if fy.Helm != nil {
		if fy.Helm.ReleaseName == "" {
			findings = append(findings, models.Finding{
				ID:             "W4",
				Title:          "Missing helm.releaseName",
				File:           file,
				Location:       "helm.releaseName",
				Issue:          "helm.releaseName is not set",
				Impact:         "Fleet will auto-generate a release name which can cause drift if the GitRepo is renamed.",
				Recommendation: "Set an explicit helm.releaseName.",
				Severity:       models.Warning,
			})
		}

		// helm.version pinning
		if fy.Helm.Chart != "" && !strings.HasPrefix(fy.Helm.Chart, "./") && fy.Helm.Version == "" {
			findings = append(findings, models.Finding{
				ID:             "W9",
				Title:          "Unpinned Helm chart version",
				File:           file,
				Location:       "helm.version",
				Issue:          "Helm chart version is not pinned",
				Impact:         "Deployments are not reproducible and may pick up unexpected upstream changes.",
				Recommendation: "Always pin Helm chart versions in production.",
				Severity:       models.Warning,
			})
		}

		// Positive findings
		if fy.Helm.Atomic {
			findings = append(findings, models.Finding{
				ID:             "I2",
				Title:          "Good practice: Atomic Helm upgrades",
				File:           file,
				Location:       "helm.atomic",
				Issue:          "helm.atomic is enabled",
				Impact:         "Rolls back on failed install/upgrade, ensuring system stability.",
				Recommendation: "Excellent choice for production environments.",
				Severity:       models.Info,
			})
		}
		if fy.Helm.WaitForJobs {
			findings = append(findings, models.Finding{
				ID:             "I3",
				Title:          "Good practice: Wait for jobs",
				File:           file,
				Location:       "helm.waitForJobs",
				Issue:          "helm.waitForJobs is enabled",
				Impact:         "Ensures database migrations or other jobs complete before proceeding.",
				Recommendation: "Recommended for complex deployments.",
				Severity:       models.Info,
			})
		}
	}

	// dependsOn
	if len(fy.DependsOn) > 0 {
		findings = append(findings, models.Finding{
			ID:             "I4",
			Title:          "Good practice: Explicit dependencies",
			File:           file,
			Location:       "dependsOn",
			Issue:          "Explicit dependencies are defined",
			Impact:         "Ensures correct ordering of deployments.",
			Recommendation: "Continues using dependsOn for prerequisites.",
			Severity:       models.Info,
		})
	}

	// Latest tag check in values
	if fy.Helm != nil {
		findings = append(findings, checkLatestTag(file, "helm.values", fy.Helm.Values)...)
	}

	// Target Customization ordering
	hasCatchAll := false
	for i, tc := range fy.TargetCustomizations {
		isCatchAll := false
		// Simplified catch-all detection: empty selector, group, and name
		if tc.ClusterSelector == nil && tc.ClusterGroup == "" && tc.ClusterName == "" {
			isCatchAll = true
		} else if m, ok := tc.ClusterSelector.(map[string]interface{}); ok && len(m) == 0 {
			isCatchAll = true
		}

		if isCatchAll {
			hasCatchAll = true
		} else if hasCatchAll {
			findings = append(findings, models.Finding{
				ID:             "C3",
				Title:          "Target customization ordering issue",
				File:           file,
				Location:       fmt.Sprintf("targetCustomizations[%d]", i),
				Issue:          fmt.Sprintf("Target '%s' comes after a catch-all target", tc.Name),
				Impact:         "The catch-all target will match first, and this customization will never be applied.",
				Recommendation: "Move more specific targets before broader or catch-all targets.",
				Severity:       models.Critical,
			})
		}
	}

	return findings
}

func checkLatestTag(file, location string, values map[string]interface{}) []models.Finding {
	findings := []models.Finding{}
	for k, v := range values {
		if k == "tag" && v == "latest" {
			findings = append(findings, models.Finding{
				ID:             "W5",
				Title:          "Using 'latest' image tag",
				File:           file,
				Location:       location,
				Issue:          "Image tag is set to 'latest'",
				Impact:         "Unpinned tags make deployments non-reproducible and can lead to unexpected version updates.",
				Recommendation: "Use specific version tags or commit SHAs.",
				Severity:       models.Warning,
			})
		}
		if nextMap, ok := v.(map[string]interface{}); ok {
			findings = append(findings, checkLatestTag(file, location+"."+k, nextMap)...)
		}
	}
	return findings
}

func detectHardcodedSecrets(file string, data []byte) []models.Finding {
	findings := []models.Finding{}
	lines := strings.Split(string(data), "\n")
	
	secretKeywords := []string{"password", "secret", "key", "token", "jwt"}
	
	for i, line := range lines {
		lowerLine := strings.ToLower(line)
		for _, kw := range secretKeywords {
			if strings.Contains(lowerLine, kw) {
				// Very simple heuristic: if it has a colon and a non-empty value that isn't a template or reference
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					val := strings.TrimSpace(parts[1])
					if val != "" && !strings.Contains(val, "{{") && !strings.Contains(val, "$") && !strings.HasPrefix(val, "secret-") {
						findings = append(findings, models.Finding{
							ID:             "C4",
							Title:          "Potential hardcoded secret",
							File:           file,
							Location:       fmt.Sprintf("line %d", i+1),
							Issue:          fmt.Sprintf("Field containing '%s' appears to have a hardcoded value", kw),
							Impact:         "Exposes sensitive credentials in Git.",
							Recommendation: "Use Kubernetes Secrets or a secret management solution.",
							Severity:       models.Critical,
						})
						break // only one finding per line
					}
				}
			}
		}
	}
	return findings
}
