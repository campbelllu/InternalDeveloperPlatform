variable "env_name" {
  type        = string
  description = "The name of this dev environment (e.g. dev-luke)"
}

variable "vpc_id" {
  type        = string
  description = "The ID of our Foundation VPC"
} 

variable "subnet_id" {
  type        = string
  description = "The ID of the subnet where the EC2 will live"
}

variable "instance_type" {
  type    = string
  default = "t3.micro"
  description = "The size of the virtual server to deploy"
}

# ==============================================================
# DEPRECATED VARIABLES (Kept for documentation/portfolio reference)
# ==============================================================

#This will be commented out as SSM comes up
# variable "dev_ip_address" {
#   type        = string
#   description = "Your home IP for the security group (e.g. 1.2.3.4/32)"
# }

# Deprecated due to ingress change from networking to SSM
# variable "key_name" {
#     type = string
#     description = "Which SSH Key Pair to Use for access"
# }