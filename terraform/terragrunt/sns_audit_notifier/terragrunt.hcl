include {
  path = find_in_parent_folders()
}

locals {
    app_config = yamldecode(file("${find_in_parent_folders("app.yml")}"))
}

terraform {
  source  = "${get_parent_terragrunt_dir()}/../modules/sns_audit_notifier"
}

inputs = {
    email = local.app_config.email
}