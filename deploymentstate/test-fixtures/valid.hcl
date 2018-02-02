deployment_state "fixture" {
  bucket = "ti-rt-deployment-state-{{.Environment}}"
  key = "{{.AwsAccountId}}/hubot/nothing/"
}