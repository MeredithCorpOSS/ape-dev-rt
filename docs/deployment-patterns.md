# Deployment Patterns

We aim to support the [blue/green deployments](http://martinfowler.com/bliki/BlueGreenDeployment.html) to reduce downtime and risk during deployments.

RT currently supports _"colorful" deployment_ only. We plan to support both in future versions
and will expect the user to choose deployment pattern at their own discretion.

## A. "Colorful" deployment :blue_book: :closed_book: :orange_book: :green_book:

**2+ versions** ("colors") can be deployed and run (be available for the customer) at the same time in each environment.

 - **Pros**
   - Allows cautiuous, slow rollout
   - Allows A/B testing of new release (performance review)
 - **Cons**
   - It may be slower to `apply`/`destroy` a version
   - Deployments cannot be fully automated as it's not clear which version(s) should be deployed & stay deployed

## B. Regular "Blue or Green" deployment :blue_book: :green_book:

Only **1 version** ("color") may exist at a time in each environment.

 - **Pros**
   - It may be quicker to `apply`/`destroy` a version
   - Deployments can be easily automated because it is clear that we always want 1 specific version
 - **Cons**
   - Rollback may be a bit more difficult if we want to do it safely

----------

## Notes

 - In the blue/green deployment, the challenge is how to do the flip/over quickly and easily - this may require implementation of some new RT commands (`promote-version`?) -> how would such command work? - i.e. what could be the flip-over? Most mechanisms won't guarantee atomicity due to the nature of distributed systems.
   - DNS record change?
   - ELB/ASG association with Cookie Stickiness enabled?
   - ECS TD / ECS service association?
   - K8S?
 - how to communicate such event from the `promote-version` event down to Terraform? 0/1 variable that can be used in `count` parameter?
