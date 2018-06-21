Terraform provider to automate the creation of [BLESS](https://github.com/Netflix/bless) deployments.
This provider generates an RSA keypair. The private part of the key is encrypted with a password.
This password is then encrypted through KMS so that it is compatible with BLESS.

This module only creates a logical resources and therefore only contributes to terraform state.
Does not create externally managed resources.
