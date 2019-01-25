#!/bin/bash

echo 'Waiting for vault...'
while true
do
  START=`docker logs dev-vault 2>&1 | grep "post-unseal setup complete"`
  if [ -n "$START" ]; then
    break
  else
    sleep 2
  fi
done

POLICY_NAME=openfaas
TOKEN=vagrant
VAULT_URL=http://127.0.0.1:8200

export VAULT_ADDR=${VAULT_URL}
export VAULT_TOKEN=${TOKEN}

vault auth enable approle

vault policy write ${POLICY_NAME} /vagrant/provisioning/scripts/policy.hcl

# create approle openfaas
curl -i \
  --header "X-Vault-Token: ${TOKEN}" \
  --request POST \
  --data '{"policies": "openfaas"}' \
  ${VAULT_URL}/v1/auth/approle/role/${POLICY_NAME}

curl -i \
  --header "X-Vault-Token: ${TOKEN}" \
  ${VAULT_URL}/v1/auth/approle/role/${POLICY_NAME}/role-id

curl -i \
  --header "X-Vault-Token: ${TOKEN}" \
  --request POST \
  ${VAULT_URL}/v1/auth/approle/role/${POLICY_NAME}/secret-id
