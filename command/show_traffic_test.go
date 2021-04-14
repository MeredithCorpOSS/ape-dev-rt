package command

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/MeredithCorpOSS/ape-dev-rt/aws"
)

func TestDecorateAndPrintSortedVersions(t *testing.T) {
	cDate, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Wed Nov 23 11:53:50 +0000 MST 2016")
	if err != nil {
		t.Fatal(err)
	}
	slots := []*slotData{
		{
			SlotId: "stable13",
			Variables: fmt.Sprintf("%s", sortedVars(map[string]string{
				"app_name":         "decanter-wine-api",
				"app_version":      "stable13",
				"docker_image_tag": "0.0.2",
				"environment":      "test",
			})),
			FinishTime: cDate,
		},
	}

	autoscalingRoutes := []*aws.MockRoute{
		{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeTags&Filters.member.1.Name=key&Filters.member.1.Values.member.1=Name&Filters.member.2.Name=value&Filters.member.2.Values.member.1=test-decanter-wine-api-vstable13-vinst&Version=2011-01-01",
			Response: aws.MockResponse{
				Code: 200,
				Body: test_asg_DescribeTags_body,
			},
		},
		{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeLoadBalancers&AutoScalingGroupName=test-decanter-wine-api-vstable13-vasg&Version=2011-01-01",
			Response: aws.MockResponse{
				Code: 200,
				Body: test_asg_DescribeLoadBalancers_body,
			},
		},
		{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeAutoScalingGroups&AutoScalingGroupNames.member.1=test-decanter-wine-api-vstable13-vasg&Version=2011-01-01",
			Response: aws.MockResponse{
				Code: 200,
				Body: test_asg_DescribeAutoScalingGroups_body,
			},
		},
	}

	autoscalingSession, closeFunc := aws.GetMockedAwsSession(autoscalingRoutes, "us-east-1")
	defer closeFunc()

	elbRoutes := []*aws.MockRoute{
		{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeLoadBalancers&AutoScalingGroupName=test-decanter-wine-api-vstable13-vasg&Version=2011-01-01",
			Response: aws.MockResponse{
				Code: 200,
				Body: test_elb_DescribeLoadBalancers_body,
			},
		},
		{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeInstanceHealth&LoadBalancerName=tf-lb-decanter-wine-api&Version=2012-06-01",
			Response: aws.MockResponse{
				Code: 200,
				Body: test_elb_DescribeInstanceHealth_body,
			},
		},
	}

	elbSession, closeFunc := aws.GetMockedAwsSession(elbRoutes, "us-east-1")
	defer closeFunc()

	ec2Routes := []*aws.MockRoute{
		{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=DescribeInstances&InstanceId.1=i-ee546206&Version=2016-11-15",
			Response: aws.MockResponse{
				Code: 200,
				Body: test_ec2_DescribeInstances_body,
			},
		},
	}

	ec2Session, closeFunc := aws.GetMockedAwsSession(ec2Routes, "us-east-1")
	defer closeFunc()

	a := aws.MockedAWS(&aws.MockedAWSInput{
		Region:          "us-east-1",
		AutoscalingSess: autoscalingSession,
		ElbSess:         elbSession,
		Ec2Sess:         ec2Session,
	})
	b := bytes.NewBufferString("")

	noColour := func(s string) string {
		return s
	}
	c := &colours{
		boldGreen:  noColour,
		boldWhite:  noColour,
		boldYellow: noColour,
		boldRed:    noColour,
		boldBlue:   noColour,
		red:        noColour,
		green:      noColour,
	}

	err = decorateAndPrintSortedVersions("decanter-wine-api", "test", slots, a, b, c)
	if err != nil {
		t.Fatal(err)
	}

	output := b.String()
	expectedOutput := `Discovering resources in AWS region us-east-1

stable13 (Wed Nov 23 11:53:50 +0000 2016) vars: map["app_name":"decanter-wine-api" "app_version":"stable13" "docker_image_tag":"0.0.2" "environment":"test"] - 1 ELB
  ELB tf-lb-decanter-wine-api Added to ASG test-decanter-wine-api-vstable13-vasg
    EC2 i-ee546206 InService (this version) - 10.108.38.201
`
	if output != expectedOutput {
		t.Fatalf("Unexpected output!\nExpected: %q\nGiven: %q\n", expectedOutput, output)
	}
}

