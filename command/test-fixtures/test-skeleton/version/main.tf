variable "app_version" {}

resource "template_file" "cloud_config" {
  filename = "cloud-config.yaml"

  vars {
    etcd_dns = "${terraform_remote_state.shared-services.output.etcd_lb_dns_name}"
    app_name = "${terraform_remote_state.app.output.app_name}"
    app_version = "${var.app_version}"
    environment = "${terraform_remote_state.shared-services.output.env_name}"
    aws_coreos_instance_type = "t2.micro"
    aws_region = "us-east-1"

    team_name = "uk-ops"
    container_name = "example"
    image_url = "ghost:latest"
  }
}

module "app-version" {
  source = "git@github.com:TimeIncOSS/ape-dev-tf-aws-app-version.git?ref=v0.1.1"

  aws_region = "us-east-1"
  aws_availability_zones = "us-east-1b,us-east-1c,us-east-1d,us-east-1e"
  aws_keypair_name = "coreos-test"

  app_elb_name = "${terraform_remote_state.app.app_elb_name}"
  aws_app_elb_sg_id = "${terraform_remote_state.app.output.app_elb_sg_id}"
  aws_compute_sg_id = "${terraform_remote_state.shared-services.output.aws_application_sg_id}"
  aws_etcd_sg_id = "${terraform_remote_state.shared-services.output.aws_etcd_sg_id}"
  aws_app_sg_id = "${terraform_remote_state.app.output.aws_app_sg_id}"
  aws_asg_subnet_ids = "${terraform_remote_state.shared-services.output.aws_private_subnet_ids}"

  team_name = "uk-ops"
  environment = "${terraform_remote_state.shared-services.output.env_name}"
  app_name = "${terraform_remote_state.app.output.app_name}"
  app_version = "${var.app_version}"

  dns_domain = "example"
  etcd_dns = "${terraform_remote_state.shared-services.output.etcd_lb_dns_name}"

  app_version_user_data = "${template_file.cloud_config.rendered}"
}

output "app_elb_dns" {
  value = "${module.app-version.app_dns_name}"
}
