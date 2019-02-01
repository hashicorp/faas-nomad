package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/faas-nomad/vault"
	hclog "github.com/hashicorp/go-hclog"
	vapi "github.com/hashicorp/vault/api"
	"github.com/openfaas/faas/gateway/requests"
)

type SecretsResponse struct {
	StatusCode int
	Body       []byte
}

func MakeSecretHandler(vs *vault.VaultService, log hclog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		body, readBodyErr := ioutil.ReadAll(r.Body)
		if readBodyErr != nil {
			log.Error("Couldn't read body of a request: %s", readBodyErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var (
			response    SecretsResponse
			responseErr error
		)

		switch r.Method {
		case http.MethodGet:
			response, responseErr = getSecrets(vs, body)
			break
		case http.MethodPost:
			response, responseErr = createNewSecret(vs, body)
			break
		case http.MethodPut:
			response, responseErr = updateSecret(vs, body)
			break
		case http.MethodDelete:
			response, responseErr = deleteSecret(vs, body)
			break
		}

		if responseErr != nil {
			log.Error(responseErr.Error())
			w.WriteHeader(response.StatusCode)
			return
		}

		w.WriteHeader(response.StatusCode)

		if response.Body != nil {
			_, writeErr := w.Write(response.Body)

			if writeErr != nil {
				log.Error("Cannot write body of a response")
				w.WriteHeader(http.StatusInternalServerError)

				return
			}
		}
	}
}

func getSecrets(vs *vault.VaultService, body []byte) (resp SecretsResponse, err error) {

	response, respErr := vs.DoRequest("LIST",
		fmt.Sprintf("/v1/secret/%s", vs.Config.DefaultPolicy), nil)

	if respErr != nil {
		return SecretsResponse{StatusCode: http.StatusInternalServerError},
			fmt.Errorf("Error in request to Vault: %s", respErr)
	}

	// If Vault finds nothing, return StatusOK according to gateway API docs
	if response.StatusCode == http.StatusNotFound {
		return SecretsResponse{
			StatusCode: http.StatusOK,
			Body:       []byte(`[]`),
		}, nil
	}

	var secretList vapi.Secret
	secretsBody, bodyErr := ioutil.ReadAll(response.Body)
	if bodyErr != nil {
		return SecretsResponse{StatusCode: http.StatusInternalServerError},
			fmt.Errorf("Error reading response body: %s", bodyErr)
	}

	unmarshalErr := json.Unmarshal(secretsBody, &secretList)
	if unmarshalErr != nil {
		return SecretsResponse{StatusCode: http.StatusInternalServerError},
			fmt.Errorf("Error in json deserialisation: %s", unmarshalErr)
	}

	secrets := []requests.Secret{}
	for _, k := range secretList.Data["keys"].([]interface{}) {
		secrets = append(secrets, requests.Secret{Name: k.(string)})
	}

	resultsJson, marshalErr := json.Marshal(secrets)
	if marshalErr != nil {
		return SecretsResponse{StatusCode: http.StatusInternalServerError},
			marshalErr
	}

	return SecretsResponse{StatusCode: http.StatusOK, Body: resultsJson}, nil
}

func createNewSecret(vs *vault.VaultService, body []byte) (resp SecretsResponse, err error) {

	var secret requests.Secret
	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return SecretsResponse{StatusCode: http.StatusBadRequest}, fmt.Errorf("Error in request json deserialisation: %s", unmarshalErr)
	}

	response, respErr := vs.DoRequest(http.MethodPost,
		fmt.Sprintf("/v1/secret/%s/%s", vs.Config.DefaultPolicy, secret.Name),
		map[string]interface{}{"value": secret.Value})

	if respErr != nil {
		return SecretsResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("Error in request to Vault: %s", respErr)
	}

	// Vault only returns 204 type success
	if response.StatusCode != http.StatusNoContent {
		return SecretsResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("Vault returned unexpected response: %v", response.StatusCode)
	}

	return SecretsResponse{StatusCode: http.StatusCreated}, nil
}

func updateSecret(vs *vault.VaultService, body []byte) (resp SecretsResponse, err error) {

	var secret requests.Secret
	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return SecretsResponse{StatusCode: http.StatusBadRequest}, fmt.Errorf("Error in request json deserialisation: %s", unmarshalErr)
	}

	response, respErr := vs.DoRequest(http.MethodPut,
		fmt.Sprintf("/v1/secret/%s/%s", vs.Config.DefaultPolicy, secret.Name),
		map[string]interface{}{"value": secret.Value})

	if respErr != nil {
		return SecretsResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("Error in request to Vault: %s", respErr)
	}

	// Vault only returns 204 type success
	if response.StatusCode != http.StatusNoContent {
		return SecretsResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("Vault returned unexpected response: %v", response.StatusCode)
	}

	return SecretsResponse{StatusCode: http.StatusOK}, nil
}

func deleteSecret(vs *vault.VaultService, body []byte) (resp SecretsResponse, err error) {

	var secret requests.Secret
	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return SecretsResponse{StatusCode: http.StatusBadRequest}, fmt.Errorf("Error in request json deserialisation: %s", unmarshalErr)
	}

	response, respErr := vs.DoRequest(http.MethodDelete,
		fmt.Sprintf("/v1/secret/%s/%s", vs.Config.DefaultPolicy, secret.Name), nil)
	if respErr != nil {
		return SecretsResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("Error in request to Vault: %s", respErr)
	}

	if response.StatusCode != http.StatusNoContent {
		return SecretsResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("Vault returned unexpected response: %v", response.StatusCode)
	}

	return SecretsResponse{StatusCode: http.StatusOK}, nil
}
