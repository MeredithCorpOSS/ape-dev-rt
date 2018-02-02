package aws

import (
	"errors"
	"fmt"
	"log"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
)

type Balancer struct {
	Name  string
	State string
}

type InstanceHealth struct {
	InstanceID string
	State      string
}

func (a *AWS) GetBalancersFromScalingGroup(scalingGroup string) ([]*Balancer, error) {
	log.Printf("[DEBUG] Discovering balancers for ASG %q", scalingGroup)
	svc := a.autoscalingConn
	resp, err := svc.DescribeLoadBalancers(&autoscaling.DescribeLoadBalancersInput{
		AutoScalingGroupName: awsSDK.String(scalingGroup),
	})
	if err != nil {
		return nil, err
	}

	var balancers []*Balancer
	for _, b := range resp.LoadBalancers {
		balancers = append(balancers, &Balancer{
			*b.LoadBalancerName,
			*b.State,
		})
	}

	return balancers, nil
}

func (a *AWS) GetInstanceIdsFromScalingGroup(scalingGroup string) ([]*string, error) {
	svc := a.autoscalingConn
	var instanceIds []*string
	resp, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{awsSDK.String(scalingGroup)},
	})
	if err != nil {
		return instanceIds, err
	}

	instances := resp.AutoScalingGroups[0].Instances
	if len(instances) == 0 {
		return nil, fmt.Errorf("No instances found in ASG %q", scalingGroup)
	}

	for _, i := range instances {
		instanceIds = append(instanceIds, i.InstanceId)
	}
	return instanceIds, nil
}

func (a *AWS) GetAppNameFromScalingGroup(scalingGroup string) (string, error) {
	svc := a.autoscalingConn
	resp, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{awsSDK.String(scalingGroup)},
	})
	if err != nil || len(resp.AutoScalingGroups) == 0 {
		return "", fmt.Errorf("Failed getting app name from ASG %s which does not exist", scalingGroup)
	}
	for _, tag := range resp.AutoScalingGroups[0].Tags {
		if *tag.Key == "App" {
			return *tag.Value, nil
		}
	}
	return "", fmt.Errorf("Failed getting app name from ASG %s which is missing \"App\" tag", scalingGroup)
}

func (a *AWS) GetScalingGroupForSlotId(environment, appName, slotId string) (string, error) {
	log.Printf("[DEBUG] Discovering autoscaling group for %q slot %q", appName, slotId)
	svc := a.autoscalingConn
	resp, err := svc.DescribeTags(&autoscaling.DescribeTagsInput{
		Filters: []*autoscaling.Filter{
			{
				Name: awsSDK.String("key"),
				Values: []*string{
					awsSDK.String("Name"),
				},
			},
			{
				Name: awsSDK.String("value"),
				Values: []*string{
					awsSDK.String(fmt.Sprintf("%s-%s-v%s-vinst", environment, appName, slotId)),
				},
			},
		},
	})
	if err != nil {
		return "", err
	}
	tagCount := len(resp.Tags)
	var scalingGroupId string
	if tagCount > 1 {
		return "", errors.New("More than one ASG for slot " + slotId)
	} else if tagCount == 0 {
		return "", nil
	}
	scalingGroupId = *resp.Tags[0].ResourceId
	return scalingGroupId, err
}

func (a *AWS) DescribeBalancedInstanceHealth(balancerName string) ([]*InstanceHealth, error) {
	svc := a.elbConn
	var instances []*InstanceHealth
	resp, err := svc.DescribeInstanceHealth(&elb.DescribeInstanceHealthInput{
		LoadBalancerName: &balancerName,
	})
	if err != nil {
		return instances, err
	}
	for _, health := range resp.InstanceStates {
		instances = append(instances, &InstanceHealth{
			*health.InstanceId,
			*health.State,
		})
	}
	return instances, nil
}

func (a *AWS) GetPrivateIpsForInstanceIds(instanceIds []*string) (map[string]string, error) {
	svc := a.ec2Conn
	data, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return nil, err
	}

	var result = make(map[string]string, len(instanceIds))
	for _, r := range data.Reservations {
		instances := r.Instances
		if l := len(instances); l != 1 {
			return nil, fmt.Errorf("Expected exactly 1 instance, %d given", l)
		}
		instance := instances[0]
		result[*instance.InstanceId] = *instance.PrivateIpAddress
	}

	return result, nil
}

func (a *AWS) GetBalancersForApp(appName string) ([]string, error) {
	log.Printf("[DEBUG] Discovering load balancers for  %q", appName)
	var balancers []string
	svc := a.elbConn
	describeAllResponse, descError := svc.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{},
	})
	if descError != nil {
		return balancers, descError
	}
	var allBalancerNames []*string
	for _, desc := range describeAllResponse.LoadBalancerDescriptions {
		allBalancerNames = append(allBalancerNames, desc.LoadBalancerName)
	}

	var tagDescriptions []*elb.TagDescription
	for len(allBalancerNames) > 0 {
		log.Printf("[DEBUG] checking tags for some load balancers: %q", allBalancerNames)
		var batchOfNames []*string
		if len(allBalancerNames) > 20 {
			batchOfNames = allBalancerNames[:20]
		} else {
			batchOfNames = allBalancerNames
		}
		tagsResponse, tagsError := svc.DescribeTags(&elb.DescribeTagsInput{
			LoadBalancerNames: batchOfNames,
		})
		if tagsError != nil {
			return balancers, tagsError
		}
		for _, desc := range tagsResponse.TagDescriptions {
			tagDescriptions = append(tagDescriptions, desc)
		}
		allBalancerNames = allBalancerNames[len(batchOfNames):]
	}
	for _, tagsDescription := range tagDescriptions {
		for _, tag := range tagsDescription.Tags {
			if *tag.Key == "App" && (*tag.Value == appName) {
				balancers = append(balancers, *tagsDescription.LoadBalancerName)
			}
		}
	}
	log.Printf("[DEBUG] found load balancers for %q: %q", appName, balancers)
	return balancers, nil
}

func (a *AWS) DetachBalancersFromScalingGroup(balancerNames []string, groupName string) error {
	svc := a.autoscalingConn
	var awsBalancerNames []*string
	for _, balancerName := range balancerNames {
		awsBalancerNames = append(awsBalancerNames, awsSDK.String(balancerName))
	}
	_, err := svc.DetachLoadBalancers(&autoscaling.DetachLoadBalancersInput{
		AutoScalingGroupName: awsSDK.String(groupName),
		LoadBalancerNames:    awsBalancerNames,
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *AWS) AttachBalancersToScalingGroup(balancerNames []string, groupName string) error {
	svc := a.autoscalingConn
	var awsBalancerNames []*string
	for _, balancerName := range balancerNames {
		awsBalancerNames = append(awsBalancerNames, awsSDK.String(balancerName))
	}
	_, err := svc.AttachLoadBalancers(&autoscaling.AttachLoadBalancersInput{
		AutoScalingGroupName: awsSDK.String(groupName),
		LoadBalancerNames:    awsBalancerNames,
	})
	if err != nil {
		return err
	}
	return nil
}
