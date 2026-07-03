### To any poor souls who actually read this

Disregard this entire document. This is just notes for what I need to do next, or scraps to add to the README.
omg why are you still reading this? save yourself, look away! <3

---------------


---------------

IAM user and identity matrix
aws cli setup
aws session manager plugin
config file downloaded and moved into root folder we're working from, cli binary moved to /usr/local/bin/ folder for terminal use


---------------

need to add notes about initializing local aws provider plugins for terraform in the modules folder
terraform -chdir=./modules/idp-env init


and then later talking about the modules .hcl files, how they'd be included with the CLI tool; which explains the above away
Template Distribution & Packaging Architecture
Architecture Choice: Local Directory Referencing vs. Native Binary Embedding / Remote Git Modules.
Current State (MVP): The CLI targets the local disk path ./modules/idp-env for rapid local prototyping and debugging by a single systems administrator.
Production Distribution Plan: To deliver a zero-footprint developer experience and completely obscure infrastructure source code from end-users, the platform would transition to Go Native Embedding (go build with the embed library) or configure Terraform to resolve source paths using a Private Git Repository Remote. This enables the distribution of a single, standalone binary to developer workstations while maintaining absolute centralized control of infrastructure logic.

---------------

rds is a cost sink and aws takes 5-10 minutes to spin up new RDS instances.



Database Layer Design DecisionArchitecture Choice: Containerized Local Databases (Docker) vs. Managed AWS RDS.Decision: Intentionally opted out of dedicated AWS RDS instances per developer sandbox.Justification: For ephemeral crash-testing, RDS introduces a 7-minute provisioning latency and a ~$15/month idle cost penalty per environment.Implementation: The IDP provisions an EC2 node pre-configured for multi-container runtimes. Developers spin up their application and database side-by-side using docker-compose. This cuts environment creation time to under 45 seconds and maintains a near-zero cost profile.

------------------------------------

The Reality: Logs are Already LocalSince your developers are connecting to the instance using AWS Systems Manager (SSM), they are dropped directly into a secure shell on the EC2 machine. They have instant, real-time access to everything they need:If their app sucks and crashes a container, they just run docker logs <container_id> or docker compose logs.If the Docker daemon itself crashes, they can look at the system logs directly using sudo journalctl -u docker.Because the environment is actively being used for a live debugging session, there is absolutely no reason to ship those logs out to AWS.

Log Aggregation and ObservabilityArchitecture Choice: Local Host Observability vs. Centralized CloudWatch Streaming.Decision: Intentionally bypassed installing CloudWatch log aggregation agents on the ephemeral instances.Justification: Centralized log streaming adds ~$0.50/GB in AWS ingestion costs and slows down instance bootstrap speeds.Implementation: Because developers maintain direct, secure shell access via AWS Systems Manager, observability is handled natively on the host using standard Linux commands (docker compose logs and journalctl). If an environment experiences a catastrophic failure, the platform philosophy dictates destroying it and recreating it via the CLI rather than debugging an ephemeral host indefinitely.


Documentation Blueprint: Observability Architecture
Status: Infrastructure Ready / Application Out-of-Scope
The Goal: The IDP is fully capable of bootstrapping a telemetry stack (Prometheus + Grafana) alongside ephemeral developer subnets using Dashboard-as-Code.
Engineering Decision: To minimize compute costs and maintain focus on core platform engineering mechanics (subnet provisioning, automated lifecycle teardown, and IAM scoping), active application performance monitoring (APM) is marked as out-of-scope for Phase 1.
Proof of Concept: The infrastructure pipeline successfully deploys Node Exporter on the developer EC2 instance to mimic infrastructure load. Below are screenshots of the automated Grafana dashboard validating the end-to-end data pipeline before the ephemeral environment is torn down:
[Insert your Grafana screenshots here]
Next Steps for Production Rollout: To fully onboard an engineering team, developers must expose a /metrics endpoint on port 8080 in their deployment manifests, allowing the platform's central Prometheus server to auto-discover their workload.

-----------------------

