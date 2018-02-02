# RT vs Elastic Beanstalk

## Elastic Beanstalk

Beanstalk allows user to deploy applications into predefined stacks, most of which are EC2 instances
running Amazon Linux (+ eventually Docker). It leverages CloudFormation for the management of underlying
resources (S3 bucket for env data & artefacts, SG, ELB, ASG, CloudWatch Alarms, Autoscaling Policies, ...).

It is not possible to specify custom CloudFormation code. All CloudFormation changes and templates are managed
from within Beanstalk.

Some resources can be managed outside of Beanstalk and IDs provided via optional settings

 - VPC + subnets
 - IAM Role/Instance Profile
 - AMI

Some resources *can* be managed by (limited) Beanstalk settings:

 - 1 SNS topic
 - 1 SQS queue
 - Cloudwatch Alarms related to ELB, ASG
 - 1 ELB
 - 1 ASG + 1 Scaling Policy based on 1 metric
 - 1 Launch Configuration
 - 1 EBS Volume
 - 1 RDS instance

CloudFormation often falls behind in terms of supporting new AWS products and services
which may be limiting factor (e.g. if there's new ELB or ASG feature application wants to use).

Since the Beanstalk model is based on ASG per application it can be expensive for applications
which may just want to share instances.

Beanstalk currently makes it very easy to spin up environment for common apps by using sensible defaults
for most settings. These settings however may lack good security, scalability, availability patterns/practices.

Beanstalk provides **default scaling policy** which makes it attractive to some people as they don't need to define this explicitely.
By default it uses `NetworkOut > 6MB` per 5 mins as a metric for scaling up.

RT UX is currently behind Beanstalk, but will improve over time to make day-to-day deployments
as easy and boring as they can be.

Beanstalk capabilities (in terms of artefacts it can deploy) will most likely be always limited to things
that can run on EC2 instances (i.e. no Lambda Functions or API Gateway resources).

Beanstalk does not support safe rollback - it has the notion of "rollback" term, but
the rollback is understood as ["roll forward"](http://support.beanstalkapp.com/article/843-how-can-i-rollback-a-deployment-to-a-previous-revision) - effectively redeploying older version.

Running containers can also get out of hand (scalability, monitoring) without schedulers like ECS.
It is possible to run multiple Docker containers via Beanstalk, but it is unclear how does Beanstalk
ensure that these containers keep running and how easy it is to run multiple instances of the same containers.

## RT

RT aims to provide templates for managing any dependencies related to the app also e.g.

 - S3 buckets
 - KMS keys
 - API Gateway
 - Lambda
 - CloudWatch Logs
 - CloudWatch Events
 - DynamoDB Tables
 - ElastiCache clusters
 - ElasticSearch
 - Glacier
 - RDS, Aurora with any settings
 - any number of CloudWatch metrics & alerts
 - any number of SNS topics with any settings
 - any number of SQS queues with any settings

If user still wants to use Beanstalk, they can do so via following Terraform's resources in RT too:

 - `aws_elastic_beanstalk_environment`
 - `aws_elastic_beanstalk_application`
 - [`aws_elastic_beanstalk_application_version`](https://github.com/hashicorp/terraform/pull/5770)
 - `aws_elastic_beanstalk_configuration_template`
