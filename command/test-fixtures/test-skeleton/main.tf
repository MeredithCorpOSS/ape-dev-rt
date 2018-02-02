module "aws-app" {
  source = "git@github.com:TimeInc/ape-dev-tf-aws-app.git?ref=v0.1.2"

  # TODO: Needs debugging to find out why aws_region can't be passed as var
  aws_region = "us-east-1"
  aws_hosted_zone_id = "Z3L9P9L4Y49RDU"
  aws_jumpbox_sg_id = "${terraform_remote_state.shared-services.output.aws_jumpbox_sg_id}"
  aws_elb_subnet_ids = "${terraform_remote_state.shared-services.output.aws_public_subnet_ids}"
  aws_vpc_id = "${terraform_remote_state.shared-services.output.aws_vpc_id}"
  environment = "${terraform_remote_state.shared-services.output.env_name}"
  app_iam_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "DenyAll",
            "Effect": "Deny",
            "Action": "*",
            "Resource": "*"
        }
    ]
}
POLICY

  client_cidr = "0.0.0.0/0"

  team_name = "uk-ops"
  app_name = "example"
}
