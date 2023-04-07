resource "permit_user_attribute" "sample" {
  key            = "sample_user_attribute"
  project_id     = "405d8375-3514-403b-8c43-83ae74cfe0e9"
  environment_id = "40ef0e48-a11f-4963-a229-e396c9f7e7c4"
  type           = "string" // bool, number, string, time, array, json
  description    = "Terraform provider sample user attribute"
}