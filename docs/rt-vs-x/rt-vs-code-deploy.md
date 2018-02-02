# RT vs Code Deploy

Code Deploy, similarly to Elastic Beanstalk is aimed to solve deployments that can happen
on EC2 instances.

Code Deploy aims to solve the transfer of artefacts from artefact storage to instances,
but it does not aim to deal with scalability, security, load balancing, monitoring etc.

Code Deploy API does not allow removal/uninstallation of an application that was previously installed on EC2 instance.
This prescribes a dedicated set of instances (likely an ASG) to be used only for a single application
and be terminated at a time of decomissioning.

Code Deploy does not allow easy [rollback](http://docs.aws.amazon.com/codedeploy/latest/userguide/how-to-rollback-redeploy.html)
(it's would be effectively roll forward - i.e. redeploying older version).

Code Deploy may not be the best tool for deployments if artefact becomes Docker container/image
or Lambda Function.

## RT

If user wants to perform deployments via Code Deploy, that _might_ be able to do so via following Terraform resources:

 - [`aws_codedeploy_app`](https://www.terraform.io/docs/providers/aws/r/codedeploy_app.html)
 - [`aws_codedeploy_deployment_group`](https://www.terraform.io/docs/providers/aws/r/codedeploy_deployment_group.html)
 - `aws_codedeploy_deployment` - TBD, not supported yet, may also lack `D` from complete `CRUD` interface due to inability of uninstallation
