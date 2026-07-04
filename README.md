# InternalDeveloperPlatform
Let's make IDP's with a CLI tool written in `Go`!

## Overview
To better understand the relationships between Infrastructure as Code (IaC) and CLI tooling, this project aims to make automating developer environments simple and cost efficient. 

Who doesn't want their own personal development env to test code in?

---

### How It Works
1. The Foundation Terraform files are run to simulate a company environment, setting up the 'foundation' upon which all developers would be working anyway.
2. The CLI tool is invoked to create a public subnet on the `VPC` where the developer may deploy and test their code. This subnet has a security group attached that allows zero ingress of networking traffic outside of what is allowed via `AWS SSM`.[^1]
3. This new IDP can now be accessed by the developer to deploy code and test in a mock-production environment. Local Session Manager Plugin required to be installed locally to access IDP's.[^2]
4. More steps to come.
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

### Installation and Use 

Stuff goes here, so you can use it. and stuff. can't forget the stuff.

</details>

---

### Architectural Decision Record \ Architecture Design Record <a id="ADR"></a>
This section will expand upon the following:
Why not include RDS instances in the IDP's?
Why AWS Lambda was not implemented for TTL?
Why no logging or monitoring out of the box?

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

