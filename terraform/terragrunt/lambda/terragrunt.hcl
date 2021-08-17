include {
  path = find_in_parent_folders()
}

locals {
    app_config = yamldecode(file("${find_in_parent_folders("app.yml")}"))
}

dependency vpc {
    config_path = "${get_parent_terragrunt_dir()}/vpc"
}

dependency cert {
    config_path = "${get_parent_terragrunt_dir()}/r53_acm"
}

dependency dynamo {
  config_path = "${get_parent_terragrunt_dir()}/dynamodb"
}

dependency sns {
  config_path = "${get_parent_terragrunt_dir()}/sns_audit_notifier"
}

terraform {
  source  = "${get_parent_terragrunt_dir()}/../modules/lambda"
}

inputs = {
    bin_path = local.app_config.bin_path
    bin_name = local.app_config.bin_name
    app_name = local.app_config.app_name
    lambda_subnets = dependency.vpc.outputs.private_subnets
    alb_subnets = dependency.vpc.outputs.public_subnets
    vpc_id = dependency.vpc.outputs.vpc_id
    cert_arn = dependency.cert.outputs.arn
    auth_table_name = dependency.dynamo.outputs.table_name
    sns_arn = dependency.sns.outputs.arn
}