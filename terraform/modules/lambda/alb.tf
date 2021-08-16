resource aws_security_group inbound {
  name        = "${var.app_name}_alb_allow_inbound"
  vpc_id      = var.vpc_id

  
    ingress {
      from_port        = 443
      to_port          = 443
      protocol         = "tcp"
      cidr_blocks      = ["0.0.0.0/0"]
    }

    ingress {
      from_port        = 80
      to_port          = 80
      protocol         = "tcp"
      cidr_blocks      = ["0.0.0.0/0"]
    }

    egress {
      from_port        = 0
      to_port          = 0
      protocol         = "-1"
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = ["::/0"]
    }
}

resource "aws_lb" "lambda_alb" {
  name               = "${var.app_name}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.inbound.id]
  subnets            = var.alb_subnets
}

resource "aws_lb_listener" http {
  load_balancer_arn = aws_lb.lambda_alb.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_listener" https {
  load_balancer_arn = aws_lb.lambda_alb.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = var.cert_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.lambda.arn
  }
}

resource "aws_lb_target_group" "lambda" {
  name        = "${var.app_name}-alb-tg"
  target_type = "lambda"
}

resource "aws_lambda_permission" "with_lb" {
  statement_id  = "AllowExecutionFromlb"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.sts_app.arn
  principal     = "elasticloadbalancing.amazonaws.com"
  source_arn    = aws_lb_target_group.lambda.arn
}

resource "aws_lb_target_group_attachment" sts_app_lb {
  target_group_arn = aws_lb_target_group.lambda.arn
  target_id        = aws_lambda_function.sts_app.arn
  depends_on       = [aws_lambda_permission.with_lb]
}