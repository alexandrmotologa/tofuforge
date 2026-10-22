# TofuForge: Master Implementation Plan

> Interactive Terraform and OpenTofu plan visualizer, dependency DAG builder, blast-radius inspector, and CI/CD security linter.
> Ingests plan JSON files, computes directed dependency graphs, highlights destructive replacements, audits cloud security posture, and generates PR comment summaries.

---

## 1. Project Overview & Architecture

Modern cloud engineering teams deploy infrastructure through Terraform and OpenTofu. Running `terraform plan` frequently produces thousands of lines of terminal output, causing critical issues to be overlooked:
1. Accidental resource destructions: Engineers miss subtle `~ forces replacement` lines that delete and recreate databases, storage buckets, or VPC networking.
2. Hidden blast radius: Modifying foundational components (subnets, security groups, route tables) cascades into unintended updates across dependent microservices.
3. Security pre-flight oversights: Pull requests introducing insecure ingress rules (0.0.0.0/0), unencrypted storage, or wildcard IAM policies pass through without early detection.
4. Poor PR review clarity: Reviewing raw terminal logs in pull requests is slow and prone to human error.

TofuForge provides an automated visual inspection and pre-flight safety control plane:
- Universal plan ingestion: Parses `terraform show -json` and `tofu show -json` across AWS, GCP, Azure, and Kubernetes.
- Dependency DAG engine: Builds a directed acyclic graph mapping resources and their explicit and implicit dependencies.
- Blast radius and destruction guard: Calculates downstream impact for every deleted or replaced resource, flagging stateful data services with critical warnings.
- Multi-cloud security linter: Evaluates over 18 pre-flight security rules across AWS, GCP, Azure, and Kubernetes before `apply`.
- Pull request reporter and CI gating: Generates compact Markdown comment summaries and enforces threshold exit codes (`--fail-on-destructive`, `--fail-on-security-critical`).
- Interactive web studio: React 19 and React Flow visualization interface with node filtering, cascade tracing, attribute diffs, and drag-and-drop plan inspection.

```
+-----------------------------------------------------------------------------------+
|                                     TOFUFORGE                                     |
+-----------------------------------------+-----------------------------------------+
                                          |
        +---------------------------------+---------------------------------+
        |                                                                   |
+-------v-------------------------+                 +-----------------------v-------+
|    Plan JSON Ingestion Engine   |                 |    Blast-Radius DAG Engine    |
|  * Terraform / OpenTofu Plan    |                 |  * Dependency Graph Generator |
|  * Resource Changes Parser      |                 |  * Downstream Impact Tracer   |
|  * Action & Statefulness Parser |                 |  * Replace Cascade Calculator |
+-------+-------------------------+                 +-----------------------+-------+
        |                                                                   |
        +---------------------------------+---------------------------------+
                                          |
                        +-----------------v-----------------+
                        |    Security & Compliance Linter   |
                        |  * AWS: Ingress, S3, RDS, IAM     |
                        |  * GCP: Firewall, Storage, SQL    |
                        |  * Azure: NSG, Storage, SQL TDE   |
                        |  * K8s: Privileged, Root, Limits  |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |   Interactive Web Studio & CLI    |
                        |  * React Flow Graph Visualizer    |
                        |  * Destruction Alert Panel        |
                        |  * Side-by-side Attribute Diff    |
                        |  * GitHub PR Markdown Generator   |
                        +-----------------------------------+
```

---

## 2. Directory Structure

