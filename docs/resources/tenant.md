---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "permit_tenant Resource - terraform-provider-permit"
subcategory: ""
description: |-
  Tenant resource
---

# permit_tenant (Resource)

Tenant resource

## Example Usage

```terraform
resource "permit_tenant" "sample" {
  key            = "sample_tenant"
  project_id     = "405d8375-3514-403b-8c43-83ae74cfe0e9"
  environment_id = "40ef0e48-a11f-4963-a229-e396c9f7e7c4"
  name           = "Sample Tenant"
  description    = "Terraform provider sample tenant"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment_id` (String) Environment identifier
- `key` (String) Tenant key
- `name` (String) Tenant name
- `project_id` (String) Project identifier

### Optional

- `description` (String) Tenant description

### Read-Only

- `id` (String) Tenant identifier
- `organization_id` (String) Organization identifier

