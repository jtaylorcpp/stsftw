include {
  path = find_in_parent_folders()
}

terraform {
  source  = "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git"
}

inputs = {
    name = "sts-app-vpc"
    cidr = "10.0.0.0/16"

    azs             = ["us-west-1a", "us-west-1c"]
    private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
    public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]

    enable_nat_gateway = false
    //single_nat_gateway  = true
    //one_nat_gateway_per_az = false
}
