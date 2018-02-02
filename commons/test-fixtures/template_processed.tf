module "aws-app" {
  source = "git@github.com:TimeInc/ape-dev-tf-aws-app.git"

  aws_access_key = ""
  aws_secret_key = ""

  aws_region = "${terraform_remote_state.shared-service.output.aws_region}"
  aws_jumpbox_sg_id = "${terraform_remote_state.shared-service.output.aws_jumpbox_sg_id}"
  aws_vpc_id = "${terraform_remote_state.shared-service.output.aws_vpc_id}"
  division_id = "Testing Team Name"
  env_id = "${terraform_remote_state.shared-service.output.env_name}"
  app_id = "test-app-name"
  app_stack_id = ""
  aws_hosted_zone_id = "${terraform_remote_state.shared-service.output.aws_hosted_zone_id}"
  dns_domain = "test-app-name"
}
