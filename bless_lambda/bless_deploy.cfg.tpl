[Bless Options]
certificate_validity_after_seconds = 3600
certificate_validity_before_seconds = 3600
entropy_minimum_bits = 2048
random_seed_bytes = 256
logging_level = INFO
username_validation = email

[Bless CA]
${region}_password = ${encrypted_password}
ca_private_key = ${encrypted_ca}

[KMS Auth]
use_kmsauth = True
kmsauth_key_id = ${kms_auth_key_id}
kmsauth_serviceid = ${name}
kmsauth_remote_usernames_allowed = *
