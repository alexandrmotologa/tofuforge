# TofuForge CLI & CI/CD Integration Guide

TofuForge provides command-line tooling to inspect Terraform and OpenTofu plans, evaluate blast radius, flag critical destruction, run security linting, and enforce automated pull request gates.

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
Parses a plan file, outputs a terminal report, audits security policies, and evaluates CI thresholds.

```bash
# Basic inspection with colorized terminal summary
tofuforge inspect tfplan.json

# Read from standard input pipe
tofu show -json tfplan.binary | tofuforge inspect -

# Output analyzed data as JSON
tofuforge inspect tfplan.json --json

# Write markdown PR summary to a file
tofuforge inspect tfplan.json -o pr-summary.md
```

#### CI Failure Flags
| Flag | Description | Exit Code |
| :--- | :--- | :--- |
| `--fail-on-destructive` | Exits with code 1 if any resource is deleted or replaced | 1 on match |
| `--fail-on-critical` | Exits with code 1 if any critical destruction alerts exist (e.g. database recreation) | 1 on match |
| `--fail-on-security` | Exits with code 1 if any critical security violations exist | 1 on match |
| `--max-blast-radius <int>` | Exits with code 1 if cumulative blast radius exceeds threshold | 1 on match |
| `--no-color` | Disables ANSI terminal color sequences | 0 |

### `tofuforge pr-comment [plan.json]`
Generates a GitHub/GitLab pull request comment formatted in GitHub Flavored Markdown.

```bash
# Print markdown comment to stdout
tofuforge pr-comment tfplan.json

# Save comment to file for CI posting
tofuforge pr-comment tfplan.json -o pr-comment.md
```

### `tofuforge serve [plan.json]`
Launches the local interactive web studio with React Flow DAG visualization, destruction alerts, and attribute diffs.

```bash
# Start server on default port 8080
tofuforge serve tfplan.json

# Specify port and open browser automatically
tofuforge serve tfplan.json --port 9000 --open

# Start empty studio (upload plan via browser drag-and-drop)
tofuforge serve
```

---

## 3. GitHub Actions Integration

Use this workflow to generate plan comments and block pull requests containing critical database destructions or open security group rules:

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
            -o pr-comment.md

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

## 4. GitLab CI Integration

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
    - tofuforge inspect tfplan.json --fail-on-critical --fail-on-security -o pr-summary.md
  artifacts:
    reports:
      dotenv: tofuforge.env
```
