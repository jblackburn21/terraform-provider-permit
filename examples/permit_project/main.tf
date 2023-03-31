terraform {
  required_providers {
    permit = {
      source = "jblackburn21/permit"
    }
  }
}

provider "permit" {}

data "permit_project" "abac_sample" {
  key = "abac-sample"
}

#resource "permit_project" "project" {
#  key         = "sample-project"
#  name        = "Sample Project"
#  description = "Terraform provider sample project"
#}


output "abac_sample_name" {
  value = data.permit_project.abac_sample.name
}

output "abac_sample_description" {
  value = data.permit_project.abac_sample.description
}