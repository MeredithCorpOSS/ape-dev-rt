resource "terraform_remote_state" "app" {
  backend = "s3"

  config {
    bucket = "ti-terraform-state"
    key = "{{.AwsAccountId}}/{{.Environment}}/{{.AppName}}/terraform.tfstate"
    region = "us-east-1"
  }
}
