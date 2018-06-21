provider "bless" {
  region  = "us-east-1"
  profile = "czi-si"
}

resource "bless_ca" "ca" {
  kms_key_id = "800cfa14-4ef4-49ca-af07-4722908ae3a0"
}

output "priv" {
  value = "${bless_ca.ca.encrypted_ca}"
}

output "pub" {
  value = "${bless_ca.ca.public_key}"
}

output "password" {
  value = "${bless_ca.ca.encrypted_password}"
}
