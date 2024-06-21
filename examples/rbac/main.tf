terraform {
  required_providers {
    permit = {
      source = "jblackburn21/permit"
    }
  }
}

provider "permit" {}

resource "permit_project" "sample" {
  key         = "sample-project"
  name        = "Sample Project"
  description = "Terraform provider sample project"
}

resource "permit_environment" "sample" {
  key         = "sample-environment"
  project_id  = permit_project.sample.id
  name        = "Sample Environment"
  description = "Terraform provider sample environment"
}

resource "permit_tenant" "sample" {
  key            = "sample-tenant"
  project_id     = permit_project.sample.id
  environment_id = permit_environment.sample.id
  name           = "Sample Tenant"
  description    = "Terraform provider sample tenant"
}

output "permit_project_sample_id" {
  value = permit_project.sample.id
}

output "permit_environment_sample_id" {
  value = permit_environment.sample.id
}

output "permit_tenant_sample_id" {
  value = permit_tenant.sample.id
}
