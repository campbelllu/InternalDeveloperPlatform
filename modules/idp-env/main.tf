# ==============================================
# Security
# ==============================================
# Egress only necessary part due to SSM setup below
resource "aws_security_group" "sg" {
  name   = "${var.env_name}-sg"
  vpc_id = var.vpc_id

  # Ingress Rules: Deprecated due to SSM below. All ingress ports closed. AWS CLI uses IAM creds to tunnel to IDP.
  # Allow Web Traffic ONLY from the Dev's local IP; HTTP/S and SSH
  # ingress {
  #   description = "HTTP from dev"
  #   from_port   = 80
  #   to_port     = 80
  #   protocol    = "tcp"
  #   cidr_blocks = [var.dev_ip_address] 
  # }
  # ingress {
  #   description = "HTTPS from dev"
  #   from_port   = 443
  #   to_port     = 443
  #   protocol    = "tcp"
  #   cidr_blocks = [var.dev_ip_address] 
  # }
  # ingress {
  #   description = "SSH from dev"
  #   from_port   = 22
  #   to_port     = 22
  #   protocol    = "tcp"
  #   cidr_blocks = [var.dev_ip_address]
  # }

  # Standard Egress (Allow the EC2 to talk to the world)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1" #all protocols
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# ==============================================
# IAM & SSM Setup
# ==============================================

# Create the Role
resource "aws_iam_role" "ssm_role" {
  name = "${var.env_name}-ssm-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

# Attach the SSM Policy to the Role
resource "aws_iam_role_policy_attachment" "ssm_attach" {
  role       = aws_iam_role.ssm_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# Create the Instance Profile
resource "aws_iam_instance_profile" "ssm_profile" {
  name = "${var.env_name}-ssm-profile"
  role = aws_iam_role.ssm_role.name
}

# Chosen to not use this as adding Ansible to this project is a good example of scope creep
# This particular sandbox-making CLI tool doesn't need that level of management
# Terraform and Go will handle the sandboxes being Docker-ready and pushing the image there for testing
# Create the specific policy for EC2 Instance Connect
# resource "aws_iam_policy" "idp_instance_connect_policy" {
#   name        = "idp-instance-connect-policy"
#   description = "Allows the IDP engine to push temporary public keys to sandboxes"

#   policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = [
#       {
#         Effect   = "Allow"
#         Action   = "ec2-instance-connect:SendSSHPublicKey"
#         Resource = "arn:aws:ec2:us-east-2:598708931098:instance/*"
#         # Note: This limits key pushing exclusively to instances in your account!
#         Condition = {
#           StringEquals = {
#             "ec2:osuser" = "ubuntu"
#           }
#         }
#       }
#     ]
#   })
# }
# Attach the policy directly to your terraformer user
# resource "aws_iam_user_policy_attachment" "attach_connect_to_terraformer" {
#   user       = "terraformer" # Matches your AWS ARN username string
#   policy_arn = aws_iam_policy.idp_instance_connect_policy.arn
# }

# ==============================================
# Data Lookups - This could be in data.tf, but it's the sole inhabitor, so to keep file structure clean, it was left here
# ==============================================

# Automatically find the latest Ubuntu 22.04 AMI for instance setup
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical's official AWS ID

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
}

# ==============================================
# Infrastructure
# ==============================================

# The Server, tagged so the dev can find it, and for the Reaper to eliminate it
resource "aws_instance" "dev_node" {
  ami                  = data.aws_ami.ubuntu.id
  instance_type        = var.instance_type
  subnet_id            = var.subnet_id
  iam_instance_profile = aws_iam_instance_profile.ssm_profile.name

  vpc_security_group_ids = [aws_security_group.sg.id]

  # THE CURFEW CRON SCRIPT:
  user_data = <<-EOF
              #!/bin/bash
              # Create a cron job inside the EC2 to shut down automatically every Friday at 21:00 (9 PM)
              echo "0 21 * * 5 root /sbin/shutdown -h now" > /etc/cron.d/weekly-idp-cleanup 
              # fed originally into: >> /etc/crontab; the above would be best practice for more permanent ec2's

              # Update the operating system package registry
              apt-get update -y
              
              # Install Docker and its core dependencies cleanly
              apt-get install -y curl git docker.io
              
              # Ensure the Docker engine starts up and stays on automatically
              systemctl start docker
              systemctl enable docker
              
              # Give the standard ubuntu user permission to use Docker without sudo
              usermod -aG docker ubuntu
              EOF

  tags = { Name = var.env_name
           ManagedBy = "IDP-CLI"
  }
}

########################################################################################
# ==============================================
# Deprecated or Notes
# ==============================================

# The dev instance (EC2)
# resource "aws_instance" "dev_node" {
#   ami           = data.aws_ami.ubuntu.id
#   instance_type = var.instance_type
#   subnet_id     = var.subnet_id # Passed in from the CLI
#   key_name      = var.key_name

#   vpc_security_group_ids = [aws_security_group.sg.id]

#   tags = {
#     Name = var.env_name
#   }
# }