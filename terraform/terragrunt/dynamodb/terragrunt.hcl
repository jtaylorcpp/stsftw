include {
  path = find_in_parent_folders()
}

locals {
    app_config = yamldecode(file("${find_in_parent_folders("app.yml")}"))
}

terraform {
  source  = "${get_parent_terragrunt_dir()}/../modules/dynamodb"
}

inputs = {
    dynamodb_name = local.app_config.dynamo_table_name
}