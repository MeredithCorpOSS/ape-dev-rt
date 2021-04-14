module "aws-app" {
  source = "git@github.com:MeredithCorpOSS/ape-dev-tf-aws-app.git"

  aws_access_key = ""
  aws_secret_key = ""

  aws_region         = "${terraform_remote_state.shared-service.output.aws_region}"
  aws_jumpbox_sg_id  = "${terraform_remote_state.shared-service.output.aws_jumpbox_sg_id}"
  aws_vpc_id         = "${terraform_remote_state.shared-service.output.aws_vpc_id}"
  division_id        = "Testing Team Name"
  env_id             = "${terraform_remote_state.shared-service.output.env_name}"
  app_id             = "test-app-name"
  app_stack_id       = ""
  aws_hosted_zone_id = "${terraform_remote_state.shared-service.output.aws_hosted_zone_id}"
  dns_domain         = "test-app-name"
}


resource "aws_s3_bucket" "0_pixel_bucket" {
  bucket = "bucket1-pixel-bucket"
  acl    = "private"
}

resource "aws_s3_bucket" "1_pixel_bucket" {
  bucket = "bucket2-pixel-bucket"
  acl    = "private"
}

resource "aws_s3_bucket" "2_pixel_bucket" {
  bucket = "bucket3-pixel-bucket"
  acl    = "private"
}

