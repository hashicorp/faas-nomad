path "secret/openfaas/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Allow our own token to be renewed.
path "auth/token/renew-self" {
  capabilities = ["update"]
}