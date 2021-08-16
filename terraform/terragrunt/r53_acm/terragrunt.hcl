include {
  path = find_in_parent_folders()
}

locals {
    app_config = yamldecode(file("${find_in_parent_folders("app.yml")}"))
}

terraform {
  source  = "${get_parent_terragrunt_dir()}/../modules/r53_acm"
}

inputs = {
    zone = local.app_config.zone
    domain = local.app_config.domain
}