```
tofuforge/
├── cmd/
│   └── tofuforge/
│       └── main.go               # CLI entrypoint (inspect, serve, pr-comment)
├── internal/
│   ├── model/
│   │   ├── plan.go               # Unified data structures for plans, changes, diffs
│   │   └── analysis.go           # Graph, blast radius, and security analysis models
│   ├── parser/
│   │   ├── plan_parser.go        # Terraform and OpenTofu JSON schema ingestion
│   │   ├── actions.go            # Action classifier (create, update, delete, replace)
│   │   └── statefulness.go       # Stateful and critical resource identification
│   ├── graph/
│   │   ├── dag.go                # Directed acyclic graph builder
│   │   ├── blast_radius.go       # Downstream impact and cascade calculator
│   │   └── topological.go        # Topological sorting and execution ordering
│   ├── security/
│   │   ├── linter.go             # Rule execution engine
│   │   ├── rules_aws.go          # Security rules for AWS (S3, SG, RDS, IAM, EBS)
│   │   ├── rules_gcp.go          # Security rules for GCP (Firewall, Storage, SQL)
│   │   ├── rules_azure.go        # Security rules for Azure (NSG, Storage, SQL)
│   │   └── rules_k8s.go          # Security rules for Kubernetes (Pod security)
│   ├── reporter/
│   │   ├── markdown.go           # GitHub PR comment table generator
│   │   └── terminal.go           # Colorized ANSI terminal report generator
│   └── server/
│       ├── api.go                # Embedded HTTP REST API for web studio
│       └── static.go             # Embedded Vite frontend assets (//go:embed)
├── web/
│   ├── src/
│   │   ├── components/
│   │   │   ├── PlanGraphView.tsx     # React Flow interactive DAG diagram
│   │   │   ├── CustomResourceNode.tsx# Custom node component with action styles
│   │   │   ├── DestructionAlerts.tsx # Critical alerts for replaced/deleted nodes
│   │   │   ├── ResourceDiffModal.tsx # Side-by-side attribute before/after diff
│   │   │   ├── SecurityAuditList.tsx # Categorized pre-flight security findings
│   │   │   ├── PlanDropzone.tsx      # Drag and drop upload for JSON plans
│   │   │   └── StatCards.tsx         # Overview metrics cards
│   │   ├── samples/
│   │   │   └── samplePlans.ts        # Built-in realistic sample plans
│   │   ├── types/
│   │   │   └── index.ts              # TypeScript models matching Go API
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── index.css
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── fixtures/
│   ├── aws_production_ecommerce.json # Real-world plan with RDS recreation
│   ├── gcp_microservices.json        # Real-world GCP plan with open firewall
│   └── clean_infrastructure.json     # Clean plan without security flaws
├── docs/
│   ├── ARCHITECTURE.md           # Engineering architecture and data flow
│   ├── SECURITY_RULES.md         # Full catalog of security linting rules
│   └── CLI_USAGE.md              # CLI options and CI/CD integration guide
├── go.mod
├── go.sum
└── README.md
```

---

## 3. Implementation Roadmap & Task Breakdown

### Phase 1: Go Core Foundation & Plan Ingestion
- [x] Task 1.1: Initialize Go module `github.com/alexandrmotologa/tofuforge`.
- [x] Task 1.2: Define unified schema models (`internal/model/plan.go` and `internal/model/analysis.go`).
- [x] Task 1.3: Implement `internal/parser/plan_parser.go` for Terraform and OpenTofu JSON plans.
- [x] Task 1.4: Implement `internal/parser/actions.go` with exact action classifiers (`create`, `update`, `delete`, `replace`, `read`, `no-op`).
- [x] Task 1.5: Implement `internal/parser/statefulness.go` classifying databases, storage, VPCs, subnets, route tables, and KMS keys across AWS, GCP, Azure, and K8s.
- [x] Task 1.6: Write unit tests for the parser and action classifier.

### Phase 2: Directed Acyclic Graph & Blast-Radius Engine
- [x] Task 2.1: Implement `internal/graph/dag.go` to construct the resource dependency graph.
- [x] Task 2.2: Extract both explicit references (`depends_on`, configuration references) and implicit provider linkages.
- [x] Task 2.3: Implement `internal/graph/blast_radius.go` to traverse downstream dependencies for destroyed/replaced resources.
- [x] Task 2.4: Implement scoring algorithm factoring direct destruction, downstream count, and stateful severity.
- [x] Task 2.5: Implement topological sorting in `internal/graph/topological.go`.
- [x] Task 2.6: Write unit tests verifying cycle detection, topological order, and blast radius calculation.

