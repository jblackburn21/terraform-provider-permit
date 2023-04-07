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

resource "permit_resource" "sample" {
  key            = "sample_resource"
  project_id     = permit_project.sample.id
  environment_id = permit_environment.sample.id
  name           = "Sample Resource"
  description    = "Terraform provider sample resource"
}

resource "permit_resource_action" "sample_create" {
  key            = "create"
  project_id     = permit_project.sample.id
  environment_id = permit_environment.sample.id
  resource_id    = permit_resource.sample.id
  name           = "create"
  description    = "Terraform provider sample resource action create"
}

resource "permit_role" "sample" {
  key            = "sample_role"
  project_id     = permit_project.sample.id
  environment_id = permit_environment.sample.id
  name           = "Sample Role"
  description    = "Terraform provider sample role"
  permissions = [
    permit_resource_action.sample_create.id
  ]
}

output "permit_project_sample_id" {
  value = permit_project.sample.id
}

output "permit_environment_sample_id" {
  value = permit_environment.sample.id
}

output "permit_resource_sample_id" {
  value = permit_resource.sample.id
}

output "permit_role_sample_id" {
  value = permit_role.sample.id
}