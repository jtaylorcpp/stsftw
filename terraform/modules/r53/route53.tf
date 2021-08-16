data aws_route53_zone app_zone {
  name         = var.zone
  private_zone = false
}

resource "aws_route53_record" "app" {
  zone_id = data.aws_route53_zone.app_zone.zone_id
  name    = var.domain
  type    = "A"

  alias {
    name                   = var.alb_dns_name
    zone_id                = var.alb_zone_id
    evaluate_target_health = true
  }
}