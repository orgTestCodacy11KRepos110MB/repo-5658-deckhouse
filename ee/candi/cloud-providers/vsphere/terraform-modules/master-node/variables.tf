# Copyright 2021 Flant CJSC
# Licensed under the Deckhouse Platform Enterprise Edition (EE) license. See https://github.com/deckhouse/deckhouse/ee/LICENSE

variable "clusterConfiguration" {
  type = any
}

variable "providerClusterConfiguration" {
  type = any

  validation {
    condition     = cidrsubnet(var.providerClusterConfiguration.internalNetworkCIDR, 0, 0) == var.providerClusterConfiguration.internalNetworkCIDR
    error_message = "Invalid internalNetworkCIDR in VsphereClusterConfiguration."
  }

  validation {
    condition     = contains(keys(var.providerClusterConfiguration.masterNodeGroup.instanceClass), "mainNetworkIPAddresses") ? length(var.providerClusterConfiguration.masterNodeGroup.instanceClass.mainNetworkIPAddresses) == length([for s in var.providerClusterConfiguration.masterNodeGroup.instanceClass.mainNetworkIPAddresses : s if s.address != cidrsubnet(s.address, 0, 0)]) : true
    error_message = "Invalid address in mainNetworkIPAddresses."
  }

}

variable "nodeIndex" {
  type    = number
  default = 0
}

variable "cloudConfig" {
  type    = string
  default = ""
}

variable "clusterUUID" {
  type = string
}

locals {
  prefix = var.clusterConfiguration.cloud.prefix

  master_instance_class = var.providerClusterConfiguration.masterNodeGroup.instanceClass

  actual_zones    = var.providerClusterConfiguration.zones
  zones           = lookup(var.providerClusterConfiguration.masterNodeGroup, "zones", null) != null ? tolist(setintersection(local.actual_zones, var.providerClusterConfiguration.masterNodeGroup["zones"])) : local.actual_zones
  zone            = element(local.zones, var.nodeIndex)

  base_resource_pool    = trim(lookup(var.providerClusterConfiguration, "baseResourcePool", ""), "/")
  default_resource_pool = join("/", local.base_resource_pool != "" ? [local.base_resource_pool, local.prefix] : [local.prefix])

  resource_pool      = lookup(local.master_instance_class, "resourcePool", local.default_resource_pool)
  additionalNetworks = lookup(local.master_instance_class, "additionalNetworks", [])
  main_ip_addresses  = lookup(local.master_instance_class, "mainNetworkIPAddresses", [])

  runtime_options               = lookup(local.master_instance_class, "runtimeOptions", {})
  calculated_memory_reservation = lookup(local.runtime_options, "memoryReservation", 80)
}
