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

#Need to add dynamo privs to this user if you want terraform to make this
# Create a DynamoDB table for state locking
resource "aws_dynamodb_table" "terraform_locks" {
  name         = "terraform-up-and-running-locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}


#Need to add master VPC below this, home for all the IDP-subnets