resource "terraform_remote_state" "shared-services" {
  backend = "s3"

  config {
    bucket = "ti-terraform-state"
    key = "{{.AwsAccountId}}/{{.Environment}}/shared-services/terraform.tfstate"
    region = "us-east-1"
  }
}
