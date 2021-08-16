locals {
    app_config = yamldecode(file("app.yml"))
}


remote_state {
  backend = "s3"
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
  config = {
    bucket = local.app_config.tf_state_bucket

    key = "${path_relative_to_include()}/terraform.tfstate"
    region         = "us-west-1"
    encrypt        = true
  }
}

generate "provider" {
  path = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents = <<EOF
provider "aws" {
  region = "us-west-1"
}
EOF
}