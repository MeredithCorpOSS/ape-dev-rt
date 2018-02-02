---
layout: "aws"
page_title: "AWS: aws_db_option_group"
sidebar_current: "docs-aws-resource-db-option-group"
---

# aws\_db\_option\_group

Provides an RDS DB option group resource.

## Example Usage

```
resource "aws_db_option_group" "bar" {
  name = "option-group-test-terraform"
  option_group_description = "Terraform Option Group"
  engine_name = "sqlserver-ee"
  major_engine_version = "11.00"

  option {
    option_name = "Timezone"
    option_settings {
      name = "TIME_ZONE"
      value = "UTC"
    }
  }

  option {
    option_name = "TDE"
  }
}
```

~> **Note**: Any modifications to the `db_option_group` are set to happen immediately as we default to applying immediately.

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Option group to be created.
* `option_group_description` - (Required) The description of the option group.
* `engine_name` - (Required) Specifies the name of the engine that this option group should be associated with..
* `major_engine_version` - (Required) Specifies the major version of the engine that this option group should be associated with.
* `option` - (Optional) A list of Options to apply.
* `tags` - (Optional) A mapping of tags to assign to the resource.

Option blocks support the following:

* `option_name` - (Required) The Name of the Option (e.g. MEMCACHED).
* `option_settings` - (Optional) A list of option settings to apply.
* `port` - (Optional) The Port number when connecting to the Option (e.g. 11211).
* `db_security_group_memberships` - (Optional) A list of DB Security Groups for which the option is enabled.
* `vpc_security_group_memberships` - (Optional) A list of VPC Security Groups for which the option is enabled.

Option Settings blocks support the following:

* `name` - (Optional) The Name of the setting.
* `value` - (Optional) The Value of the setting.

## Attributes Reference

The following attributes are exported:

* `arn` - The ARN of the db option group.

## Import

DB Option groups can be imported using the `name`, e.g. 

```
$ terraform import aws_db_option_group.bar mysql-option-group
```