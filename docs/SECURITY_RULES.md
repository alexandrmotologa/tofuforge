# TofuForge Security Rules Catalog

TofuForge includes a pre-flight static analysis engine designed to evaluate Terraform and OpenTofu plans before changes are applied to cloud providers.

---

## AWS Rules

### AWS-SEC-001: Security Group Allows Open Ingress on Sensitive Ports
- **Severity**: `CRITICAL`
- **Target Resources**: `aws_security_group`, `aws_security_group_rule`, `aws_vpc_security_group_ingress_rule`
- **Description**: Identifies ingress rules allowing `0.0.0.0/0` or `::/0` on sensitive management and database ports:
  - Port 22: SSH
  - Port 3389: RDP
  - Port 5432: PostgreSQL
  - Port 3306: MySQL
  - Port 1433: MSSQL
  - Port 27017: MongoDB
  - Port 6379: Redis
- **Remediation**: Scope CIDR blocks to internal corporate CIDRs or trusted VPN bastion IPs.

### AWS-SEC-002: Security Group Ingress Allows All Traffic From Internet
- **Severity**: `CRITICAL`
- **Target Resources**: `aws_security_group`, `aws_security_group_rule`, `aws_vpc_security_group_ingress_rule`
- **Description**: Identifies ingress rules where protocol is set to `-1` (all) or port range covers 0 to 65535 with source CIDR `0.0.0.0/0`.
- **Remediation**: Explicitly restrict protocols and port ranges required by the application.

### AWS-SEC-003: S3 Bucket Missing Server-Side Encryption
- **Severity**: `HIGH`
- **Target Resources**: `aws_s3_bucket`
- **Description**: Flags S3 buckets created without server-side encryption configuration (`server_side_encryption_configuration`).
- **Remediation**: Attach an `aws_s3_bucket_server_side_encryption_configuration` resource using AES256 or AWS KMS.

### AWS-SEC-004: RDS Database Publicly Accessible
- **Severity**: `CRITICAL`
- **Target Resources**: `aws_db_instance`, `aws_rds_cluster_instance`
- **Description**: Detects database instances configured with `publicly_accessible = true`, which assigns a public IPv4 address.
- **Remediation**: Set `publicly_accessible = false` and locate the database in private subnets.

### AWS-SEC-005: RDS Database Storage Encryption Disabled
- **Severity**: `HIGH`
- **Target Resources**: `aws_db_instance`, `aws_rds_cluster`
- **Description**: Flags database instances with `storage_encrypted = false`.
- **Remediation**: Enable `storage_encrypted = true` and specify a KMS key ID.

### AWS-SEC-006: IAM Policy Contains Wildcard Action
- **Severity**: `HIGH`
- **Target Resources**: `aws_iam_policy`, `aws_iam_role_policy`, `aws_iam_user_policy`
- **Description**: Flags IAM policy statements where `Action` contains `"*"`.
- **Remediation**: Follow least privilege by declaring specific API actions (e.g. `s3:GetObject`).

### AWS-SEC-007: EBS Volume Encryption Disabled
- **Severity**: `MEDIUM`
- **Target Resources**: `aws_ebs_volume`
- **Description**: Detects unencrypted EBS storage volumes (`encrypted = false`).
- **Remediation**: Set `encrypted = true` on the volume resource.

---

## GCP Rules

### GCP-SEC-001: GCP Firewall Allows Public Ingress on Sensitive Ports
- **Severity**: `CRITICAL`
- **Target Resources**: `google_compute_firewall`
- **Description**: Identifies firewall ingress rules allowing traffic from `0.0.0.0/0` to sensitive administrative and database ports (22, 3389, 5432, 3306).
- **Remediation**: Limit source ranges to internal VPC ranges or Identity-Aware Proxy (IAP) CIDRs (`35.235.240.0/20`).

### GCP-SEC-002: Cloud SQL Instance Has Public IP With Open Ingress
- **Severity**: `CRITICAL`
- **Target Resources**: `google_sql_database_instance`
- **Description**: Identifies instances with `ipv4_enabled = true` and authorized network entries matching `0.0.0.0/0`.
- **Remediation**: Use private IP connectivity with Google Cloud SQL Auth Proxy or VPC Service Networking.

### GCP-SEC-003: IAM Policy Grants Role To Public Internet
- **Severity**: `CRITICAL`
- **Target Resources**: All Google IAM member and binding resources
- **Description**: Flags policies granting permissions to `allUsers` or `allAuthenticatedUsers`.
- **Remediation**: Remove public identities from sensitive IAM roles.

### GCP-SEC-004: Cloud Storage Uniform Bucket-Level Access Disabled
- **Severity**: `MEDIUM`
- **Target Resources**: `google_storage_bucket`
- **Description**: Flags buckets where `uniform_bucket_level_access = false`.
- **Remediation**: Enable uniform bucket-level access to prevent object ACL discrepancies.

---

## Azure Rules

### AZR-SEC-001: Network Security Group Inbound Allows Internet to Sensitive Ports
- **Severity**: `CRITICAL`
- **Target Resources**: `azurerm_network_security_group`, `azurerm_network_security_rule`
- **Description**: Flags inbound rules permitting traffic from `*` or `Internet` to sensitive ports.
- **Remediation**: Restrict source addresses to corporate Virtual Network CIDRs.

### AZR-SEC-002: Storage Account Allows Insecure HTTP Traffic
- **Severity**: `HIGH`
- **Target Resources**: `azurerm_storage_account`
- **Description**: Identifies storage accounts with `enable_https_traffic_only = false`.
- **Remediation**: Set `enable_https_traffic_only = true`.

### AZR-SEC-003: Storage Account Allows Public Blob Access
- **Severity**: `HIGH`
- **Target Resources**: `azurerm_storage_account`
- **Description**: Flags accounts where `allow_blob_public_access` or `allow_nested_items_to_be_public` is true.
- **Remediation**: Enforce private access by setting public access attributes to false.

### AZR-SEC-004: MSSQL Database Transparent Data Encryption Disabled
- **Severity**: `HIGH`
- **Target Resources**: `azurerm_mssql_database`
- **Description**: Detects databases where `transparent_data_encryption_enabled = false`.
- **Remediation**: Set `transparent_data_encryption_enabled = true`.

---

## Kubernetes Rules

### K8S-SEC-001: Kubernetes Container Configured in Privileged Mode
- **Severity**: `CRITICAL`
- **Target Resources**: `kubernetes_pod`, `kubernetes_deployment`, `kubernetes_stateful_set`
- **Description**: Identifies containers where `security_context.privileged = true`.
- **Remediation**: Disable privileged mode and grant minimal Linux capabilities instead.

### K8S-SEC-002: Kubernetes Container Permitted to Run as Root
- **Severity**: `HIGH`
- **Target Resources**: `kubernetes_pod`, `kubernetes_deployment`, `kubernetes_stateful_set`
- **Description**: Identifies containers where `run_as_non_root = false` or `run_as_user = 0`.
- **Remediation**: Set `run_as_non_root = true` and specify a non-zero UID.

### K8S-SEC-003: Kubernetes Workload Missing Resource Limits
- **Severity**: `MEDIUM`
- **Target Resources**: `kubernetes_pod`, `kubernetes_deployment`, `kubernetes_stateful_set`
- **Description**: Flags container specifications that omit CPU or memory resource limits.
- **Remediation**: Specify `resources.limits.cpu` and `resources.limits.memory`.