### Phase 3: Multi-Cloud Pre-Flight Security Linter
- [x] Task 3.1: Implement `internal/security/linter.go` rule engine with severities (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`).
- [x] Task 3.2: Implement AWS rules (`rules_aws.go`): open security groups, unencrypted S3/RDS/EBS, S3 public access block, IAM wildcard.
- [x] Task 3.3: Implement GCP rules (`rules_gcp.go`): open firewall ingress, public storage, public SQL, public IAM bindings.
- [x] Task 3.4: Implement Azure rules (`rules_azure.go`): open NSG rules, unencrypted storage, public blob access, unencrypted SQL.
- [x] Task 3.5: Implement Kubernetes rules (`rules_k8s.go`): privileged containers, root execution, missing resource limits.
- [x] Task 3.6: Write unit tests for all security rules across realistic resource changes.

### Phase 4: Reporters & CLI Implementation
- [x] Task 4.1: Implement `internal/reporter/terminal.go` for ANSI color-coded terminal reports with risk scores.
- [x] Task 4.2: Implement `internal/reporter/markdown.go` for clean GitHub PR comment generation.
- [x] Task 4.3: Implement CLI commands in `cmd/tofuforge/main.go`:
  - `inspect <plan.json>`: Runs analysis, outputs terminal or JSON, supports CI failure flags.
  - `pr-comment <plan.json>`: Generates Markdown PR summary to stdout or file.
  - `serve <plan.json>`: Starts embedded web server on specified port.
- [x] Task 4.4: Add CI gating flags: `--fail-on-destructive`, `--fail-on-security-critical`, `--max-blast-radius <int>`.

### Phase 5: Web Studio Frontend (React 19 + React Flow + Tailwind CSS)
- [x] Task 5.1: Scaffold React 19 application in `web/` with Vite, Tailwind CSS, Lucide icons, and React Flow.
- [x] Task 5.2: Create `PlanGraphView.tsx` with custom node rendering, action badges, and color indicators.
- [x] Task 5.3: Implement interactive cascade highlighter: clicking a node highlights its entire downstream blast radius.
- [x] Task 5.4: Build `DestructionAlerts.tsx` highlighting critical destructive replacements.
- [x] Task 5.5: Build `SecurityAuditList.tsx` displaying grouped security findings with severity filters.
- [x] Task 5.6: Build `ResourceDiffModal.tsx` showing granular attribute before and after diffs with replacement triggers.
- [x] Task 5.7: Build `PlanDropzone.tsx` allowing drag-and-drop ingestion of any plan file directly in the browser.
- [x] Task 5.8: Include pre-loaded realistic sample scenarios for instant demonstration.
- [x] Task 5.9: Embed Vite build output into Go binary with `//go:embed`.

### Phase 6: Documentation, Test Fixtures & Validation
- [x] Task 6.1: Create realistic test plan fixtures in `fixtures/` (AWS e-commerce destruction, GCP microservices, clean plan).
- [x] Task 6.2: Write comprehensive documentation:
  - `README.md` with architecture diagram, quickstart, CLI reference, and screenshots.
  - `docs/ARCHITECTURE.md` detailing internal components and DAG traversal.
  - `docs/SECURITY_RULES.md` listing all security checks with remediation instructions.
  - `docs/CLI_USAGE.md` covering CI/CD pipelines (GitHub Actions, GitLab CI).
- [x] Task 6.3: Run full verification suite (Go tests, React build, binary build, CLI execution).
- [x] Task 6.4: Capture authentic visual assets (UI screenshot, terminal run, animated demo) and trigger Phase 2 portfolio sync.
