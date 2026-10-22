# TofuForge CLI & CI/CD Integration Guide

TofuForge provides command-line tooling to inspect Terraform and OpenTofu plans, evaluate blast radius, flag critical destruction, run security linting, estimate cost deltas, compare revisions, and enforce automated pull request gates.

---

## 1. Installation

### From Source (Go 1.23+)
```bash
git clone https://github.com/alexandrmotologa/tofuforge.git
cd tofuforge
go build -o tofuforge ./cmd/tofuforge
```

---

## 2. Command Reference

### `tofuforge inspect [plan.json]`
Parses a plan file, outputs a terminal report, audits security policies, estimates cost changes, and evaluates CI thresholds.

```bash
# Basic inspection with colorized terminal summary
tofuforge inspect tfplan.json

# Read from standard input pipe
tofu show -json tfplan.binary | tofuforge inspect -

# Output analyzed data as JSON
tofuforge inspect tfplan.json --json

# Write markdown PR summary to a file
tofuforge inspect tfplan.json -o pr-summary.md

# Load custom governance policy
tofuforge inspect tfplan.json -c .tofuforge.yaml
```

#### CI Failure Flags
| Flag | Description | Exit Code |
| :--- | :--- | :--- |
| `--fail-on-destructive` | Exits with code 1 if any resource is deleted or replaced | 1 on match |
| `--fail-on-critical` | Exits with code 1 if any critical destruction alerts exist (e.g. database recreation) | 1 on match |
| `--fail-on-security` | Exits with code 1 if any critical security violations exist | 1 on match |
| `--max-blast-radius <int>` | Exits with code 1 if cumulative blast radius exceeds threshold | 1 on match |
| `--max-cost-delta <float>` | Exits with code 1 if monthly cost delta exceeds threshold in USD | 1 on match |
| `--no-color` | Disables ANSI terminal color sequences | 0 |

### `tofuforge diff [base.json] [target.json]`
Compares two plan revisions to verify resolved destruction risks, action changes, and blast radius delta.

```bash
tofuforge diff base_plan.json target_plan.json
```

Example output:
```
================================================================================
                    TOFUFORGE PLAN REVISION COMPARISON                          
================================================================================
Blast Radius Score Delta:  -38 (Prior: 38 -> Current: 0)
Critical Alerts Delta:      -1
Security Flaws Delta:       -2

[RESOLVED DESTRUCTIONS]:
  ✓ aws_db_instance.customer_orders_postgresql - no longer marked for replace

[RESOLVED SECURITY FINDINGS]:
  ✓ AWS-SEC-001: Security Group Ingress Allows Public SSH (Port 22)
  ✓ AWS-SEC-003: S3 Bucket Missing Server-Side Encryption Configuration

[ACTION TRANSITIONS]:
  • aws_db_instance.customer_orders_postgresql: replace -> update
================================================================================
```

### `tofuforge report [plan.json]`
Generates a standalone, self-contained single-file HTML report embedding the complete React 19 visual studio for offline viewing or CI artifact archiving.

```bash
tofuforge report tfplan.json -o tofuforge-report.html
```

### `tofuforge pr-comment [plan.json]`
Generates a GitHub/GitLab pull request comment formatted in GitHub Flavored Markdown.

```bash
# Print markdown comment to stdout
tofuforge pr-comment tfplan.json

# Save comment to file for CI posting
tofuforge pr-comment tfplan.json -o pr-comment.md
```

### `tofuforge serve [plan.json]`
Launches the local interactive web studio with React Flow DAG visualization, destruction alerts, layout controls, and attribute diffs.

```bash
# Start server on default port 8080
tofuforge serve tfplan.json

# Specify port and open browser automatically
tofuforge serve tfplan.json --port 9000 --open

# Start empty studio (upload plan via browser drag-and-drop)
tofuforge serve
```

---

## 3. Governance Policy Configuration (`.tofuforge.yaml`)

Place `.tofuforge.yaml` in your project root or pass via `-c / --config`:

```yaml
version: 1

rules:
  # Rule IDs to ignore
  ignore:
    - "AWS-SEC-007"

  # High-priority rules that trigger strict failure
  strict_rules:
    - "AWS-SEC-001"
    - "GCP-SEC-001"

# Mandatory tags required for all taggable resources
mandatory_tags:
  - "Environment"
  - "Owner"
  - "CostCenter"
```

---

## 4. GitHub Actions Integration

```yaml
name: Terraform Safety Gate

on:
  pull_request:
    branches: [main]

jobs:
  plan-and-inspect:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Setup OpenTofu
        uses: opentofu/setup-opentofu@v1

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install TofuForge
        run: go install github.com/alexandrmotologa/tofuforge/cmd/tofuforge@latest

      - name: Tofu Init & Plan
        run: |
          tofu init
          tofu plan -out=tfplan.binary
          tofu show -json tfplan.binary > tfplan.json

      - name: Run TofuForge Inspection
        run: |
          tofuforge inspect tfplan.json \
            --fail-on-critical \
            --fail-on-security \
            --max-cost-delta 150 \
            -o pr-comment.md

      - name: Generate Offline HTML Artifact
        if: always()
        run: tofuforge report tfplan.json -o tofuforge-report.html

      - name: Upload HTML Report Artifact
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: tofuforge-report
          path: tofuforge-report.html

      - name: Post PR Summary
        if: always()
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            if (fs.existsSync('pr-comment.md')) {
              const body = fs.readFileSync('pr-comment.md', 'utf8');
              github.rest.issues.createComment({
                issue_number: context.issue.number,
                owner: context.repo.owner,
                repo: context.repo.repo,
                body: body
              });
            }
```

---

## 5. GitLab CI Integration

```yaml
stages:
  - plan
  - inspect

tofu_plan:
  stage: plan
  script:
    - tofu init
    - tofu plan -out=tfplan.binary
    - tofu show -json tfplan.binary > tfplan.json
  artifacts:
    paths:
      - tfplan.json

tofuforge_inspect:
  stage: inspect
  dependencies:
    - tofu_plan
  script:
    - tofuforge inspect tfplan.json --fail-on-critical --fail-on-security --max-cost-delta 150 -o pr-summary.md
    - tofuforge report tfplan.json -o tofuforge-report.html
  artifacts:
    paths:
      - tofuforge-report.html
    reports:
      dotenv: tofuforge.env
```
