include {
  path = find_in_parent_folders()
}

locals {
    app_config = yamldecode(file("${find_in_parent_folders("app.yml")}"))
}

terraform {
  source  = "${get_parent_terragrunt_dir()}/../modules/r53"
}

dependency lambda {
    config_path = "${get_parent_terragrunt_dir()}/lambda"
}

inputs = {
    zone = local.app_config.zone
    domain = local.app_config.domain
    alb_zone_id = dependency.lambda.outputs.alb_zone_id
    alb_dns_name = dependency.lambda.outputs.alb_dns_name
}