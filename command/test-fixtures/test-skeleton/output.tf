output "app_name" {
  value = "${module.aws-app.app_name}"
}

output "app_elb_sg_id" {
  value = "${module.aws-app.app_elb_sg_id}"
}

output "aws_app_sg_id" {
  value = "${module.aws-app.app_sg_id}"
}
