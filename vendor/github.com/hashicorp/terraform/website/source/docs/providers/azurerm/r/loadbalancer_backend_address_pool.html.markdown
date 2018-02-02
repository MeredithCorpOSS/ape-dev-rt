---
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_lb_backend_address_pool"
sidebar_current: "docs-azurerm-resource-loadbalancer-backend-address-pool"
description: |-
  Create a LoadBalancer Backend Address Pool.
---

# azurerm\_lb\_backend\_address\_pool

Create a LoadBalancer Backend Address Pool.

~> **NOTE When using this resource, the LoadBalancer needs to have a FrontEnd IP Configuration Attached

## Example Usage

```
resource "azurerm_resource_group" "test" {
    name = "LoadBalancerRG"
    location = "West US"
}

resource "azurerm_public_ip" "test" {
    name = "PublicIPForLB"
    location = "West US"
    resource_group_name = "${azurerm_resource_group.test.name}"
    public_ip_address_allocation = "static"
}

resource "azurerm_lb" "test" {
    name = "TestLoadBalancer"
    location = "West US"
    resource_group_name = "${azurerm_resource_group.test.name}"

    frontend_ip_configuration {
      name = "PublicIPAddress"
      public_ip_address_id = "${azurerm_public_ip.test.id}"
    }
}

resource "azurerm_lb_backend_address_pool" "test" {
  location = "West US"
  resource_group_name = "${azurerm_resource_group.test.name}"
  loadbalancer_id = "${azurerm_lb.test.id}"
  name = "BackEndAddressPool"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the Backend Address Pool.
* `resource_group_name` - (Required) The name of the resource group in which to create the resource.
* `location` - (Required) Specifies the supported Azure location where the resource exists.
* `loadbalancer_id` - (Required) The ID of the LoadBalancer in which to create the Backend Address Pool. 

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the LoadBalancer to which the resource is attached.
