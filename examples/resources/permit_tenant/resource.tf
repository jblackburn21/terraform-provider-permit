resource "permit_tenant" "sample" {
  key            = "sample_tenant"
  project_id     = "405d8375-3514-403b-8c43-83ae74cfe0e9"
  environment_id = "40ef0e48-a11f-4963-a229-e396c9f7e7c4"
  name           = "Sample Tenant"
  description    = "Terraform provider sample tenant"
}