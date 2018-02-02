# Remote State

If you have just upgraded to RT 0.6.x from previous version, you'll probably see this kind of message:

```
[ERROR] No 'remote_state' block found in ".../infra/deployment-state.hcl.tpl". See ... for more details.
```

Don't panic, this is fine!

If you don't intend to save the state anywhere else and just carry on using the one that was implicitly
used before, just copy and paste the following block of code into your `deployment-state.hcl.tpl`:

```hcl
remote_state "main" {
  backend = "s3"
  config {
    bucket = "ti-terraform-state-{{.Environment}}"
    region = "us-east-1"
  }
}
```
