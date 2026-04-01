# Fleet Configuration Evaluation Report

**Date:** 2026-03-25
**Files evaluated:**
- `fleet-eval/evals/files/scenario-1/fleet.yaml`
- `fleet-eval/evals/files/scenario-1/gitrepo-prod.yaml`
- `fleet-eval/evals/files/scenario-2/clustergroups.yaml`
- `fleet-eval/evals/files/scenario-2/fleet-auth.yaml`
- `fleet-eval/evals/files/scenario-2/gitrepo-dev.yaml`
- `fleet-eval/evals/files/scenario-2/gitrepo-prod.yaml`
- `fleet-eval/evals/files/scenario-3/fleet-grafana.yaml`
- `fleet-eval/evals/files/scenario-3/fleet-prometheus.yaml`
- `fleet-eval/evals/files/scenario-3/gitrepo-monitoring.yaml`

## Summary

| Severity | Count |
|----------|-------|
| CRITICAL | 9 |
| WARNING  | 10 |
| INFO     | 6 |

Overall assessment: Several critical issues were found that need immediate attention.

## Findings

### CRITICAL

#### [C3] Target customization ordering issue
- **File:** `fleet-eval/evals/files/scenario-1/fleet.yaml`
- **Location:** `targetCustomizations[2]`
- **Issue:** Target 'production' comes after a catch-all target
- **Impact:** The catch-all target will match first, and this customization will never be applied.
- **Recommendation:** Move more specific targets before broader or catch-all targets.

#### [C1] Deploying from repo root
- **File:** `fleet-eval/evals/files/scenario-1/gitrepo-prod.yaml`
- **Location:** `spec.paths`
- **Issue:** GitRepo is configured to deploy from the root directory ('/')
- **Impact:** Risks deploying unintended manifests like CI configs, docs, or test files.
- **Recommendation:** Specify explicit paths for Kubernetes manifests.

#### [C2] Insecure TLS verification
- **File:** `fleet-eval/evals/files/scenario-1/gitrepo-prod.yaml`
- **Location:** `spec.insecureSkipTLSVerify`
- **Issue:** insecureSkipTLSVerify is set to true
- **Impact:** Disables TLS certificate validation, making the connection vulnerable to MITM attacks.
- **Recommendation:** Use a valid CA bundle instead of skipping verification.

#### [C4] Potential hardcoded secret
- **File:** `fleet-eval/evals/files/scenario-2/fleet-auth.yaml`
- **Location:** `line 10`
- **Issue:** Field containing 'key' appears to have a hardcoded value
- **Impact:** Exposes sensitive credentials in Git.
- **Recommendation:** Use Kubernetes Secrets or a secret management solution.

#### [C4] Potential hardcoded secret
- **File:** `fleet-eval/evals/files/scenario-2/fleet-auth.yaml`
- **Location:** `line 11`
- **Issue:** Field containing 'password' appears to have a hardcoded value
- **Impact:** Exposes sensitive credentials in Git.
- **Recommendation:** Use Kubernetes Secrets or a secret management solution.

#### [C4] Potential hardcoded secret
- **File:** `fleet-eval/evals/files/scenario-2/gitrepo-dev.yaml`
- **Location:** `line 9`
- **Issue:** Field containing 'secret' appears to have a hardcoded value
- **Impact:** Exposes sensitive credentials in Git.
- **Recommendation:** Use Kubernetes Secrets or a secret management solution.

#### [C4] Potential hardcoded secret
- **File:** `fleet-eval/evals/files/scenario-2/gitrepo-prod.yaml`
- **Location:** `line 9`
- **Issue:** Field containing 'secret' appears to have a hardcoded value
- **Impact:** Exposes sensitive credentials in Git.
- **Recommendation:** Use Kubernetes Secrets or a secret management solution.

#### [C4] Potential hardcoded secret
- **File:** `fleet-eval/evals/files/scenario-3/fleet-grafana.yaml`
- **Location:** `line 7`
- **Issue:** Field containing 'password' appears to have a hardcoded value
- **Impact:** Exposes sensitive credentials in Git.
- **Recommendation:** Use Kubernetes Secrets or a secret management solution.

#### [C4] Potential hardcoded secret
- **File:** `fleet-eval/evals/files/scenario-3/gitrepo-monitoring.yaml`
- **Location:** `line 16`
- **Issue:** Field containing 'secret' appears to have a hardcoded value
- **Impact:** Exposes sensitive credentials in Git.
- **Recommendation:** Use Kubernetes Secrets or a secret management solution.

### WARNING

#### [W3] Missing defaultNamespace
- **File:** `fleet-eval/evals/files/scenario-1/fleet.yaml`
- **Location:** `defaultNamespace`
- **Issue:** defaultNamespace is not set in fleet.yaml
- **Impact:** Resources might deploy to the GitRepo's namespace in the management cluster by mistake.
- **Recommendation:** Always set defaultNamespace explicitly.

#### [W4] Missing helm.releaseName
- **File:** `fleet-eval/evals/files/scenario-1/fleet.yaml`
- **Location:** `helm.releaseName`
- **Issue:** helm.releaseName is not set
- **Impact:** Fleet will auto-generate a release name which can cause drift if the GitRepo is renamed.
- **Recommendation:** Set an explicit helm.releaseName.

