deployment_state "fixture" {
  key = "{{.AwsAccountId}}/{{.Environment}}/tada/"
  bucket = "ti-rt-deployment-state"
}

random_thing_oink "fixture" {
}