var test_asg_DescribeTags_body = `<DescribeTagsResponse xmlns="http://autoscaling.amazonaws.com/doc/2011-01-01/">
  <DescribeTagsResult>
    <Tags>
      <member>
        <ResourceId>test-decanter-wine-api-vstable13-vasg</ResourceId>
        <PropagateAtLaunch>true</PropagateAtLaunch>
        <Value>test-decanter-wine-api-vstable13-vinst</Value>
        <Key>Name</Key>
        <ResourceType>auto-scaling-group</ResourceType>
      </member>
    </Tags>
  </DescribeTagsResult>
  <ResponseMetadata>
    <RequestId>e3716c0c-b632-11e6-924a-e73cbafc637f</RequestId>
  </ResponseMetadata>
</DescribeTagsResponse>`

var test_asg_DescribeLoadBalancers_body = `<DescribeLoadBalancersResponse xmlns="http://autoscaling.amazonaws.com/doc/2011-01-01/">
  <DescribeLoadBalancersResult>
    <LoadBalancers>
      <member>
        <LoadBalancerName>tf-lb-decanter-wine-api</LoadBalancerName>
        <State>Added</State>
      </member>
    </LoadBalancers>
  </DescribeLoadBalancersResult>
  <ResponseMetadata>
    <RequestId>e3775f7d-b632-11e6-924a-e73cbafc637f</RequestId>
  </ResponseMetadata>
</DescribeLoadBalancersResponse>`

var test_asg_DescribeAutoScalingGroups_body = `<DescribeAutoScalingGroupsResponse xmlns="http://autoscaling.amazonaws.com/doc/2011-01-01/">
  <DescribeAutoScalingGroupsResult>
    <AutoScalingGroups>
      <member>
        <HealthCheckType>EC2</HealthCheckType>
        <LoadBalancerNames>
          <member>tf-lb-decanter-wine-api</member>
        </LoadBalancerNames>
        <Instances>
          <member>
            <LaunchConfigurationName>test-decanter-wine-api-vstable13-vlc</LaunchConfigurationName>
            <LifecycleState>InService</LifecycleState>
            <InstanceId>i-ee546206</InstanceId>
            <HealthStatus>Healthy</HealthStatus>
            <ProtectedFromScaleIn>false</ProtectedFromScaleIn>
            <AvailabilityZone>eu-west-1b</AvailabilityZone>
          </member>
        </Instances>
        <TerminationPolicies>
          <member>Default</member>
        </TerminationPolicies>
        <DefaultCooldown>300</DefaultCooldown>
        <AutoScalingGroupARN>arn:aws:autoscaling:eu-west-1:924684194228:autoScalingGroup:ca5ef2dc-dded-4d78-b48f-ce75db8ad819:autoScalingGroupName/test-decanter-wine-api-vstable13-vasg</AutoScalingGroupARN>
        <EnabledMetrics/>
        <MaxSize>3</MaxSize>
        <AvailabilityZones>
          <member>eu-west-1b</member>
          <member>eu-west-1c</member>
          <member>eu-west-1a</member>
        </AvailabilityZones>
        <TargetGroupARNs/>
        <Tags>
          <member>
            <ResourceId>test-decanter-wine-api-vstable13-vasg</ResourceId>
            <PropagateAtLaunch>true</PropagateAtLaunch>
            <Value>decanter-wine-api</Value>
            <Key>App</Key>
            <ResourceType>auto-scaling-group</ResourceType>
          </member>
          <member>
            <ResourceId>test-decanter-wine-api-vstable13-vasg</ResourceId>
            <PropagateAtLaunch>true</PropagateAtLaunch>
            <Value>test</Value>
            <Key>Environment</Key>
            <ResourceType>auto-scaling-group</ResourceType>
          </member>
          <member>
            <ResourceId>test-decanter-wine-api-vstable13-vasg</ResourceId>
            <PropagateAtLaunch>true</PropagateAtLaunch>
            <Value>test-decanter-wine-api-vstable13-vinst</Value>
            <Key>Name</Key>
            <ResourceType>auto-scaling-group</ResourceType>
          </member>
          <member>
            <ResourceId>test-decanter-wine-api-vstable13-vasg</ResourceId>
            <PropagateAtLaunch>true</PropagateAtLaunch>
            <Value>vinst</Value>
            <Key>Role</Key>
            <ResourceType>auto-scaling-group</ResourceType>
          </member>
          <member>
            <ResourceId>test-decanter-wine-api-vstable13-vasg</ResourceId>
            <PropagateAtLaunch>true</PropagateAtLaunch>
            <Value>decanter</Value>
            <Key>Team</Key>
            <ResourceType>auto-scaling-group</ResourceType>
          </member>
          <member>
            <ResourceId>test-decanter-wine-api-vstable13-vasg</ResourceId>
            <PropagateAtLaunch>true</PropagateAtLaunch>
            <Value>stable13</Value>
            <Key>Version</Key>
            <ResourceType>auto-scaling-group</ResourceType>
          </member>
        </Tags>
        <LaunchConfigurationName>test-decanter-wine-api-vstable13-vlc</LaunchConfigurationName>
        <AutoScalingGroupName>test-decanter-wine-api-vstable13-vasg</AutoScalingGroupName>
        <HealthCheckGracePeriod>120</HealthCheckGracePeriod>
        <NewInstancesProtectedFromScaleIn>false</NewInstancesProtectedFromScaleIn>
        <CreatedTime>2016-11-29T11:05:21.773Z</CreatedTime>
        <MinSize>1</MinSize>
        <SuspendedProcesses/>
        <DesiredCapacity>1</DesiredCapacity>
        <VPCZoneIdentifier>subnet-a9bedade,subnet-878ff3e2,subnet-1551ef4c</VPCZoneIdentifier>
      </member>
    </AutoScalingGroups>
  </DescribeAutoScalingGroupsResult>
  <ResponseMetadata>
    <RequestId>e37bcc4e-b632-11e6-924a-e73cbafc637f</RequestId>
  </ResponseMetadata>
</DescribeAutoScalingGroupsResponse>`

