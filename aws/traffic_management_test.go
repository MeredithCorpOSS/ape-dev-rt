package aws

import (
	"reflect"
	"testing"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
)

func TestGetBalancersFromScalingGroup(t *testing.T) {
	routes := []*MockRoute{
		&MockRoute{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeLoadBalancers&AutoScalingGroupName=scaling-group-name&Version=2011-01-01",
			Response: MockResponse{
				Code: 200,
				Body: test_autoscaling_balancers_apiBody,
			},
		},
	}
	mockedSession, closeFunc := GetMockedAwsSession(routes, "us-east-1")
	defer closeFunc()

	a := &AWS{
		Region:          awsSDK.String("us-east-1"),
		autoscalingConn: autoscaling.New(mockedSession),
	}

	balancers, err := a.GetBalancersFromScalingGroup("scaling-group-name")
	if err != nil {
		t.Fatalf("Failed getting balancers: %s", err)
	}
	expectedBalancers := []*Balancer{
		&Balancer{Name: "internal-loadbalancer", State: "Added"},
		&Balancer{Name: "external-loadbalancer", State: "Added"},
	}
	if !reflect.DeepEqual(balancers, expectedBalancers) {
		t.Fatalf("Load balancers don't match:\nGiven: %q\nExpected: %q", balancers, expectedBalancers)
	}
}

func TestGetInstanceIdsFromScalingGroup(t *testing.T) {
	routes := []*MockRoute{
		&MockRoute{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeAutoScalingGroups&AutoScalingGroupNames.member.1=my-asg&Version=2011-01-01",
			Response: MockResponse{
				Code: 200,
				Body: test_autoscaling_describeGroups_apiBody,
			},
		},
	}
	mockedSession, closeFunc := GetMockedAwsSession(routes, "us-east-1")
	defer closeFunc()

	a := &AWS{
		Region:          awsSDK.String("us-east-1"),
		autoscalingConn: autoscaling.New(mockedSession),
	}

	ids, err := a.GetInstanceIdsFromScalingGroup("my-asg")
	if err != nil {
		t.Fatalf("Failed getting instance IDs: %s", err)
	}

	expectedIds := []string{"i-12345678", "i-01111108"}
	if !reflect.DeepEqual(pointerSliceToStringSlice(ids), expectedIds) {
		t.Fatalf("Wrong instance IDs received.\nGiven: %q\nExpected: %q\n",
			pointerSliceToStringSlice(ids), expectedIds)
	}
}

func pointerSliceToStringSlice(in []*string) []string {
	out := make([]string, len(in))
	for k, v := range in {
		out[k] = *v
	}
	return out
}

func TestGetAppNameFromScalingGroup(t *testing.T) {
	routes := []*MockRoute{
		&MockRoute{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeAutoScalingGroups&AutoScalingGroupNames.member.1=my-asg&Version=2011-01-01",
			Response: MockResponse{
				Code: 200,
				Body: test_autoscaling_describeGroups_apiBody,
			},
		},
	}
	mockedSession, closeFunc := GetMockedAwsSession(routes, "us-east-1")
	defer closeFunc()

	a := &AWS{
		Region:          awsSDK.String("us-east-1"),
		autoscalingConn: autoscaling.New(mockedSession),
	}

	name, err := a.GetAppNameFromScalingGroup("my-asg")
	if err != nil {
		t.Fatalf("Failed getting app name: %s", err)
	}

	expectedName := "elmo"
	if name != expectedName {
		t.Fatalf("Wrong app name received.\nGiven: %q\nExpected: %q\n",
			name, expectedName)
	}
}

func TestGetScalingGroupForSlotId(t *testing.T) {
	routes := []*MockRoute{
		&MockRoute{
			ExpectedURI: "/",
			ExpectedRequestBody: "Action=DescribeTags&Filters.member.1.Name=key&" +
				"Filters.member.1.Values.member.1=Name&Filters.member.2.Name=value&" +
				"Filters.member.2.Values.member.1=prod-cookie_monster-vfirst-vinst&" +
				"Version=2011-01-01",
			Response: MockResponse{
				Code: 200,
				Body: test_autoscaling_describeTags_apiBody,
			},
		},
	}
	mockedSession, closeFunc := GetMockedAwsSession(routes, "us-east-1")
	defer closeFunc()

	a := &AWS{
		Region:          awsSDK.String("us-east-1"),
		autoscalingConn: autoscaling.New(mockedSession),
	}

	id, err := a.GetScalingGroupForSlotId("prod", "cookie_monster", "first")
	if err != nil {
		t.Fatalf("Failed getting ASG: %s", err)
	}

	expectedId := "my-random-cookie-asg"
	if id != expectedId {
		t.Fatalf("Wrong ASG received.\nGiven: %q\nExpected: %q\n",
			id, expectedId)
	}
}