why have --list return values from local tf state file when it won't work like that in production? ideally, it'd use AWS CLI to query for tags created by this CLI tool to return to the user. AI Overlord helped me cook up a first draft for how to explain this design choice for this project. (aws ec2 describe-instances --filters "Name=tag:ManagedBy,Values=IDP-CLI")
### 📋 Platform Visibility & State Sync
**Architecture Choice**: Local Directory Scanning vs. Remote S3 Object Inventory.
* **Current State (MVP)**: The `--list` command queries the local `terraform.tfstate.d` workspace directory for state receipts. This provides a zero-latency, cost-free demonstration environment for single-operator testing.
* **Production Scaling Plan**: In a multi-developer enterprise environment, this function would be refactored to utilize the AWS SDK to scan the shared S3 remote state bucket prefix or run an `ec2:DescribeInstances` query filtered by the `ManagedBy = "IDP-CLI"` tag. This ensures a single, centralized source of truth across all developer workstations without local state fragmentation.


------------------------

reaper or ttl automated janitor
the cost containment protocol
"For this MVP project, I used a localized EC2 cron-bomb to stop compute costs and kept the architecture simple to avoid a $30/month AWS Load Balancer fee. However, because I know developers always forget to clean up, I've outlined the exact enterprise strategy in my ADRs. In a real production environment, you would swap my local strategy for a central AWS Lambda 'Reaper' function that scans lifecycle tags and triggers a total cleanup. I even left a placeholder architectural file in the repo to show exactly how that script would map out."

--------------------

intro?
"An automated Internal Developer Platform (IDP) suite that enables engineers to spin up identity-secured, zero-ingress cloud sandboxes with a single command. Built using Go as the execution orchestration binary, Terraform as the immutable infrastructure engine, and Ansible as the OS hardening and container deployment framework."

2. The Architectural Design Records (ADRs)This is where you address the advanced topics we discussed (Remote state vs. config files, cross-compilation pipelines, Shared runways vs. dynamic subnets). Instead of weaving these long paragraphs into your installation steps, group them under a dedicated "Architectural Decisions" section.Use bold bullet points to state the Constraint, your MVP Choice, and your Production Scaling Plan. This keeps the reading punchy and incredibly high-utility.

3. Onboarding & InstallationKeep this strictly restricted to the terminal code-blocks we mapped out. Developers just want to copy-paste commands to get the tool running; they don't want to read essays while configuring their path.



### Prerequisites: Local Session Manager Plugin (Ubuntu/Linux Mint)
Because this platform utilizes identity-based tunneling instead of legacy SSH key pairs, developers must install the AWS Session Manager plugin locally to access their sandboxes.

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


-----

another installation instruction template for later
## 📦 Developer Installation & Onboarding

Because this tool compiles natively down to pure machine code, developers must build the binary locally to ensure perfect compatibility with their workstation hardware (Linux/macOS/Windows).

### 1. Clone the Platform Universe
```bash
git clone https://github.com
cd idp-platform
```

### 2. Initialize Local Network Boundaries
The Platform Team manages core network shielding centrally. To bridge your CLI tool to the active corporate testing domain, create a hidden configuration file at the root of this project:

```bash
cat <<EOF > .idp-config.json
{
  "vpc_id": "vpc-YOUR_CORPORATE_VPC_ID",
  "subnet_id": "subnet-YOUR_TARGET_RUNWAY_ID"
}
EOF
```
*(Note: Ask your Platform Administrator for the active AWS VPC and Subnet string tokens).*

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

> 💡 **Production Note**: For the purposes of this MVP portfolio demonstration, the tool utilizes a source-level compilation workflow (`go build`). In an enterprise deployment, this repository would configure a **CI/CD Pipeline (GitHub Actions)** utilizing Go's native cross-compilation engines (`GOOS`/`GOARCH`) to automatically publish pre-compiled standalone binary packages for Linux, macOS, and Windows directly to the GitHub Releases runway, requiring zero local Go dependencies for end-user developers.



---

part of setting up for ansible
"Step 1: Ensure you have a standard local SSH identity key generated. If your ~/.ssh/ folder is empty, run ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa -N "" once before launching the CLI."