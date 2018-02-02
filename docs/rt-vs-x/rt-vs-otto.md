# RT vs Otto

Otto is a new tool - 1st release announced at HashiConf in September 2015
(1st release of RT was in May 2015).

RT shares some concepts with Otto, e.g.

 - uses Terraform for describing both infrastructure and application
 - respects 2 distinct TF contexts (`otto infra` & `otto deploy` VS `rt apply-infra` & `rt deploy`)

Otto aims to solve much wider range of problems (as opposed to RT):

 - local development via Vagrant/Docker (`otto dev`)
 - packaging via Packer

RT needs (or will need) to cater a few specific needs that are not currently covered by Otto:

 - deployment of multiple versions (A/B testing, safe rollback)
 - [Terraform "workarounds" via Go Templates](https://github.com/TimeInc/ape-dev-rt/blob/master/docs/internals.md#go-templates)
 - custom functionality
   - `*-traffic` commands
   - (`wait-for-*` commands)
   - [temp branch deployment](https://trello.com/c/wxOBPTHp/188-rt-allow-exposing-any-version-temporarily-w-out-having-to-deploy-it)
   - [_deployment groups_](https://trello.com/c/D0V5vln3/179-rt-support-application-groups-e-g-keystone-platform)

There are things Otto is unlikely to support since Hashicorp has such functionality in Atlas (commercial offering), e.g.

 - [notifications](https://trello.com/c/e1s9hg24/189-rt-allow-adding-removing-hooks)
 - tfstate locking

Atlas could provide a nice UI for deployment, but it currently doesn't have any nice support for Otto. It does not have any notion of application/version nor common authentication mechanisms like SAML or OAuth. This may change over time. The current priorities as of Apr 2016 are rather to keep improving support for existing tools like Terraform, Consul, Vault etc. (as discussed w/ Hashicorp).

Otto shares the base principle with RT (Terraform configs in two directories) which will allow us to revisit
Otto as an alternative to RT in the future in regards to missing features mentioned above.

Similar to RT, Otto will always support basically unlimited number of infrastructure resources.
The only limitation is provider/resource support in Terraform (which is open-source and plugin-based).

Both RT and Otto should allow easy deployment of

 - Lambda Functions
 - containers via Docker Swarm, ECS or K8S
 - EC2 instances with custom AMIs via ASGs

via (intentionally) restricted DSL - [HCL](https://github.com/hashicorp/hcl).
