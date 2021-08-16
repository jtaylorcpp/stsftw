resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name           = var.dynamodb_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "issuer"
  range_key      = "account_name"

  attribute {
    name = "issuer"
    type = "S"
  }

  attribute {
    name = "account_name"
    type = "S"
  }
}