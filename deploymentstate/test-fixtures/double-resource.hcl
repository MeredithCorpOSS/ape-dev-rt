deployment_state "fixture" {
  bucket = "aaa"
  key = "oink"
}

deployment_state "fixture" {
  bucket = "bbb"
  key = "pooh"
}

deployment_state "oink" {
  bucket = "ti-rt-deployment-state-{{.Environment}}"
  key = "{{.AwsAccountId}}/hubot/yadada/"
}
