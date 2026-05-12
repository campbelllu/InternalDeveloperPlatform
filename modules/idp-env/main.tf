# Define the security group for the instance
# SSM comes online and removes all of the ingress blocks
resource "aws_security_group" "sg" {
  name   = "${var.env_name}-sg"
  vpc_id = var.vpc_id

  # Rule 1: Allow Web Traffic ONLY from the Dev's local IP; HTTP/S and SSH
  ingress {
    description = "HTTP from dev"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = [var.dev_ip_address] 
  }
  ingress {
    description = "HTTPS from dev"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [var.dev_ip_address] 
  }
  ingress {
    description = "SSH from dev"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.dev_ip_address]
  }

  # Rule 2: Standard Egress (Allow the EC2 to talk to the world)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1" #all protocols
    cidr_blocks = ["0.0.0.0/0"]
  }
}
### Notes going forward: Setting up a corporate VPN would simplify this quite a bit: 
### Any instances made allow ingress from the VPN IP, the dev simply connects to that VPN to access their IDP
### I am veto'ing this idea for now, go make a VPN on another project. Focus on other skills for this CLI.
### ALT: SSM/Identity mgmt. All ingress ports closed. AWS CLI uses IAM creds to tunnel to IDP.

### Generated code to start this SSM integration, needs to be integrated into this page before any of this works, but I'm sleeping now

# 1. Create the Role (The 'Job Description')
# resource "aws_iam_role" "ssm_role" {
#   name = "${var.env_name}-ssm-role"

#   assume_role_policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = [{
#       Action = "sts:AssumeRole"
#       Effect = "Allow"
#       Principal = { Service = "://amazonaws.com" }
#     }]
#   })
# }

# # 2. Attach the SSM Policy to the Role
# resource "aws_iam_role_policy_attachment" "ssm_attach" {
#   role       = aws_iam_role.ssm_role.name
#   policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
# }

# # 3. Create the Instance Profile (The 'Physical Badge')
# resource "aws_iam_instance_profile" "ssm_profile" {
#   name = "${var.env_name}-ssm-profile"
#   role = aws_iam_role.ssm_role.name
# }

# # 4. The Server (Now wearing the badge)
# resource "aws_instance" "dev_node" {
#   ami                  = data.aws_ami.ubuntu.id
#   instance_type        = var.instance_type
#   subnet_id            = var.subnet_id
#   iam_instance_profile = aws_iam_instance_profile.ssm_profile.name # <--- The Magic Link

#   # Security Group now only needs EGRESS (to talk to SSM)
#   # Ingress can be empty!
#   vpc_security_group_ids = [aws_security_group.sg.id]

#   tags = { Name = var.env_name }
# }




# Automatically find the latest Ubuntu 22.04 AMI for instance setup
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical's official AWS ID

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
}


# The dev instance (EC2)
resource "aws_instance" "dev_node" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type
  subnet_id     = var.subnet_id # Passed in from the CLI
  key_name      = var.key_name

  vpc_security_group_ids = [aws_security_group.sg.id]

  tags = {
    Name = var.env_name
  }
}
