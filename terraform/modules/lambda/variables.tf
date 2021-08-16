variable bin_path {
    type = string
}

variable bin_name {
    type = string
}

variable app_name {
    type = string
}

variable lambda_subnets {
    type = list(string)
}

variable alb_subnets {
    type = list(string) 
}

variable vpc_id {
    type = string
}

variable cert_arn {
    type = string
}

variable auth_table_name {
    type = string
}