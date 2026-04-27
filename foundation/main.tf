#S3 bucket to store tf state
resource "aws_s3_bucket" "terraform_state" {
  bucket = "idp-tf-state-storage"
 
  # Prevent accidental deletion of this bucket
  lifecycle {
    prevent_destroy = true
  }
}

# Enable versioning so you can see the history of your state files
resource "aws_s3_bucket_versioning" "enabled" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Create a DynamoDB table for state locking
resource "aws_dynamodb_table" "terraform_locks" {
  name         = "terraform-running-locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}

# VPC is the blank stadium upon which all IDP's shall be painted
resource "aws_vpc" "stadium" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "vpc-0"
  }
}

# IGW
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.stadium.id

  tags = { Name = "igw-0" }
}

# Route Table
resource "aws_route_table" "public_rt" {
  vpc_id = aws_vpc.stadium.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name = "public-rt-0"
  }
}

# public subnet
resource "aws_subnet" "public_subnet" {
  vpc_id                  = aws_vpc.stadium.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name = "public-subnet-0"
  }
}

# RT Associated with public_subnet
resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public_subnet.id
  route_table_id = aws_route_table.public_rt.id
}


# For reference or debugging later; run terraform apply to print to screen; also saves ID's to s3 for later programmatical calls to their ID's
output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.stadium.id
}

output "public_subnet_id" {
  description = "The ID of the original subnet"
  value       = aws_subnet.public_subnet.id
}