func TestDescribeBalancedInstanceHealth(t *testing.T) {
	routes := []*MockRoute{
		&MockRoute{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeInstanceHealth&LoadBalancerName=my-lb&Version=2012-06-01",
			Response: MockResponse{
				Code: 200,
				Body: test_elb_describeInstanceHealth_apiBody,
			},
		},
	}
	mockedSession, closeFunc := GetMockedAwsSession(routes, "us-east-1")
	defer closeFunc()

	a := &AWS{
		Region:  awsSDK.String("us-east-1"),
		elbConn: elb.New(mockedSession),
	}

	healths, err := a.DescribeBalancedInstanceHealth("my-lb")
	if err != nil {
		t.Fatalf("Failed getting ASG: %s", err)
	}

	expectedHealths := []*InstanceHealth{
		&InstanceHealth{"i-90d8c2a5", "InService"},
		&InstanceHealth{"i-aaaaaaaa", "InService"},
		&InstanceHealth{"i-00000000", "OutOfService"},
	}
	if !reflect.DeepEqual(healths, expectedHealths) {
		t.Fatalf("Wrong healths received.\nGiven: %q\nExpected: %q\n",
			healths, expectedHealths)
	}
}

func TestGetBalancersForApp(t *testing.T) {
	routes := []*MockRoute{
		&MockRoute{
			ExpectedURI: "/",
			ExpectedRequestBody: "Action=DescribeLoadBalancers&" +
				"LoadBalancerNames=&Version=2012-06-01",
			Response: MockResponse{
				Code: 200,
				Body: test_elb_describeLoadBalancers_apiBody,
			},
		},
		&MockRoute{
			ExpectedURI: "/",
			ExpectedRequestBody: "Action=DescribeTags&" +
				"LoadBalancerNames.member.1=my-loadbalancer&" +
				"LoadBalancerNames.member.2=second-elb&" +
				"Version=2012-06-01",
			Response: MockResponse{
				Code: 200,
				Body: test_elb_describeTags_apiBody,
			},
		},
	}
	mockedSession, closeFunc := GetMockedAwsSession(routes, "us-east-1")
	defer closeFunc()

	a := &AWS{
		Region:  awsSDK.String("us-east-1"),
		elbConn: elb.New(mockedSession),
	}

	balancerNames, err := a.GetBalancersForApp("tasty_cookie_generator")
	if err != nil {
		t.Fatalf("Failed getting balancer names for app: %s", err)
	}

	expectedBalancerNames := []string{"my-loadbalancer", "second-elb"}
	if !reflect.DeepEqual(balancerNames, expectedBalancerNames) {
		t.Fatalf("Wrong balancer names received.\nGiven: %q\nExpected: %q\n",
			balancerNames, expectedBalancerNames)
	}
}

func TestDetachBalancersFromScalingGroup(t *testing.T) {
	routes := []*MockRoute{
		&MockRoute{
			ExpectedURI: "/",
			ExpectedRequestBody: "Action=DetachLoadBalancers&" +
				"AutoScalingGroupName=asg-xyz&" +
				"LoadBalancerNames.member.1=lb-1&" +
				"LoadBalancerNames.member.2=lb-2&" +
				"Version=2011-01-01",
			Response: MockResponse{
				Code: 200,
				Body: "",
			},
		},
	}
	mockedSession, closeFunc := GetMockedAwsSession(routes, "us-east-1")
	defer closeFunc()

	a := &AWS{
		Region:          awsSDK.String("us-east-1"),
		autoscalingConn: autoscaling.New(mockedSession),
	}

	err := a.DetachBalancersFromScalingGroup([]string{"lb-1", "lb-2"}, "asg-xyz")
	if err != nil {
		t.Fatalf("Failed detaching LBs: %s", err)
	}
}

