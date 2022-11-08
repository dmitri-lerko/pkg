provider "aws" {}

resource "random_pet" "suffix" {}

locals {
  name = "flux-test-${random_pet.suffix.id}"
}

module "eks" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/aws/eks"

  name = local.name
}

module "test_ecr" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/aws/ecr"

  name = "test-repo-${local.name}"
}

module "test_app_ecr" {
  source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/aws/ecr"

  name = "test-app-${local.name}"
}
