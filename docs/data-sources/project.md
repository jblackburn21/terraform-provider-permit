---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "permit_project Data Source - terraform-provider-permit"
subcategory: ""
description: |-
  Project data source
---

# permit_project (Data Source)

Project data source

## Example Usage

```terraform
data "permit_project" "project" {
  key = "tf-example"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `key` (String) Project key

### Read-Only

- `description` (String) Project description
- `id` (String) Project identifier
- `name` (String) Project name
- `organization_id` (String) Organization identifier
