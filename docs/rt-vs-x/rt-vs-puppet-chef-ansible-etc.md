# RT vs Puppet, Chef, Ansible, Salt Stack etc.

Most configuration management solutions like Puppet or Chef do allow (RPM, DEB, ...) package installations,
however the deployment artefact may not always be in the form of system packages.

It may not be trivial to schedule containers reliably and scalably into a cluster of EC2 instances.
Most of the cfg mgmt tools don't even aim to solve that problem.

There can be benefits in using these tools to install & configure services inside Docker image
before it's shipped off to chosen registery. `Dockerfile`s seem to be the preffered approach
in the early days of Docker adoption which can make dependency management and generally
maintainence of the configuration difficult.

There is Puppet Cloud provisioner is similar to Terraform (in a way that they can both deploy AWS resources)
but it uses tags-based discovery which may impose risk of tampering with wrong set of resources.

------

_TODO: This needs more care from someone who knows more about deployments via config mgmt tools_
