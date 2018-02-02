package schema

import (
	"testing"
)

func TestApplication_v0_to_v1(t *testing.T) {
	v0_data := `{"v":0,"use_central_git_repo":true,"is_active":true,"infra_outputs":{"one":"1111","two":"22222"},"last_rt_version":"old","last_terraform_version":"old-as-hell","last_deployment_time":"2016-03-30T15:04:05+01:00","last_infra_change_time":"2016-03-30T15:04:05+01:00","slot_counters":{"stable":1234}}`
	expected_v1 := `{"v":1,"use_central_git_repo":true,"is_active":true,"infra_outputs":{"one":"1111","two":"22222"},"last_rt_version":"old","last_terraform_version":"old-as-hell","last_deployment_time":"2016-03-30T15:04:05+01:00","last_infra_change_time":"2016-03-30T15:04:05+01:00","slot_counters":{"stable":1234}}`

	ad := &ApplicationData{}
	ad.FromJSON([]byte(v0_data))

	if ad.SchemaVersion != 1 {
		t.Fatalf("Expected schema v1, v%d given", ad.SchemaVersion)
	}

	b, err := ad.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	migratedData := string(b)
	if expected_v1 != migratedData {
		t.Fatalf("Unexpected data after migration.\nExpected: %s\nGiven: %s\n",
			v0_data, migratedData)
	}
}

func TestSlot_v0_to_v1(t *testing.T) {
	v0_data := `{"v":0,"is_active":true,"last_deployment_start_time":"2016-09-12T13:52:12.853050642Z","last_deploy_pilot":{"aws_api_caller":"arn:aws:iam::123456789012:user/rsimko1016","ip_address":"8.8.8.8"},"last_terraform_run":{"plan_start_time":"2016-09-12T13:51:54.799975226Z","plan_finish_time":"2016-09-12T13:52:01.326031911Z","start_time":"2016-09-12T13:52:14.828070419Z","finish_time":"2016-09-12T13:53:49.447318858Z","is_destroy":false,"resource_diff":{"Created":5,"Removed":0,"Changed":0},"variables":{"app_name":"ape_git_cop","app_version":"24a7079","environment":"test"},"outputs":{"app":"git-cop","aws_region":"us-east-1","environment":"test","team":"devops","version":"24a7079","version_asg_desired_capacity":"1","version_asg_health_check_grace_period":"120","version_asg_launch_configuration":"test-git-cop-v24a7079-vlc","version_asg_max_size":"1","version_asg_min_size":"1","version_asg_name":"test-git-cop-v24a7079-vasg","version_launch_configuration_id":"test-git-cop-v24a7079-vlc"},"terraform_version":"0.6.16","stderr":"\u001b[33mWarnings:\n\u001b[0m\u001b[0m\n\u001b[33m  * template_file.cloud_config: \"filename\": [DEPRECATED] Use the 'template' attribute instead.\u001b[0m\u001b[0m\n\u001b[33m\nNo errors found. Continuing with 1 warning(s).\n\u001b[0m\u001b[0m\n"}}`
	expected_v1 := `{"v":1,"is_active":true,"last_deployment_start_time":"2016-09-12T13:52:12.853050642Z","last_deploy_pilot":{"aws_api_caller":"arn:aws:iam::123456789012:user/rsimko1016","ip_address":"8.8.8.8"},"last_terraform_run":{"plan_start_time":"2016-09-12T13:51:54.799975226Z","plan_finish_time":"2016-09-12T13:52:01.326031911Z","start_time":"2016-09-12T13:52:14.828070419Z","finish_time":"2016-09-12T13:53:49.447318858Z","is_destroy":false,"resource_diff":{"Created":5,"Removed":0,"Changed":0},"variables":{"app_name":"ape_git_cop","app_version":"24a7079","environment":"test"},"outputs":{"app":"git-cop","aws_region":"us-east-1","environment":"test","team":"devops","version":"24a7079","version_asg_desired_capacity":"1","version_asg_health_check_grace_period":"120","version_asg_launch_configuration":"test-git-cop-v24a7079-vlc","version_asg_max_size":"1","version_asg_min_size":"1","version_asg_name":"test-git-cop-v24a7079-vasg","version_launch_configuration_id":"test-git-cop-v24a7079-vlc"},"terraform_version":"0.6.16","stderr":"\u001b[33mWarnings:\n\u001b[0m\u001b[0m\n\u001b[33m  * template_file.cloud_config: \"filename\": [DEPRECATED] Use the 'template' attribute instead.\u001b[0m\u001b[0m\n\u001b[33m\nNo errors found. Continuing with 1 warning(s).\n\u001b[0m\u001b[0m\n"}}`

	ad := &SlotData{}
	ad.FromJSON([]byte(v0_data))

	if ad.SchemaVersion != 1 {
		t.Fatalf("Expected schema v1, v%d given", ad.SchemaVersion)
	}

	b, err := ad.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	migratedData := string(b)
	if expected_v1 != migratedData {
		t.Fatalf("Unexpected data after migration.\nExpected: %s\nGiven: %s\n",
			v0_data, migratedData)
	}
}

