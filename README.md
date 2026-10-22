<div align="center">

<img src="docs/images/tofuforge_logo.svg" alt="TofuForge Logo" width="160" height="160" />

# TofuForge

**Interactive Terraform and OpenTofu Plan Visualizer, Blast-Radius Inspector, and Resource Destruction Guard**

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![React Version](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react)](https://react.dev/)
[![OpenTofu](https://img.shields.io/badge/OpenTofu-Supported-FFDA44?style=flat-square&logo=opentofu)](https://opentofu.org/)
[![Terraform](https://img.shields.io/badge/Terraform-Supported-844FBA?style=flat-square&logo=terraform)](https://www.terraform.io/)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

<p align="center">
  <a href="#overview">Overview</a> •
  <a href="#key-capabilities">Key Capabilities</a> •
  <a href="#quickstart">Quickstart</a> •
  <a href="#cli-reference">CLI Reference</a> •
  <a href="#pre-flight-security-rules">Security Rules</a> •
  <a href="#ci-cd-integration">CI/CD Gating</a>
</p>

</div>

---

## Interactive Demo

<div align="center">
  <img src="docs/images/tofuforge_demo.gif" alt="TofuForge Interactive Web Studio Demo" width="900" />
</div>

---

## Overview

Terraform and OpenTofu execution plans often exceed several thousand lines of terminal text. During pull request reviews, engineers routinely overlook critical lines such as `~ forces replacement`, which can silently destroy and recreate production databases, route tables, or VPC subnets.

TofuForge is a single-binary static analysis engine and interactive visual studio designed to prevent catastrophic infrastructure deployments before `apply`:
1. Ingests raw JSON plan files (`terraform show -json` and `tofu show -json`).
2. Builds an explicit and implicit dependency directed acyclic graph (DAG).
3. Computes downstream blast radius for every resource scheduled for destruction or recreation.
4. Audits pre-flight security configurations against 18 multi-cloud rules (AWS, GCP, Azure, Kubernetes).
5. Generates compact GitHub Flavored Markdown summaries for pull request comments.
6. Serves an interactive React 19 web studio with real-time cascade tracing.

---

## Visual Studio Interface

<div align="center">
  <img src="docs/images/tofuforge-1.png" alt="TofuForge Dependency DAG and Destruction Alerts" width="900" />
</div>

<br/>

<div align="center">
  <img src="docs/images/tofuforge-2.png" alt="Side-by-Side Attribute Diff Inspector" width="900" />
</div>

---

## Key Capabilities

- **Universal Plan Support**: Works across all major cloud providers (AWS, Google Cloud, Microsoft Azure, Kubernetes) with standard JSON plan outputs from Terraform 0.12+ and all OpenTofu versions.
- **Blast-Radius DAG Engine**: Traces transitive downstream dependencies using breadth-first traversal. Computes a cumulative risk score that factors resource statefulness and dependency depth.
- **Destruction & Recreation Guard**: Flags database instances, object storage buckets, encryption keys, and virtual networks undergoing replacement with high-visibility alerts.
- **Pre-Flight Cloud Security Linter**: Detects `0.0.0.0/0` ingress rules on sensitive ports (22, 3389, 5432), unencrypted S3 buckets, public RDS instances, wildcard IAM statements, and privileged Kubernetes containers.
- **Automated PR Reporter**: Generates structured Markdown tables summarizing changes (+Create, ~Update, -Delete, ±Replace), critical destruction callouts, security findings, and collapsible attribute diffs.
- **Self-Contained Executable**: The React 19 frontend is compiled and embedded into the Go binary via `//go:embed`. No external node runtime, python, or docker daemon required.

---

## Quickstart

### 1. Installation

Build directly from source:
```bash
git clone https://github.com/alexandrmotologa/tofuforge.git
cd tofuforge
go build -o tofuforge ./cmd/tofuforge
```

### 2. Generate a Plan JSON

Using OpenTofu:
```bash
tofu plan -out=tfplan.binary
tofu show -json tfplan.binary > tfplan.json
```

Using Terraform:
```bash
terraform plan -out=tfplan.binary
terraform show -json tfplan.binary > tfplan.json
```

### 3. Inspect in Terminal

Run pre-flight checks and view a colorized analysis report:
```bash
./tofuforge inspect tfplan.json
```

Or pipe directly from stdin:
```bash
tofu show -json tfplan.binary | ./tofuforge inspect -
```

### 4. Launch Interactive Web Studio

Launch the embedded visual studio:
```bash
./tofuforge serve tfplan.json --open
```
Navigate to `http://localhost:8080` to explore the DAG, trace cascading impact, and inspect attribute before/after diffs.

---

## CLI Reference

### `tofuforge inspect [plan.json]`
Evaluates changes, scores blast radius, checks security policies, and executes CI gates.

```bash
# Basic inspection
tofuforge inspect tfplan.json

# Export analysis as JSON
tofuforge inspect tfplan.json --json

# Generate PR comment markdown file
tofuforge inspect tfplan.json -o pr-comment.md

# CI Gating: Fail if any resource is deleted or replaced
tofuforge inspect tfplan.json --fail-on-destructive

# CI Gating: Fail if stateful resources (databases, storage) undergo recreation
tofuforge inspect tfplan.json --fail-on-critical

# CI Gating: Fail on critical security flaws (open SSH/RDP/database ports)
tofuforge inspect tfplan.json --fail-on-security

# CI Gating: Fail if cumulative blast radius exceeds threshold
tofuforge inspect tfplan.json --max-blast-radius 25
```

### `tofuforge pr-comment [plan.json]`
Outputs a clean Markdown report formatted for GitHub and GitLab pull request comments.

```bash
tofuforge pr-comment tfplan.json
```

### `tofuforge serve [plan.json]`
Starts the embedded web studio HTTP server.

```bash
tofuforge serve tfplan.json --port 8080 --host localhost
```

---

## Pre-Flight Security Rules

| Provider | Rule ID | Title | Severity |
| :--- | :--- | :--- | :--- |
| **AWS** | `AWS-SEC-001` | Security Group allows public ingress on sensitive ports (22, 3389, 5432, 3306) | `CRITICAL` |
| **AWS** | `AWS-SEC-002` | Security Group ingress permits unrestricted internet traffic (all protocols/ports) | `CRITICAL` |
| **AWS** | `AWS-SEC-003` | S3 bucket missing server-side encryption configuration | `HIGH` |
| **AWS** | `AWS-SEC-004` | RDS database instance is publicly accessible | `CRITICAL` |
| **AWS** | `AWS-SEC-005` | RDS database storage encryption is disabled | `HIGH` |
| **AWS** | `AWS-SEC-006` | IAM policy contains wildcard actions (`*`) | `HIGH` |
| **AWS** | `AWS-SEC-007` | EBS volume encryption is disabled | `MEDIUM` |
| **GCP** | `GCP-SEC-001` | Compute firewall allows public ingress to administrative or database ports | `CRITICAL` |
| **GCP** | `GCP-SEC-002` | Cloud SQL instance has public IPv4 with 0.0.0.0/0 authorized network | `CRITICAL` |
| **GCP** | `GCP-SEC-003` | IAM policy grants roles to public internet (`allUsers` or `allAuthenticatedUsers`) | `CRITICAL` |
| **GCP** | `GCP-SEC-004` | Cloud Storage uniform bucket-level access is disabled | `MEDIUM` |
| **Azure** | `AZR-SEC-001` | NSG inbound rule allows internet traffic to sensitive ports | `CRITICAL` |
| **Azure** | `AZR-SEC-002` | Storage account allows unencrypted HTTP traffic | `HIGH` |
| **Azure** | `AZR-SEC-003` | Storage account public blob access is enabled | `HIGH` |
| **Azure** | `AZR-SEC-004` | MSSQL database transparent data encryption is disabled | `HIGH` |
| **K8s** | `K8S-SEC-001` | Container runs in privileged mode | `CRITICAL` |
| **K8s** | `K8S-SEC-002` | Container does not enforce non-root execution | `HIGH` |
| **K8s** | `K8S-SEC-003` | Workload container missing CPU or memory resource limits | `MEDIUM` |

---

## CI/CD Gating Example

```yaml
name: TofuForge Safety Check

on:
  pull_request:
    branches: [main]

jobs:
  audit-plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: opentofu/setup-opentofu@v1
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install TofuForge
        run: go install github.com/alexandrmotologa/tofuforge/cmd/tofuforge@latest

      - name: Generate Plan
        run: |
          tofu init
          tofu plan -out=tfplan.binary
          tofu show -json tfplan.binary > tfplan.json

      - name: Inspect Plan & Block Dangerous Destructions
        run: |
          tofuforge inspect tfplan.json \
            --fail-on-critical \
            --fail-on-security \
            -o pr-summary.md

      - name: Post Comment to PR
        if: always()
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            if (fs.existsSync('pr-summary.md')) {
              const body = fs.readFileSync('pr-summary.md', 'utf8');
              github.rest.issues.createComment({
                issue_number: context.issue.number,
                owner: context.repo.owner,
                repo: context.repo.repo,
                body: body
              });
            }
```

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
