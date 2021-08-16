output alb_zone_id {
    value = aws_lb.lambda_alb.zone_id
}

output alb_dns_name {
    value = aws_lb.lambda_alb.dns_name
}