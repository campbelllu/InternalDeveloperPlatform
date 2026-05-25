# InternalDeveloperPlatform
Let's make IDP's with a CLI tool written in `Go`!

## Overview
To better understand the relationships between Infrastructure as Code (IaC) and CLI tooling, this project aims to make automating developer environments simple and cost efficient. 

Who doesn't want their own personal development env to test code in?

### Directory Structure

Foundation folder contains the underlying infrastructure, simulating what would be found in a corporate environment. The `VPC`, an internet gateway, a single public subnet, and an `S3 bucket` along with `DynamoDB` table to track and lock state of this and subsequent infrastructure. Ran once to create the necessary infrastructure, then unused thereafter except to tear down this project's underlying pieces. 

---

Modules folder contains any infrastructure or tooling that the CLI will later implement. `idp-env` holds the original CLI functionality for building individual developer internal development platforms. Further folders may be added as functionality expands for the CLI.

---

CLI folder will contain relevant Golang code that is the actual CLI tool.

---

Further folders will contain other IaC resources as they are added for functionality. 



### How It Works
1. The Foundation Terraform files are run to simulate a company environment, setting up the 'foundation' upon which all developers would be working anyway.
2. The CLI tool is invoked to create a public subnet on the `VPC` where the developer may deploy and test their code. This subnet has a security group attached that allows zero ingress of networking traffic outside of what is allowed via `AWS SSM`. [^1]
3. This new IDP can now be accessed by the developer to deploy code and test in a mock-production environment.
4. More steps to come.


### References
[^1] This IDP, the subnet created, could be configured to be a private subnet, which would require a `NAT Gateway` present on the `VPC`, and this was avoided for this project to keep costs low during production and testing. `VPN`'s could also be used to grant access to these IDP's, but setting up a corporate `VPN` for these purposes is outside of the scope of this project.