var test_elb_DescribeLoadBalancers_body = `<DescribeLoadBalancersResponse xmlns="http://elasticloadbalancing.amazonaws.com/doc/2015-12-01/">
  <DescribeLoadBalancersResult>
    <LoadBalancers>
      <member>
        <LoadBalancerArn>arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188</LoadBalancerArn>
        <Scheme>internet-facing</Scheme>
        <LoadBalancerName>my-load-balancer</LoadBalancerName>
        <VpcId>vpc-3ac0fb5f</VpcId>
        <CanonicalHostedZoneId>Z2P70J7EXAMPLE</CanonicalHostedZoneId>
        <CreatedTime>2016-03-25T21:26:12.920Z</CreatedTime>
        <AvailabilityZones>
          <member>
            <SubnetId>subnet-8360a9e7</SubnetId>
            <ZoneName>us-west-2a</ZoneName>
          </member>
          <member>
            <SubnetId>subnet-b7d581c0</SubnetId>
            <ZoneName>us-west-2b</ZoneName>
          </member>
        </AvailabilityZones>
        <SecurityGroups>
          <member>sg-5943793c</member>
        </SecurityGroups>
        <DNSName>my-load-balancer-424835706.us-west-2.elb.amazonaws.com</DNSName>
        <State>
          <Code>active</Code>
        </State>
        <Type>application</Type>
      </member>
    </LoadBalancers>
  </DescribeLoadBalancersResult>
  <ResponseMetadata>
    <RequestId>6581c0ac-f39f-11e5-bb98-57195a6eb84a</RequestId>
  </ResponseMetadata>
</DescribeLoadBalancersResponse>`

var test_elb_DescribeInstanceHealth_body = `<DescribeInstanceHealthResponse xmlns="http://elasticloadbalancing.amazonaws.com/doc/2012-06-01/">
  <DescribeInstanceHealthResult>
    <InstanceStates>
      <member>
        <Description>N/A</Description>
        <InstanceId>i-ee546206</InstanceId>
        <ReasonCode>N/A</ReasonCode>
        <State>InService</State>
      </member>
    </InstanceStates>
  </DescribeInstanceHealthResult>
  <ResponseMetadata>
    <RequestId>e412dcbb-b632-11e6-9dd0-e1f5c9454d04</RequestId>
  </ResponseMetadata>
</DescribeInstanceHealthResponse>`

