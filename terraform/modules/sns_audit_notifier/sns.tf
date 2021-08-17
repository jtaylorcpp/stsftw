resource "aws_sns_topic" "audit_channel" {
  name                        = "sts_audit_channel"
}

resource aws_sns_topic_subscription audit_email {
    endpoint = var.email
    protocol = "email"
    topic_arn = aws_sns_topic.audit_channel.arn
}