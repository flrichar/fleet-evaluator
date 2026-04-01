# Fleet Best Practices Reference

## Table of Contents
1. [GitRepo Configuration](#gitrepo-configuration)
2. [fleet.yaml Configuration](#fleetyaml-configuration)
3. [Targeting and Customization](#targeting-and-customization)
4. [Helm Integration](#helm-integration)
5. [Security](#security)
6. [Multi-Cluster Patterns](#multi-cluster-patterns)
7. [Drift Detection and Correction](#drift-detection-and-correction)
8. [Common Anti-Patterns](#common-anti-patterns)

---

## GitRepo Configuration

### Resource Structure (API group: fleet.cattle.io/v1alpha1)

```yaml
apiVersion: fleet.cattle.io/v1alpha1
kind: GitRepo
metadata:
  name: my-app
  namespace: fleet-default  # or fleet-local for local cluster
spec:
  repo: https://github.com/org/repo
  branch: main
  paths:
    - charts/my-app
  targets:
    - clusterSelector:
        matchLabels:
          env: production
  pollingInterval: 60s
  clientSecretName: git-credentials
  serviceAccount: fleet-deployer
```

### Key fields and recommendations

| Field | Best Practice |
|-------|--------------|
| `spec.repo` | Use HTTPS+credentials or SSH for private repos. Public HTTPS is OK for open-source in non-prod. |
| `spec.branch` | Fine for dev/staging. For prod, prefer `spec.revision` with a tag or SHA. |
| `spec.paths` | Always specify — deploying repo root risks deploying unintended manifests. |
| `spec.pollingInterval` | Default 15s is aggressive for large repos or rate-limited Git providers. 60s-300s is more reasonable for production. |
| `spec.clientSecretName` | Required for private repos. Should reference a Secret of type `kubernetes.io/basic-auth` or `kubernetes.io/ssh-auth`. |
| `spec.helmSecretName` | For Helm repos requiring auth. Same Secret types apply. |
| `spec.serviceAccount` | Dedicated SA with minimal RBAC is recommended over default. |
| `spec.forceSyncGeneration` | Incrementing this forces a redeploy. Should not be set to arbitrary high values permanently. |
| `spec.correctDrift.enabled` | Reverts manual changes to match Git state. Enable in prod, consider disabling in dev. |
| `spec.correctDrift.force` | Force-deletes and recreates resources on drift. Use with extreme caution. |
| `spec.imageScanInterval` | If using image scanning, tune to avoid excessive registry load. |
| `spec.insecureSkipTLSVerify` | Should never be true in production. |
| `spec.caBundle` | Required if using self-signed certs for Git server. |
| `spec.keepResources` | When true, deployed resources are not deleted when the GitRepo is removed. Consider for stateful workloads. |

### Namespace conventions
- `fleet-default`: For managing downstream clusters
- `fleet-local`: For managing the local/management cluster
- Using the wrong namespace is a common source of "nothing deploys" confusion

---

## fleet.yaml Configuration

The `fleet.yaml` file sits in the Git repo alongside the manifests it configures.

```yaml
defaultNamespace: my-app
helm:
  releaseName: my-app
  chart: ./chart
  values:
    replicaCount: 3
  valuesFiles:
    - values.yaml
    - values-prod.yaml
targetCustomizations:
  - name: staging
    clusterSelector:
      matchLabels:
        env: staging
    helm:
      values:
        replicaCount: 1
      valuesFiles:
        - values-staging.yaml
dependsOn:
  - name: cert-manager
```

### Key fields

| Field | Best Practice |
|-------|--------------|
| `defaultNamespace` | Always set. Without it, resources deploy to the GitRepo's namespace in the management cluster, which is almost never what you want. |
| `namespace` | Overrides the namespace for all resources. Use when you need strict namespace control. |
| `helm.releaseName` | Always set explicitly. Auto-generated names cause issues on GitRepo renames. |
| `helm.version` | Pin to a specific version in production (e.g., `1.2.3`). Never use `*`. |
| `helm.chart` | Can be a local path (`./chart`), a chart name from `helm.repo`, or an OCI reference. |
| `helm.repo` | URL of the Helm repository. Required when `helm.chart` is not a local path. |
| `helm.values` | Inline values — OK for small overrides. For anything substantial, use `valuesFiles`. |
| `helm.valuesFiles` | List of value files, applied in order (later files override earlier). Good for layered config. |
| `helm.atomic` | Rolls back on failed install/upgrade. Recommended for production. |
| `helm.waitForJobs` | Waits for Jobs to complete during install. Enable for migration-dependent deployments. |
| `kustomize.dir` | Path to kustomization directory. |
| `yaml.overlays` | List of overlay files for plain YAML (non-Helm, non-Kustomize). |
| `dependsOn` | Bundles that must be ready before this one deploys. Verify referenced bundles exist. |
| `targetCustomizations` | Per-cluster/group overrides. See targeting section. |

---

## Targeting and Customization

### Target selection hierarchy
Fleet evaluates targets in order. The first matching target wins. This means:
- More specific targets should come before broader ones
- A catch-all target at the end is fine, but be aware it matches everything

### Selector types
```yaml
targets:
  # Match by labels
  - clusterSelector:
      matchLabels:
        env: production
      matchExpressions:
        - key: region
          operator: In
          values: [us-east-1, us-west-2]

  # Match by cluster group name
  - clusterGroup: production-clusters

  # Match by cluster name
  - clusterName: specific-cluster-01
```

### Common targeting mistakes
- **Empty targets array** — deploys nowhere (the GitRepo is a no-op)
- **Missing targets entirely** — deploys to all clusters in the namespace (often unintentional)
- **Overlapping selectors in targetCustomizations** — if two customizations match the same cluster, the first one wins. This is subtle and can cause "why aren't my values applying?" bugs
- **Mismatched label keys** — `env` vs `environment`, `region` vs `topology.kubernetes.io/region`. Inconsistent labeling across clusters is the #1 source of targeting bugs.

---

## Helm Integration

### Chart sources
- **Local chart** (`helm.chart: ./chart`): Versioned with Git. Good for apps you own.
- **Helm repo** (`helm.repo` + `helm.chart`): For third-party charts. Always pin `helm.version`.
- **OCI** (`helm.chart: oci://registry/chart`): Modern approach. Same version pinning advice applies.

### Value layering
Values are applied in this order (later overrides earlier):
1. Chart defaults
2. `helm.valuesFiles` (in list order)
3. `helm.values` (inline)
4. `targetCustomizations[].helm.valuesFiles`
5. `targetCustomizations[].helm.values`

This layering is powerful but can be confusing. Document the intended layering in comments.

---

## Security

### Secrets management
- Git credentials (`clientSecretName`), Helm credentials (`helmSecretName`): Must reference existing Secrets in the same namespace as the GitRepo
- Never hardcode credentials in fleet.yaml or GitRepo manifests
- Use `kubernetes.io/basic-auth` or `kubernetes.io/ssh-auth` Secret types
- Look for base64-encoded values in YAML — these are NOT encrypted, just encoded

### RBAC
- Fleet controller needs broad permissions by default. In hardened environments, scope down the `fleet-controller` ClusterRole
- `spec.serviceAccount` on GitRepo: Use a dedicated SA with only the permissions needed for that deployment
- In multi-tenant setups, use separate Fleet namespaces per tenant

### Network / TLS
- `insecureSkipTLSVerify: true` is a security risk — should never appear in production
- `caBundle` should be used for internal CAs instead of skipping TLS
- If using private Git servers, ensure network policies allow Fleet controller to reach them

---

## Multi-Cluster Patterns

### Recommended label taxonomy
```yaml
# On Cluster resources
metadata:
  labels:
    env: production          # or staging, development
    region: us-east-1        # cloud region or datacenter
    provider: aws            # aws, gcp, azure, on-prem
    tier: frontend           # workload tier
    team: platform           # owning team
```

Consistent labels across all clusters are essential. One cluster missing `env` label means it won't match `env`-based selectors.

### ClusterGroup best practices
- Name groups by purpose, not just environment: `production-us-east`, `staging-gpu`
- Use `spec.selector.matchLabels` for stable grouping
- Avoid `matchExpressions` with `NotIn` or `DoesNotExist` — these create implicit groups that grow unexpectedly as new clusters join

### Environment promotion pattern
```
repo/
├── base/           # shared manifests
├── overlays/
│   ├── dev/        # dev-specific fleet.yaml + values
│   ├── staging/    # staging-specific
│   └── prod/       # prod-specific
```
Each environment's GitRepo points to a different `spec.paths` entry. This keeps environment configs isolated while sharing a base.

---

## Drift Detection and Correction

### How drift correction works
When `correctDrift.enabled: true`, Fleet periodically compares the actual cluster state with the desired state from Git. If they differ, Fleet reverts the change.

### Settings
| Field | Effect |
|-------|--------|
| `correctDrift.enabled` | Enables drift detection and correction |
| `correctDrift.force` | Force-deletes and recreates resources. Dangerous for stateful workloads. |
| `correctDrift.keepFailHistory` | Keeps failed correction attempts for debugging |

### Recommendations
- **Production**: Enable `correctDrift.enabled`. This is the whole point of GitOps.
- **Development**: Consider disabling to allow manual experimentation.
- **Never** use `correctDrift.force` on StatefulSets, PVCs, or CRDs without understanding the consequences.

---

## Common Anti-Patterns

1. **Deploying from repo root** — `paths: ["/"]` or omitting paths entirely. Risks deploying test files, docs, CI configs as Kubernetes manifests.

2. **Unpinned chart versions** — Using `version: "*"` or omitting `helm.version`. Production deployments should be reproducible.

3. **Missing defaultNamespace** — Resources end up in `fleet-default` or `fleet-local` namespace on the downstream cluster, which is confusing and wrong.

4. **Overly broad targets** — Deploying to all clusters when only some should receive the workload. Always be explicit.

5. **Mixing concerns in one GitRepo** — One massive GitRepo that deploys everything. Better to split by team, application, or concern for independent lifecycle management.

6. **Ignoring pollingInterval** — Default 15s polling on many GitRepos can hit Git provider rate limits. Especially on GitHub with many repos.

7. **No dependsOn for prerequisites** — Deploying an app before its CRDs or namespace are ready. Use `dependsOn` to express ordering.

8. **Hardcoded cluster names in targets** — `clusterName: prod-cluster-1` instead of label selectors. Breaks when clusters are replaced or scaled.

9. **Inconsistent label taxonomy** — Different teams using `env` vs `environment` vs `stage`. Standardize across the organization.

10. **Using insecureSkipTLSVerify in production** — Always provide `caBundle` for internal CAs instead.