var test_ec2_DescribeInstances_body = `<DescribeInstancesResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
    <requestId>0a549ff0-d1c0-4b02-8312-dc19cd5926a1</requestId>
    <reservationSet>
        <item>
            <reservationId>r-a88d4c20</reservationId>
            <ownerId>924684194228</ownerId>
            <groupSet/>
            <instancesSet>
                <item>
                    <instanceId>i-ee546206</instanceId>
                    <imageId>ami-92035de1</imageId>
                    <instanceState>
                        <code>16</code>
                        <name>running</name>
                    </instanceState>
                    <privateDnsName>ip-10-108-38-201.eu-west-1.compute.internal</privateDnsName>
                    <dnsName/>
                    <reason/>
                    <keyName>tiuk-ops</keyName>
                    <amiLaunchIndex>0</amiLaunchIndex>
                    <productCodes/>
                    <instanceType>t2.medium</instanceType>
                    <launchTime>2016-11-29T11:05:29.000Z</launchTime>
                    <placement>
                        <availabilityZone>eu-west-1b</availabilityZone>
                        <groupName/>
                        <tenancy>default</tenancy>
                    </placement>
                    <monitoring>
                        <state>enabled</state>
                    </monitoring>
                    <subnetId>subnet-878ff3e2</subnetId>
                    <vpcId>vpc-6fba390a</vpcId>
                    <privateIpAddress>10.108.38.201</privateIpAddress>
                    <sourceDestCheck>true</sourceDestCheck>
                    <groupSet>
                        <item>
                            <groupId>sg-1e309278</groupId>
                            <groupName>test-decanter-wine-api-instsg</groupName>
                        </item>
                    </groupSet>
                    <architecture>x86_64</architecture>
                    <rootDeviceType>ebs</rootDeviceType>
                    <rootDeviceName>/dev/xvda</rootDeviceName>
                    <blockDeviceMapping>
                        <item>
                            <deviceName>/dev/xvda</deviceName>
                            <ebs>
                                <volumeId>vol-c4056946</volumeId>
                                <status>attached</status>
                                <attachTime>2016-11-29T11:05:30.000Z</attachTime>
                                <deleteOnTermination>true</deleteOnTermination>
                            </ebs>
                        </item>
                    </blockDeviceMapping>
                    <virtualizationType>hvm</virtualizationType>
                    <clientToken>3f63e426-7f77-4495-9611-7beb93433b35_subnet-878ff3e2_1</clientToken>
                    <tagSet>
                        <item>
                            <key>Version</key>
                            <value>stable13</value>
                        </item>
                        <item>
                            <key>Team</key>
                            <value>decanter</value>
                        </item>
                        <item>
                            <key>Environment</key>
                            <value>test</value>
                        </item>
                        <item>
                            <key>aws:autoscaling:groupName</key>
                            <value>test-decanter-wine-api-vstable13-vasg</value>
                        </item>
                        <item>
                            <key>Role</key>
                            <value>vinst</value>
                        </item>
                        <item>
                            <key>App</key>
                            <value>decanter-wine-api</value>
                        </item>
                        <item>
                            <key>Name</key>
                            <value>test-decanter-wine-api-vstable13-vinst</value>
                        </item>
                    </tagSet>
                    <hypervisor>xen</hypervisor>
                    <networkInterfaceSet>
                        <item>
                            <networkInterfaceId>eni-b79d0cca</networkInterfaceId>
                            <subnetId>subnet-878ff3e2</subnetId>
                            <vpcId>vpc-6fba390a</vpcId>
                            <description/>
                            <ownerId>924684194228</ownerId>
                            <status>in-use</status>
                            <macAddress>02:7c:75:87:e1:4b</macAddress>
                            <privateIpAddress>10.108.38.201</privateIpAddress>
                            <privateDnsName>ip-10-108-38-201.eu-west-1.compute.internal</privateDnsName>
                            <sourceDestCheck>true</sourceDestCheck>
                            <groupSet>
                                <item>
                                    <groupId>sg-1e309278</groupId>
                                    <groupName>test-decanter-wine-api-instsg</groupName>
                                </item>
                            </groupSet>
                            <attachment>
                                <attachmentId>eni-attach-50a63adf</attachmentId>
                                <deviceIndex>0</deviceIndex>
                                <status>attached</status>
                                <attachTime>2016-11-29T11:05:29.000Z</attachTime>
                                <deleteOnTermination>true</deleteOnTermination>
                            </attachment>
                            <privateIpAddressesSet>
                                <item>
                                    <privateIpAddress>10.108.38.201</privateIpAddress>
                                    <privateDnsName>ip-10-108-38-201.eu-west-1.compute.internal</privateDnsName>
                                    <primary>true</primary>
                                </item>
                            </privateIpAddressesSet>
                        </item>
                    </networkInterfaceSet>
                    <iamInstanceProfile>
                        <arn>arn:aws:iam::924684194228:instance-profile/test-decanter-wine-api-instprofile</arn>
                        <id>AIPAI6ILP642PR3AIHC3K</id>
                    </iamInstanceProfile>
                    <ebsOptimized>false</ebsOptimized>
                </item>
            </instancesSet>
            <requesterId>226008221399</requesterId>
        </item>
    </reservationSet>
</DescribeInstancesResponse>`
