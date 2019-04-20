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
  --data '{"policies": "openfaas", "period": "5m"}' \
  ${VAULT_URL}/v1/auth/approle/role/${POLICY_NAME}

curl -i \
  --header "X-Vault-Token: ${TOKEN}" \
  ${VAULT_URL}/v1/auth/approle/role/${POLICY_NAME}/role-id -o ./role_id.json

curl -i \
  --header "X-Vault-Token: ${TOKEN}" \
  --request POST \
  ${VAULT_URL}/v1/auth/approle/role/${POLICY_NAME}/secret-id -o ./secret_id.json

echo 'enabling pki backend...'
curl -i -H "X-Vault-Token: ${TOKEN}" -H "Content-Type: application/json" \
  -XPOST -d '{"type":"pki"}' ${VAULT_URL}/v1/sys/mounts/pki

echo 'generate root ca...'
curl -i -H "X-Vault-Token: ${TOKEN}" -H "Content-Type: application/json" \
  -XPOST -d '{"common_name":"nomad.local", "ip_sans": "192.168.50.2"}' ${VAULT_URL}/v1/pki/root/generate/internal

echo 'configure issuing urls...'
curl -i -H "X-Vault-Token: ${TOKEN}" -H "Content-Type: application/json" \
  -XPOST -d '{"issuing_certificates": ["http://localhost:8200/v1/pki/ca"], "crl_distribution_points": ["http://localhost:8200/v1/pki/crl"]}' ${VAULT_URL}/v1/pki/config/urls

echo 'create role...'
curl -i -H "X-Vault-Token: ${TOKEN}" -H "Content-Type: application/json" \
  -XPOST -d '{"allowed_domains": ["nomad.local"], "allow_subdomains": true, "max_ttl": "72h"}' ${VAULT_URL}/v1/pki/roles/faas-nomad

echo 'get certficates...'
curl -H "X-Vault-Token: ${TOKEN}" -H "Content-Type: application/json" \
  -XPOST -d '{"common_name": "server.nomad.local", "ip_sans": "192.168.50.2"}' ${VAULT_URL}/v1/pki/issue/faas-nomad -o ./output.json

apt-get install jq -y

jq -r '.data.issuing_ca' < ./output.json > ./placeholder-ca.crt
jq -r '.data.certificate' < ./output.json > ./placeholder.crt
jq -r '.data.private_key' < ./output.json > ./placeholder.key