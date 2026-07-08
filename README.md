# InternalDeveloperPlatform
Let's make IDP's with a CLI tool written in `Go`!

## Overview
To better understand the relationships between Infrastructure as Code (IaC) and CLI tooling, this project aims to make automating developer environments simple and cost efficient. 

Who doesn't want their own personal development env to test code in?

---

### How It Works
1. The Foundation Terraform files are run once to simulate an existing company environment. The `VPC` and public subnet represent a predefined Test or Dev environment.
2. The CLI tool is invoked to create an EC2 on the public subnet where the developer may deploy and test their code, their own personal IDP. This subnet has a security group attached that allows zero ingress of networking traffic outside of what is allowed via `AWS SSM`.[^1]
3. This new IDP can now be accessed by the developer to deploy code as a docker image and test in a mock-production environment. Local Session Manager Plugin required to be installed locally to access IDP's.[^2]
4. When finished, the CLI tool can also be used to tear down any IDP’s no longer needed.
5. [Installation and Use Instructions](#use) below

---

### Directory Structure

Foundation folder contains the underlying infrastructure, simulating what would be found in a corporate environment. The `VPC`, an internet gateway, a single public subnet, and an `S3 bucket` along with `DynamoDB` table to track and lock state of this and subsequent infrastructure. Ran once to create the necessary infrastructure, then unused thereafter except to tear down this project's underlying pieces. 

Modules folder contains any infrastructure or tooling that the CLI will later implement. `idp-env` holds the CLI functionality for building individual developer internal development platforms. 

Ansible folder contains the `Ansible` playbook and supporting files necessary to integrate itself into the CLI tool. Using `AWS SSM`, it would bypass the need for an SSH connection, install dependencies, in this case Docker. This was included to show how this tool would be used properly in a production setting, deprecated in this Minimum Viable Product, explained further in the [ADR](#ADR).

`Go` code for the CLI functionality is located in the root directory, `main.go`.

Further folders may be added as functionality expands for the CLI.

<a id="use"></a>
<details>
<summary><b>Click to view Local Installation & Use Instructions</b></summary> 

## Installation and Use 

Make sure Golang is installed and AWS CLI is configured locally (`aws configure`). The configured AWS profile must have an IAM policy attached that grants sufficient permissions to manage the specific resources defined in the Terraform files (e.g., EC2). For testing, an Administrator access policy can be used, but for production, ensure the profile adheres to the principle of least privilege for the target infrastructure. 

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

## 📦 Developer Installation & Onboarding

Because this tool compiles natively down to pure machine code, developers must build the binary locally to ensure perfect compatibility with their workstation hardware (Linux/macOS/Windows).

### 1. Clone the Platform Universe
```bash
git clone https://github.com/campbelllu/InternalDeveloperPlatform.git
cd idp-platform
```

### 2. Initialize Local Network Boundaries
The Platform Team manages core network shielding centrally. To bridge your CLI tool to the active corporate testing domain, create a hidden configuration file at the root of this project:

Verify your target `VPC` and public subnet exist, or make them with the Foundation .hcl files. Note the `VPC` and subnet ID’s for later use.

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

You can now make IDP’s via: `idp –name YOUR-CHOSEN-ENV-NAME`
You can see all environments currently active with: `idp –list’
You can tear down environments with: `idp –destroy –name YOUR-CHOSEN-ENV-NAME`

You can also verify that `Docker` is installed and running with the following; replace the indicated line with your instance ID, first:

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
This section will expand upon the following:
Why not include RDS instances in the IDP's?
Why AWS Lambda was not implemented for TTL?
Why no logging or monitoring out of the box?

> 💡 **Production Note**: For the purposes of this MVP portfolio demonstration, the tool utilizes a source-level compilation workflow (`go build`). In an enterprise deployment, this repository would configure a **CI/CD Pipeline (GitHub Actions)** utilizing Go's native cross-compilation engines (`GOOS`/`GOARCH`) to automatically publish pre-compiled standalone binary packages for Linux, macOS, and Windows directly to the GitHub Releases runway, requiring zero local Go dependencies for end-user developers.

---

### Platform Roadmap & Enterprise Scalability

This architecture represents a high-performance Minimum Viable Product (MVP) optimized for local developer speed and lightweight cloud costs.

For large-scale enterprise environments, the platform is designed to scale across the following vectors:

* **Centralized Secrets Governance (Vault):** The configuration file lookup engine (`.idp-config.json`) is built currently to read from local disk, but could transition seamlessly to authenticated API secret fetches via HashiCorp Vault or AWS Secrets Manager.
* **Orchestration Scaling (EKS/Kubernetes):** While this sandbox relies on standalone Docker daemons for extreme cost savings and rapid teardowns, the Go orchestration logic can be refactored to hand images off to an AWS EKS cluster using the native AWS Go SDK.

---

### References <a id="references"></a>
[^1]: This IDP, the subnet created, could be configured to be a private subnet, which would require a `NAT Gateway` present on the `VPC`, and this was avoided for this project to keep costs low during production and testing. `VPN`'s could also be used to grant access to these IDP's, but setting up a corporate `VPN` for these purposes is outside of the scope of this project.
[^2]: Steps to be included here.

---