func TestDeployment_v0_to_v1(t *testing.T) {
	v0_data := `{"v":0,"deploy_pilot":{"aws_api_caller":"arn:aws:iam::123456789012:user/rsimko1016","ip_address":"8.8.8.8"},"start_time":"2016-09-12T13:52:12.853050642Z","terraform":{"plan_start_time":"2016-09-12T13:51:54.799975226Z","plan_finish_time":"2016-09-12T13:52:01.326031911Z","start_time":"2016-09-12T13:52:14.828070419Z","finish_time":"2016-09-12T13:53:49.447318858Z","is_destroy":false,"resource_diff":{"Created":5,"Removed":0,"Changed":0},"variables":{"app_name":"ape_git_cop","app_version":"24a7079","environment":"test"},"outputs":{"app":"git-cop","aws_region":"us-east-1","environment":"test","team":"devops","version":"24a7079","version_asg_desired_capacity":"1","version_asg_health_check_grace_period":"120","version_asg_launch_configuration":"test-git-cop-v24a7079-vlc","version_asg_max_size":"1","version_asg_min_size":"1","version_asg_name":"test-git-cop-v24a7079-vasg","version_launch_configuration_id":"test-git-cop-v24a7079-vlc"},"terraform_version":"0.6.16","stderr":"\u001b[33mWarnings:\n\u001b[0m\u001b[0m\n\u001b[33m  * template_file.cloud_config: \"filename\": [DEPRECATED] Use the 'template' attribute instead.\u001b[0m\u001b[0m\n\u001b[33m\nNo errors found. Continuing with 1 warning(s).\n\u001b[0m\u001b[0m\n"},"rt_version":"0.5.0"}`
	expected_v1 := `{"v":1,"deploy_pilot":{"aws_api_caller":"arn:aws:iam::123456789012:user/rsimko1016","ip_address":"8.8.8.8"},"start_time":"2016-09-12T13:52:12.853050642Z","terraform":{"plan_start_time":"2016-09-12T13:51:54.799975226Z","plan_finish_time":"2016-09-12T13:52:01.326031911Z","start_time":"2016-09-12T13:52:14.828070419Z","finish_time":"2016-09-12T13:53:49.447318858Z","is_destroy":false,"resource_diff":{"Created":5,"Removed":0,"Changed":0},"variables":{"app_name":"ape_git_cop","app_version":"24a7079","environment":"test"},"outputs":{"app":"git-cop","aws_region":"us-east-1","environment":"test","team":"devops","version":"24a7079","version_asg_desired_capacity":"1","version_asg_health_check_grace_period":"120","version_asg_launch_configuration":"test-git-cop-v24a7079-vlc","version_asg_max_size":"1","version_asg_min_size":"1","version_asg_name":"test-git-cop-v24a7079-vasg","version_launch_configuration_id":"test-git-cop-v24a7079-vlc"},"terraform_version":"0.6.16","stderr":"\u001b[33mWarnings:\n\u001b[0m\u001b[0m\n\u001b[33m  * template_file.cloud_config: \"filename\": [DEPRECATED] Use the 'template' attribute instead.\u001b[0m\u001b[0m\n\u001b[33m\nNo errors found. Continuing with 1 warning(s).\n\u001b[0m\u001b[0m\n"},"rt_version":"0.5.0"}`

	ad := &DeploymentData{}
	ad.FromJSON([]byte(v0_data))

	if ad.SchemaVersion != 1 {
		t.Fatalf("Expected schema v1, v%d given", ad.SchemaVersion)
	}

	b, err := ad.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	migratedData := string(b)
	if expected_v1 != migratedData {
		t.Fatalf("Unexpected data after migration.\nExpected: %s\nGiven: %s\n",
			v0_data, migratedData)
	}
}
