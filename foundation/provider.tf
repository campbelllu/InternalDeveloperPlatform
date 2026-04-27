terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "s3" {
    bucket         = "idp-tf-state-storage"
    key            = "foundation/terraform.tfstate"
    region         = "us-east-2"
    dynamodb_table = "terraform-running-locks"
    encrypt        = true
  }
}

provider "aws" {
  region = "us-east-2"
}