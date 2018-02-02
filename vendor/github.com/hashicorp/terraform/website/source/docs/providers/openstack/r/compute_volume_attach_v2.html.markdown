---
layout: "openstack"
page_title: "OpenStack: openstack_compute_volume_attach_v2"
sidebar_current: "docs-openstack-resource-compute-volume-attach-v2"
description: |-
  Attaches a Block Storage Volume to an Instance.
---

# openstack\_compute\_volume_attach_v2

Attaches a Block Storage Volume to an Instance using the OpenStack
Compute (Nova) v2 API.

## Example Usage

```
resource "openstack_blockstorage_volume_v2" "volume_1" {
  name = "volume_1"
  size = 1
}

resource "openstack_compute_instance_v2" "instance_1" {
  name = "instance_1"
  security_groups = ["default"]
}

resource "openstack_compute_volume_attach_v2" "va_1" {
  instance_id = "${openstack_compute_instance_v2.instance_1.id}"
  volume_id = "${openstack_blockstorage_volume_v2.volume_1.id}"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Required) The region in which to obtain the V2 Compute client.
    A Compute client is needed to create a volume attachment. If omitted, the
    `OS_REGION_NAME` environment variable is used. Changing this creates a
    new volume attachment.

* `instance_id` - (Required) The ID of the Instance to attach the Volume to.

* `volume_id` - (Required) The ID of the Volume to attach to an Instance.

## Attributes Reference

The following attributes are exported:

* `region` - See Argument Reference above.
* `instance_id` - See Argument Reference above.
* `volume_id` - See Argument Reference above.
* `device` - The device of the volume attachment (ex: `/dev/vdc`).
  _NOTE_: This is the device reported by the Compute API and the real device
  might actually differ depending on the hypervisor being used. This should
  not be used as an authoritative piece of information.

## Import

Volume Attachments can be imported using the Instance ID and Volume ID
separated by a slash, e.g.

```
$ terraform import openstack_compute_volume_attach_v2.va_1 89c60255-9bd6-460c-822a-e2b959ede9d2/45670584-225f-46c3-b33e-6707b589b666
```
