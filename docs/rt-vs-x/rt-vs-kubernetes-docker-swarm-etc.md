# RT vs Kubernetes, Docker Swarm, ECS etc.

RT is not trying to solve the same problem as container schedulers like K8S, Swarm or ECS are.

RT aims to allow deployment of non-containerised apps and services, e.g.

 - Lambda Functions
 - API Gateway
 - databases (RDS, Aurora)
 - S3

Not everything can be easily containerised :whale: today (off-the-shelf databases are not trivial).

Since RT leverages Terraform it is possible to deploy containers into K8S, Swarm or ECS cluster:

 - `aws_ecs_cluster`
 - `aws_ecs_service`
 - `aws_ecs_task_definition`
 - [K8S provider](https://github.com/hashicorp/terraform/pull/3453)
 - [Docker provider](https://www.terraform.io/docs/providers/docker/index.html)
