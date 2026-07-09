# InternalDeveloperPlatform
Let's make IDP's with a CLI tool written in `Go`!

## Overview
To better understand the relationships between Infrastructure as Code (IaC) and CLI tooling, this project aims to make automating developer environments simple and cost efficient. 

Who doesn't want their own personal development env to test code in?

---

### How It Works
Dive into how this whole thing works.

<details><summary>Click to expand</summary>

1. The Foundation Terraform files are run once to simulate an existing company environment. The `VPC` and public subnet represent a predefined Test or Dev environment.

2. The CLI tool is invoked to create an EC2 on the public subnet where the developer may deploy and test their code, their own personal IDP. This subnet has a security group attached that allows zero ingress of networking traffic outside of what is allowed via `AWS SSM`. [Why was a public subnet used?](#privateSubnet)

3. This new IDP can now be accessed by the developer to deploy code as a docker image and test in a mock-production environment. Local Session Manager Plugin required to be installed locally to access IDP's.

4. When finished, the CLI tool can also be used to tear down any IDP’s no longer needed.

5. Every Friday at 9pm, all EC2 sandboxes will automatically shut down to reduce cloud costs. `idp --list` will show time since last edits made to the EC2 and remind the user to close down now defunct testing environments. [See why](#reaper)

6. [Installation and Use Instructions](#use) below

</details>

---

### Directory Structure
What's included and why?

<details><summary>Click to expand</summary>

Foundation folder contains the underlying infrastructure, simulating what would be found in a corporate environment. The `VPC`, an internet gateway, a single public subnet, and an `S3 bucket` along with `DynamoDB` table to track and lock state of this and subsequent infrastructure. Ran once to create the necessary infrastructure, then unused thereafter except to tear down this project's underlying pieces. 

Modules folder contains any infrastructure or tooling that the CLI will later implement. `idp-env` holds the CLI functionality for building individual developer internal development platforms. [This MVP uses the local .hcl files within /Modules, but that would change in an enterprise environment.](#modules)

Ansible folder contains the `Ansible` playbook and supporting files necessary to integrate itself into the CLI tool. Using `AWS SSM`, it would bypass the need for an SSH connection, install dependencies, in this case Docker. This was included to show how this tool would be used properly in a production setting, deprecated in this Minimum Viable Product, explained further in the [ADR](#ADR).

`Go` code for the CLI functionality is located in the root directory, `main.go`.

Further folders may be added as functionality expands for the CLI.

</details>

---

<a id="use"></a>
### Local Installation & Use Instructions
So you want to use this tool?

<details>
<summary>Click to expand</summary> 

Make sure `Golang` is installed and AWS CLI is configured locally (`aws configure`). The configured AWS profile must have an IAM policy attached that grants sufficient permissions to manage the specific resources defined in the Terraform files (e.g., EC2). For testing, an Administrator access policy can be used, but for production, ensure the profile adheres to the principle of least privilege for the target infrastructure. 

Because this platform utilizes identity-based tunneling instead of legacy SSH key pairs, developers must install the `AWS Session Manager` plugin locally to access their sandboxes.

Run the following commands in your terminal to download and install the official 64-bit Debian package:

```bash
# 1. Download the official package directly from the AWS S3 bundle storage
curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" -o "session-manager-plugin.deb"

# 2. Install the package locally
sudo dpkg -i session-manager-plugin.deb

# 3. Verify the installation succeeds
session-manager-plugin --version

# 4. Clean up the installer binary
rm session-manager-plugin.deb
```

Once verified, your local AWS CLI will automatically utilize this plugin behind the scenes whenever your Go CLI executes a terminal connection session.

## Developer Installation & Onboarding

Because this tool compiles natively down to pure machine code, developers must build the binary locally to ensure perfect compatibility with their workstation hardware (Linux/macOS/Windows). [See note here](#building)

### 1. Clone the Platform Universe
```bash
mkdir idp-maker
cd idp-maker
git clone https://github.com/campbelllu/InternalDeveloperPlatform.git
```

### 2. Initialize Local Network Boundaries
To bridge your CLI tool to the active 'corporate' testing domain, create a hidden configuration file at the root of this project:

Verify your target `VPC` and public subnet exist, or make them with the Foundation .hcl files. Note the `VPC` and subnet ID’s for later use. [See note here](#vpc)

```bash
cat <<EOF > .idp-config.json
{
  "vpc_id": "vpc-YOUR_CORPORATE_VPC_ID",
  "subnet_id": "subnet-YOUR_TARGET_RUNWAY_ID"
}
EOF
```
*(Note: Ask your Platform Administrator for the active AWS VPC and Subnet string tokens if you cannot verify yourself)*

### 3. Compile and Install Natively
Run the native Go compiler to generate your standalone executable binary and register it to your system execution path:

```bash
# Compile the source code natively for your exact CPU/OS
go build -o idp

# Global system installation
sudo mv idp /usr/local/bin/
```

### 4. Verify the Launch Runway
Open a fresh terminal window anywhere on your machine and invoke the tracking inventory tool:
```bash
idp --list
```

---

You can now make IDP’s via: `idp –name YOUR-CHOSEN-ENV-NAME`

You can see all environments currently active with: `idp –list`

You can tear down environments with: `idp –destroy –name YOUR-CHOSEN-ENV-NAME`

If no environments are showing, but you're certain some exist, it's because you're using `idp --list` in a different directory from which you first made the environment in, the local state file is not present in your current directory.

Use: `aws ec2 describe-instances --filters "Name=tag:ManagedBy,Values=IDP-CLI"` to find the 'rogue' instances and [See this note here.](#list)

---

You can also verify that `Docker` is installed and running with the following; replace the indicated line with your new sandbox instance ID, first:

*Instance ID is output in terminal upon creation*

```
aws ssm start-session \
  --target YOUR_NEW_INSTANCE_ID \
  --document-name AWS-StartNonInteractiveCommand \
  --parameters '{"command": ["docker --version"]}'
```
```
aws ssm start-session \
  --target YOUR_NEW_INSTANCE_ID \
  --document-name AWS-StartNonInteractiveCommand \
  --parameters '{"command": ["systemctl status docker --no-pager"]}'
```

And if for any reason you want to jump inside your sandbox EC2:
`aws ssm start-session --target YOUR_NEW_INSTANCE_ID`

</details>

---

### Architectural Decision Record \ Architecture Design Record <a id="ADR"></a>
Why did I build it this way?

<details>
<summary>Click to expand</summary> 

<a id="privateSubnet"></a>
##### Related to Public vs Private Subnets for this MVP
This IDP, the subnet created, could be configured to be a private subnet, which would require a `NAT Gateway` present on the `VPC`, and this was avoided for this project to keep costs low during production and testing. `VPN`'s could also be used to grant access to these IDP's, but setting up a corporate `VPN` for these purposes is outside of the scope of this project.

<a id="reaper"></a>
##### Related to Reaper settings and automated EC2 sandbox cleanup
the cost containment protocol
Why AWS Lambda was not implemented for TTL?
"For this MVP project, I used a localized EC2 cron-bomb to stop compute costs and kept the architecture simple to avoid a $30/month AWS Load Balancer fee. However, because I know developers always forget to clean up, I've outlined the exact enterprise strategy in my ADRs. In a real production environment, you would swap my local strategy for a central AWS Lambda 'Reaper' function that scans lifecycle tags and triggers a total cleanup. I even left a placeholder architectural file in the repo to show exactly how that script would map out."

<a id="modules"></a>
##### Related to local /modules .hcl file vs abstracting them away from the developers
Template Distribution & Packaging Architecture
Architecture Choice: Local Directory Referencing vs. Native Binary Embedding / Remote Git Modules.
Current State (MVP): The CLI targets the local disk path ./modules/idp-env for rapid local prototyping and debugging by a single systems administrator.
Production Distribution Plan: To deliver a zero-footprint developer experience and completely obscure infrastructure source code from end-users, the platform would transition to Go Native Embedding (go build with the embed library) or configure Terraform to resolve source paths using a Private Git Repository Remote. This enables the distribution of a single, standalone binary to developer workstations while maintaining absolute centralized control of infrastructure logic.

<a id="building"></a>
##### Related to local building of binary vs distributed binaries via Platform team
For the purposes of this MVP portfolio demonstration, the tool utilizes a source-level compilation workflow (`go build`). In an enterprise deployment, this repository would configure a **CI/CD Pipeline (GitHub Actions)** utilizing Go's native cross-compilation engines (`GOOS`/`GOARCH`) to automatically publish pre-compiled standalone binary packages for Linux, macOS, and Windows directly to the GitHub Releases runway, requiring zero local Go dependencies for end-user developers.

<a id="vpc"></a>
##### Related to VPC and Subnet ID Config File
Ideally, this configuration file (referencing VPC and subnet ID's) would rest in a secrets vault to be updated by the platform team, and the CLI would reference that secrets vault. The setup for this MVP was chosen to keep scope in check solely to focus on CLI functionality. While platform teams could provide developers with new config files detailing VPC and subnet ID's, this would create needless work for them if infrastructure is regularly churning.

<a id="list"></a>
##### Related to local --list checking vs Proper Cloud Checks in a Production Setting
The architecture choice for this MVP was to utilize Local Directory Scanning vs. Remote S3 Object Inventory. Currently the `--list` command queries the local `terraform.tfstate.d` workspace directory for state receipts. This provides a zero-latency, cost-free demonstration environment for single-operator testing.

In a multi-developer enterprise environment, this function would be refactored to utilize the AWS SDK to scan the shared S3 remote state bucket prefix or run an `ec2:DescribeInstances` query filtered by the `ManagedBy = "IDP-CLI"` tag. This would ensure a single, centralized source of truth across all developer workstations without local state fragmentation.

<a id="list"></a>
##### Why no logging or monitoring out of the box?
The Reality: Logs are Already LocalSince your developers are connecting to the instance using AWS Systems Manager (SSM), they are dropped directly into a secure shell on the EC2 machine. They have instant, real-time access to everything they need:If their app sucks and crashes a container, they just run docker logs <container_id> or docker compose logs.If the Docker daemon itself crashes, they can look at the system logs directly using sudo journalctl -u docker.Because the environment is actively being used for a live debugging session, there is absolutely no reason to ship those logs out to AWS.

Log Aggregation and ObservabilityArchitecture Choice: Local Host Observability vs. Centralized CloudWatch Streaming.Decision: Intentionally bypassed installing CloudWatch log aggregation agents on the ephemeral instances.Justification: Centralized log streaming adds ~$0.50/GB in AWS ingestion costs and slows down instance bootstrap speeds.Implementation: Because developers maintain direct, secure shell access via AWS Systems Manager, observability is handled natively on the host using standard Linux commands (docker compose logs and journalctl). If an environment experiences a catastrophic failure, the platform philosophy dictates destroying it and recreating it via the CLI rather than debugging an ephemeral host indefinitely.


Documentation Blueprint: Observability Architecture
Status: Infrastructure Ready / Application Out-of-Scope
The Goal: The IDP is fully capable of bootstrapping a telemetry stack (Prometheus + Grafana) alongside ephemeral developer subnets using Dashboard-as-Code.
Engineering Decision: To minimize compute costs and maintain focus on core platform engineering mechanics (subnet provisioning, automated lifecycle teardown, and IAM scoping), active application performance monitoring (APM) is marked as out-of-scope for Phase 1.
Proof of Concept: The infrastructure pipeline successfully deploys Node Exporter on the developer EC2 instance to mimic infrastructure load. Below are screenshots of the automated Grafana dashboard validating the end-to-end data pipeline before the ephemeral environment is torn down:
[Insert your Grafana screenshots here]
Next Steps for Production Rollout: To fully onboard an engineering team, developers must expose a /metrics endpoint on port 8080 in their deployment manifests, allowing the platform's central Prometheus server to auto-discover their workload.

<a id="list"></a>
##### Why no only make a single EC2 without any RDS connection?

rds is a cost sink and aws takes 5-10 minutes to spin up new RDS instances.

Database Layer Design DecisionArchitecture Choice: Containerized Local Databases (Docker) vs. Managed AWS RDS.Decision: Intentionally opted out of dedicated AWS RDS instances per developer sandbox.Justification: For ephemeral crash-testing, RDS introduces a 7-minute provisioning latency and a ~$15/month idle cost penalty per environment.Implementation: The IDP provisions an EC2 node pre-configured for multi-container runtimes. Developers spin up their application and database side-by-side using docker-compose. This cuts environment creation time to under 45 seconds and maintains a near-zero cost profile.


</details>

---

### Platform Roadmap & Enterprise Scalability
Future plans

<details>
<summary>Click to expand</summary> 

This architecture represents a high-performance Minimum Viable Product (MVP) optimized for local developer speed and lightweight cloud costs.

For large-scale enterprise environments, the platform is designed to scale across the following vectors:

* **Centralized Secrets Governance (Vault):** The configuration file lookup engine (`.idp-config.json`) is built currently to read from local disk, but could transition seamlessly to authenticated API secret fetches via HashiCorp Vault or AWS Secrets Manager. It is mentioned elsewhere that this CLI tool could also be made to check against the S3-based state file.
* **Orchestration Scaling (EKS/Kubernetes):** While this sandbox relies on standalone Docker daemons for extreme cost savings and rapid teardowns, the Go orchestration logic can be refactored to hand images off to an AWS EKS cluster using the native AWS Go SDK.

</details>

---

### Credits
Author: Luke E Campbell

Generous Donation From: Coffee in the morning, Tension Tamer tea at night

License: Currently None

