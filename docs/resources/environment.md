---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "permit_environment Resource - terraform-provider-permit"
subcategory: ""
description: |-
  Environment resource
---

# permit_environment (Resource)

Environment resource

## Example Usage

```terraform
resource "permit_project" "sample" {
  key         = "sample-environment"
  project_id  = "405d8375-3514-403b-8c43-83ae74cfe0e9"
  name        = "Sample Environment"
  description = "Terraform provider sample environment"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `key` (String) Environment key
- `name` (String) Environment name
- `project_id` (String) Project identifier

### Optional

- `description` (String) Environment description

### Read-Only

- `id` (String) Environment identifier
- `organization_id` (String) Organization identifier