func TestAttachBalancersToScalingGroup(t *testing.T) {
	routes := []*MockRoute{
		&MockRoute{
			ExpectedURI: "/",
			ExpectedRequestBody: "Action=AttachLoadBalancers&" +
				"AutoScalingGroupName=asg-xyz&" +
				"LoadBalancerNames.member.1=lb-1&" +
				"LoadBalancerNames.member.2=lb-2&" +
				"Version=2011-01-01",
			Response: MockResponse{
				Code: 200,
				Body: "",
			},
		},
	}
	mockedSession, closeFunc := GetMockedAwsSession(routes, "us-east-1")
	defer closeFunc()

	a := &AWS{
		Region:          awsSDK.String("us-east-1"),
		autoscalingConn: autoscaling.New(mockedSession),
	}

	err := a.AttachBalancersToScalingGroup([]string{"lb-1", "lb-2"}, "asg-xyz")
	if err != nil {
		t.Fatalf("Failed detaching LBs: %s", err)
	}
}

var test_autoscaling_balancers_apiBody = `<DescribeLoadBalancersResponse xmlns="http://autoscaling.amazonaws.com/doc/2011-01-01/">
  <DescribeLoadBalancersResult>
    <LoadBalancers>
      <member>
        <LoadBalancerName>internal-loadbalancer</LoadBalancerName>
        <State>Added</State>
      </member>
      <member>
        <LoadBalancerName>external-loadbalancer</LoadBalancerName>
        <State>Added</State>
      </member>
    </LoadBalancers>
  </DescribeLoadBalancersResult>
  <ResponseMetadata>
    <RequestId>7c6e177f-f082-11e1-ac58-3714bEXAMPLE</RequestId>
  </ResponseMetadata>
</DescribeLoadBalancersResponse>`

var test_autoscaling_describeGroups_apiBody = `<DescribeAutoScalingGroupsResponse xmlns="http://autoscaling.amazonaws.com/doc/2011-01-01/">
  <DescribeAutoScalingGroupsResult>
    <AutoScalingGroups>
      <member>
        <HealthCheckType>ELB</HealthCheckType>
        <LoadBalancerNames>
          <member>my-loadbalancer</member>
        </LoadBalancerNames>
        <Instances>
          <member>
            <LaunchConfigurationName>my-lc</LaunchConfigurationName>
            <LifecycleState>InService</LifecycleState>
            <InstanceId>i-12345678</InstanceId>
            <ProtectedFromScaleIn>false</ProtectedFromScaleIn>
            <AvailabilityZone>us-east-1c</AvailabilityZone>
          </member>
          <member>
            <LaunchConfigurationName>my-lc</LaunchConfigurationName>
            <LifecycleState>InService</LifecycleState>
            <InstanceId>i-01111108</InstanceId>
            <ProtectedFromScaleIn>false</ProtectedFromScaleIn>
            <AvailabilityZone>us-east-1c</AvailabilityZone>
          </member>
        </Instances>
        <TerminationPolicies>
          <member>Default</member>
        </TerminationPolicies>
        <DefaultCooldown>300</DefaultCooldown>
        <AutoScalingGroupARN>arn:aws:autoscaling:us-east-1:123456789012:autoScalingGroup:12345678-1234-1234-1234-123456789012:autoScalingGroupName/my-asg</AutoScalingGroupARN>
        <EnabledMetrics />
        <AvailabilityZones>
          <member>us-east-1b</member>
          <member>us-east-1a</member>
        </AvailabilityZones>
        <Tags>
          <member>
            <ResourceId>my-asg</ResourceId>
            <PropagateAtLaunch>true</PropagateAtLaunch>
            <Value>elmo</Value>
            <Key>App</Key>
            <ResourceType>auto-scaling-group</ResourceType>
          </member>
          <member>
            <ResourceId>my-asg</ResourceId>
            <PropagateAtLaunch>true</PropagateAtLaunch>
            <Value>test</Value>
            <Key>environment</Key>
            <ResourceType>auto-scaling-group</ResourceType>
          </member>
        </Tags>
        <LaunchConfigurationName>my-lc</LaunchConfigurationName>
        <AutoScalingGroupName>my-asg</AutoScalingGroupName>
        <HealthCheckGracePeriod>300</HealthCheckGracePeriod>
        <NewInstancesProtectedFromScaleIn>false</NewInstancesProtectedFromScaleIn>
        <SuspendedProcesses />
        <CreatedTime>2015-05-06T17:47:15.107Z</CreatedTime>
        <MinSize>2</MinSize>
        <MaxSize>10</MaxSize>
        <DesiredCapacity>2</DesiredCapacity>
        <VPCZoneIdentifier>subnet-12345678,subnet-98765432</VPCZoneIdentifier>
      </member>
    </AutoScalingGroups>
  </DescribeAutoScalingGroupsResult>
  <ResponseMetadata>
    <RequestId>7c6e177f-f082-11e1-ac58-3714bEXAMPLE</RequestId>
  </ResponseMetadata>
</DescribeAutoScalingGroupsResponse>`

