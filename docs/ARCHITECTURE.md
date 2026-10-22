# TofuForge Architecture & Data Flow

TofuForge is a static analysis and visualization platform for Terraform and OpenTofu execution plans. It processes JSON plan output, computes resource dependency graphs (DAGs), identifies destructive resource recreation, performs pre-flight cloud security auditing, and renders interactive visual graphs.

---

## 1. System Components

```
+---------------------------------------------------------------------------------+
|                                    TofuForge                                    |
+---------------------------------------------------------------------------------+
                                         |
     +-----------------------------------+-----------------------------------+
     |                                                                       |
+----v-----------------------------+                   +---------------------v---+
|       Ingestion & Parser         |                   |      DAG & Blast Radius |
| - Unmarshal Terraform JSON Plan  |                   | - Directed Acyclic Graph|
| - Action Classifier (CRUD/Repl)  |                   | - Transitive Downstream |
| - Statefulness Evaluator         |                   | - Cumulative Risk Score |
+----+-----------------------------+                   +---------------------+---+
     |                                                                       |
     +-----------------------------------+-----------------------------------+
                                         |
                       +-----------------v-----------------+
                       |    Pre-Flight Security Engine     |
                       | - AWS Rules (SG, S3, RDS, IAM)    |
                       | - GCP Rules (Firewall, SQL, IAM)  |
                       | - Azure Rules (NSG, Storage, SQL) |
                       | - K8s Rules (Privileged, Root)    |
                       +-----------------+-----------------+
                                         |
                       +-----------------v-----------------+
                       |       Reporters & Web Studio      |
                       | - Terminal Colorized ANSI Output  |
                       | - Compact PR Markdown Generator   |
                       | - React 19 Interactive Web Studio |
                       | - Single Binary Go //go:embed     |
                       +-----------------------------------+
```

---

## 2. Plan Ingestion & Classification

Terraform and OpenTofu generate plan files that can be exported to JSON via:
```bash
terraform show -json tfplan.binary > tfplan.json
# or
tofu show -json tfplan.binary > tfplan.json
```

TofuForge unmarshals this JSON into internal models (`internal/model/plan.go`). Each item in `resource_changes` is processed:
1. **Action Categorization**:
   - `["no-op"]` -> `ActionNoOp`
   - `["create"]` -> `ActionCreate`
   - `["update"]` -> `ActionUpdate`
   - `["delete"]` -> `ActionDelete`
   - `["delete", "create"]` or `["create", "delete"]` -> `ActionReplace`
2. **Statefulness Classification**:
   - Resources that manage durable persistent data (RDS, DynamoDB, CloudSQL, Spanner, S3, PersistentVolumeClaims) or foundational network routing (VPC, Transit Gateways, Subnets, KMS keys) are classified as stateful.
   - Any stateful resource scheduled for `delete` or `replace` triggers a high-severity alert.
3. **Attribute Diffing**:
   - Compares `change.before` and `change.after`.
   - Correlates attributes with `change.replace_paths` to flag attributes that caused Terraform to force a full resource recreation.

---

## 3. Directed Acyclic Graph (DAG) & Blast Radius

Infrastructure changes often have unintended ripple effects. Modifying a subnet or database configuration can force dependent workloads or DNS records to update or fail.

### 3.1 Edge Extraction
TofuForge constructs directed edges (`Source -> Target`, meaning `Target` depends on `Source`):
- **Explicit configuration**: Extracted from `configuration.root_module.resources[].depends_on` and expression references (`references`).
- **Implicit cloud relationships**: Fallback heuristics for infrastructure linkages (e.g., Subnet to VPC, Security Group Rule to Security Group, DB Instance to Subnet, KMS Key to S3/RDS).

### 3.2 Blast Radius Calculation
For each resource undergoing destruction or recreation:
1. Breadth-First Search (BFS) traverses outgoing edges to compute the set of all transitively reachable downstream nodes.
2. A Blast Radius Score is calculated:
   - Base weight: 25 points for stateful resources, 5 points for non-stateful resources.
   - Downstream penalty: 15 points per stateful downstream dependent, 4 points per non-stateful downstream dependent.
   - Additional penalty for recreation (`replace`) over clean deletion.
3. Resources with high blast radius or stateful destruction generate high-priority `DestructionAlert` items.

---

## 4. Pre-Flight Security Engine

The security engine inspects planned states (`change.after`) before deployment. It identifies misconfigurations across major cloud providers:
- **AWS**: Public ingress (0.0.0.0/0) on sensitive ports (SSH 22, RDP 3389, DB ports), unencrypted S3 buckets, publicly accessible RDS instances, wildcard IAM policies.
- **GCP**: Unrestricted firewall ingress, Cloud SQL with public IPs and open authorized networks, IAM bindings granting roles to `allUsers`.
- **Azure**: Open NSG inbound rules, storage accounts without HTTPS enforcement, unencrypted SQL databases.
- **Kubernetes**: Privileged containers, root execution, missing compute limits.

Findings include severity levels (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`), exact resource addresses, descriptions, and concrete remediation instructions.

---

## 5. Web Studio & Embedded Delivery

The frontend is built using React 19, React Flow (`@xyflow/react`), Tailwind CSS, and Lucide icons.
- **Interactive Graph**: Visualizes resources as color-coded nodes based on action (Green: Create, Yellow: Update, Red: Delete, Purple: Replace).
- **Cascade Highlighter**: Selecting any node dynamically illuminates its downstream blast radius cascade while dimming unaffected components.
- **Single Static Binary**: The compiled frontend is embedded into the Go executable via `//go:embed`, allowing deployment without external runtime dependencies.
