# Terraform-provider-bless
----

**Please note**: If you believe you have found a security issue, _please responsibly disclose_ by contacting us at [security@chanzuckerberg.com](mailto:security@chanzuckerberg.com).

----

Terraform provider to automate the creation of [BLESS](https://github.com/Netflix/bless) deployments.
This provider generates a BLESS CA without leaking any sensitive material to the terraform state store. The private part of the key is encrypted with a password. This password is then encrypted through KMS so that it is compatible with BLESS.

### Example usage

```hcl
provider "bless" {
  region  = "us-east-1"
  profile = "<aws_profile>"
}

resource "bless" "example" {
  kms_key_id = "<kms_key_id>"
}

# The encrypted CA private key
output "encrypted_ca" {
  value = "${bless.example.encrypted_ca}"
}

# The CA public key
output "ca" {
  value = "${bless.example.public_key}"
}

# The KMS encrypted CA password
output "password" {
  value = "${bless.example.encrypted_password}"
}

```
This module only creates logical resources and therefore only contributes to terraform state. Does not create externally managed resources. In order to generate a new key then, you must taint the resource. Terraform will then generate a new key on the next run.

```sh
terraform taint bless.example
```