var test_autoscaling_describeTags_apiBody = `<DescribeTagsResponse xmlns="http://autoscaling.amazonaws.com/doc/2011-01-01/">
  <DescribeTagsResult>
    <Tags>
      <member>
        <ResourceId>my-random-cookie-asg</ResourceId>
        <PropagateAtLaunch>true</PropagateAtLaunch>
        <Value>prod-cookie_monster-vfirst-vinst</Value>
        <Key>Name</Key>
        <ResourceType>auto-scaling-group</ResourceType>
      </member>
    </Tags>
  </DescribeTagsResult>
  <ResponseMetadata>
    <RequestId>7c6e177f-f082-11e1-ac58-3714bEXAMPLE</RequestId>
  </ResponseMetadata>
</DescribeTagsResponse>`

var test_elb_describeInstanceHealth_apiBody = `<DescribeInstanceHealthResponse xmlns="http://elasticloadbalancing.amazonaws.com/doc/2012-06-01/">
  <DescribeInstanceHealthResult>
    <InstanceStates>
      <member>
        <Description>N/A</Description>
        <InstanceId>i-90d8c2a5</InstanceId>
        <State>InService</State>
        <ReasonCode>N/A</ReasonCode>
      </member>
      <member>
        <Description>N/A</Description>
        <InstanceId>i-aaaaaaaa</InstanceId>
        <State>InService</State>
        <ReasonCode>N/A</ReasonCode>
      </member>
      <member>
        <Description>N/A</Description>
        <InstanceId>i-00000000</InstanceId>
        <State>OutOfService</State>
        <ReasonCode>N/A</ReasonCode>
      </member>
    </InstanceStates>
  </DescribeInstanceHealthResult>
  <ResponseMetadata>
    <RequestId>1549581b-12b7-11e3-895e-1334aEXAMPLE</RequestId>
  </ResponseMetadata>
</DescribeInstanceHealthResponse>`