#### [W5] Using 'latest' image tag
- **File:** `fleet-eval/evals/files/scenario-1/fleet.yaml`
- **Location:** `helm.values.image`
- **Issue:** Image tag is set to 'latest'
- **Impact:** Unpinned tags make deployments non-reproducible and can lead to unexpected version updates.
- **Recommendation:** Use specific version tags or commit SHAs.

#### [W1] Aggressive polling interval
- **File:** `fleet-eval/evals/files/scenario-1/gitrepo-prod.yaml`
- **Location:** `spec.pollingInterval`
- **Issue:** Polling interval is set to 5s
- **Impact:** Very low intervals increase API load on Git providers and may lead to rate limiting.
- **Recommendation:** Increase polling interval to 60s or more.

#### [W2] Empty targets array
- **File:** `fleet-eval/evals/files/scenario-1/gitrepo-prod.yaml`
- **Location:** `spec.targets`
- **Issue:** Targets array is empty
- **Impact:** GitRepo will not deploy to any clusters.
- **Recommendation:** Specify cluster selectors or group names in targets.

#### [W6] Missing serviceAccount for production
- **File:** `fleet-eval/evals/files/scenario-1/gitrepo-prod.yaml`
- **Location:** `spec.serviceAccount`
- **Issue:** Production GitRepo does not have a dedicated serviceAccount
- **Impact:** Fleet will use its default service account, which may have excessive permissions.
- **Recommendation:** Use a dedicated serviceAccount with minimal RBAC for production deployments.

#### [W1] Aggressive polling interval
- **File:** `fleet-eval/evals/files/scenario-2/gitrepo-dev.yaml`
- **Location:** `spec.pollingInterval`
- **Issue:** Polling interval is set to 30s
- **Impact:** Very low intervals increase API load on Git providers and may lead to rate limiting.
- **Recommendation:** Increase polling interval to 60s or more.

#### [W6] Missing serviceAccount for production
- **File:** `fleet-eval/evals/files/scenario-2/gitrepo-prod.yaml`
- **Location:** `spec.serviceAccount`
- **Issue:** Production GitRepo does not have a dedicated serviceAccount
- **Impact:** Fleet will use its default service account, which may have excessive permissions.
- **Recommendation:** Use a dedicated serviceAccount with minimal RBAC for production deployments.

#### [W9] Unpinned Helm chart version
- **File:** `fleet-eval/evals/files/scenario-3/fleet-grafana.yaml`
- **Location:** `helm.version`
- **Issue:** Helm chart version is not pinned
- **Impact:** Deployments are not reproducible and may pick up unexpected upstream changes.
- **Recommendation:** Always pin Helm chart versions in production.

#### [W7] Inconsistent label keys
- **File:** `fleet-eval/evals/files/scenario-2/gitrepo-prod.yaml`
- **Location:** `spec.targets.clusterSelector.matchLabels`
- **Issue:** Both 'env' and 'environment' label keys are used across GitRepos
- **Impact:** Inconsistent labeling makes targeting confusing and prone to errors.
- **Recommendation:** Standardize on a single label taxonomy (e.g., always use 'env').

### INFO

#### [I4] Good practice: Explicit dependencies
- **File:** `fleet-eval/evals/files/scenario-2/fleet-auth.yaml`
- **Location:** `dependsOn`
- **Issue:** Explicit dependencies are defined
- **Impact:** Ensures correct ordering of deployments.
- **Recommendation:** Continues using dependsOn for prerequisites.

#### [I4] Good practice: Explicit dependencies
- **File:** `fleet-eval/evals/files/scenario-3/fleet-grafana.yaml`
- **Location:** `dependsOn`
- **Issue:** Explicit dependencies are defined
- **Impact:** Ensures correct ordering of deployments.
- **Recommendation:** Continues using dependsOn for prerequisites.

#### [I2] Good practice: Atomic Helm upgrades
- **File:** `fleet-eval/evals/files/scenario-3/fleet-prometheus.yaml`
- **Location:** `helm.atomic`
- **Issue:** helm.atomic is enabled
- **Impact:** Rolls back on failed install/upgrade, ensuring system stability.
- **Recommendation:** Excellent choice for production environments.

#### [I3] Good practice: Wait for jobs
- **File:** `fleet-eval/evals/files/scenario-3/fleet-prometheus.yaml`
- **Location:** `helm.waitForJobs`
- **Issue:** helm.waitForJobs is enabled
- **Impact:** Ensures database migrations or other jobs complete before proceeding.
- **Recommendation:** Recommended for complex deployments.

#### [I4] Good practice: Explicit dependencies
- **File:** `fleet-eval/evals/files/scenario-3/fleet-prometheus.yaml`
- **Location:** `dependsOn`
- **Issue:** Explicit dependencies are defined
- **Impact:** Ensures correct ordering of deployments.
- **Recommendation:** Continues using dependsOn for prerequisites.

#### [I1] Good practice: Revision pinning
- **File:** `fleet-eval/evals/files/scenario-3/gitrepo-monitoring.yaml`
- **Location:** `spec.revision`
- **Issue:** GitRepo uses a specific revision (tag or SHA)
- **Impact:** Ensures reproducible deployments.
- **Recommendation:** Continue using revision pinning for production.

## Environment Comparison

| Feature | Development | Production |
|---------|-------------|------------|
| Files | `fleet-eval/evals/files/scenario-2/gitrepo-dev.yaml` | `fleet-eval/evals/files/scenario-2/gitrepo-prod.yaml` |
| Consistent Labels | No | No |
| Drift Correction | Disabled | Enabled |