var test_elb_describeLoadBalancers_apiBody = `<DescribeLoadBalancersResponse xmlns="http://elasticloadbalancing.amazonaws.com/doc/2012-06-01/">
  <DescribeLoadBalancersResult>
    <LoadBalancerDescriptions>
      <member>
        <SecurityGroups/>
        <LoadBalancerName>my-loadbalancer</LoadBalancerName>
        <CreatedTime>2013-05-24T21:15:31.280Z</CreatedTime>
        <HealthCheck>
          <Interval>90</Interval>
          <Target>HTTP:80/</Target>
          <HealthyThreshold>2</HealthyThreshold>
          <Timeout>60</Timeout>
          <UnhealthyThreshold>10</UnhealthyThreshold>
        </HealthCheck>
        <ListenerDescriptions>
          <member>
            <PolicyNames/>
            <Listener>
              <Protocol>HTTP</Protocol>
              <LoadBalancerPort>80</LoadBalancerPort>
              <InstanceProtocol>HTTP</InstanceProtocol>
              <InstancePort>80</InstancePort>
            </Listener>
          </member>
        </ListenerDescriptions>
        <Instances>
          <member>
            <InstanceId>i-e4cbe38d</InstanceId>
          </member>
        </Instances>
        <Policies>
          <AppCookieStickinessPolicies/>
          <OtherPolicies/>
          <LBCookieStickinessPolicies/>
        </Policies>
        <AvailabilityZones>
          <member>us-east-1a</member>
        </AvailabilityZones>
        <CanonicalHostedZoneNameID>ZZZZZZZZZZZ123X</CanonicalHostedZoneNameID>
        <CanonicalHostedZoneName>my-loadbalancer-123456789.us-east-1.elb.amazonaws.com</CanonicalHostedZoneName>
        <Scheme>internet-facing</Scheme>
        <SourceSecurityGroup>
          <OwnerAlias>amazon-elb</OwnerAlias>
          <GroupName>amazon-elb-sg</GroupName>
        </SourceSecurityGroup>
        <DNSName>my-loadbalancer-123456789.us-east-1.elb.amazonaws.com</DNSName>
        <BackendServerDescriptions/>
        <Subnets/>
      </member>
      <member>
        <SecurityGroups/>
        <LoadBalancerName>second-elb</LoadBalancerName>
        <CreatedTime>2013-05-24T21:15:31.280Z</CreatedTime>
        <HealthCheck>
          <Interval>90</Interval>
          <Target>HTTP:80/</Target>
          <HealthyThreshold>2</HealthyThreshold>
          <Timeout>60</Timeout>
          <UnhealthyThreshold>10</UnhealthyThreshold>
        </HealthCheck>
        <ListenerDescriptions>
          <member>
            <PolicyNames/>
            <Listener>
              <Protocol>HTTP</Protocol>
              <LoadBalancerPort>80</LoadBalancerPort>
              <InstanceProtocol>HTTP</InstanceProtocol>
              <InstancePort>80</InstancePort>
            </Listener>
          </member>
        </ListenerDescriptions>
        <Instances>
          <member>
            <InstanceId>i-e4cbe38d</InstanceId>
          </member>
        </Instances>
        <Policies>
          <AppCookieStickinessPolicies/>
          <OtherPolicies/>
          <LBCookieStickinessPolicies/>
        </Policies>
        <AvailabilityZones>
          <member>us-east-1a</member>
        </AvailabilityZones>
        <CanonicalHostedZoneNameID>ZZZZZZZZZZZ123X</CanonicalHostedZoneNameID>
        <CanonicalHostedZoneName>second-elb-123456789.us-east-1.elb.amazonaws.com</CanonicalHostedZoneName>
        <Scheme>internet-facing</Scheme>
        <SourceSecurityGroup>
          <OwnerAlias>amazon-elb</OwnerAlias>
          <GroupName>amazon-elb-sg</GroupName>
        </SourceSecurityGroup>
        <DNSName>my-loadbalancer-123456789.us-east-1.elb.amazonaws.com</DNSName>
        <BackendServerDescriptions/>
        <Subnets/>
      </member>
    </LoadBalancerDescriptions>
  </DescribeLoadBalancersResult>
  <ResponseMetadata>
    <RequestId>83c88b9d-12b7-11e3-8b82-87b12EXAMPLE</RequestId>
  </ResponseMetadata>
</DescribeLoadBalancersResponse>`

var test_elb_describeTags_apiBody = `<DescribeTagsResponse xmlns="http://elasticloadbalancing.amazonaws.com/doc/2012-06-01/">
  <DescribeTagsResult>
    <TagDescriptions>
      <member>
        <Tags>
          <member>
            <Value>tasty_cookie_generator</Value>
            <Key>App</Key>
          </member>
          <member>
            <Value>digital-media</Value>
            <Key>department</Key>
          </member>
        </Tags>
        <LoadBalancerName>my-loadbalancer</LoadBalancerName>
      </member>
      <member>
        <Tags>
          <member>
            <Value>tasty_cookie_generator</Value>
            <Key>App</Key>
          </member>
          <member>
            <Value>digital-media</Value>
            <Key>department</Key>
          </member>
        </Tags>
        <LoadBalancerName>second-elb</LoadBalancerName>
      </member>
    </TagDescriptions>
  </DescribeTagsResult>
  <ResponseMetadata>
    <RequestId>07b1ecbc-1100-11e3-acaf-dd7edEXAMPLE</RequestId>
  </ResponseMetadata>
</DescribeTagsResponse>`
