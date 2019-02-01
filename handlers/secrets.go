package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/faas-nomad/vault"

	"github.com/hashicorp/faas-nomad/types"
	hclog "github.com/hashicorp/go-hclog"
	vapi "github.com/hashicorp/vault/api"
	"github.com/openfaas/faas/gateway/requests"
)

func MakeSecretHandler(vs *vault.VaultService, log hclog.Logger, providerConfig types.ProviderConfig) http.HandlerFunc {
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
			responseStatus int
			responseBody   []byte
			responseErr    error
		)

		switch r.Method {
		case http.MethodGet:
			responseStatus, responseBody, responseErr = getSecrets(vs, providerConfig, body)
			break
		case http.MethodPost:
			responseStatus, responseBody, responseErr = createNewSecret(http.MethodPost, vs, providerConfig, body)
			break
		case http.MethodPut:
			responseStatus, responseBody, responseErr = createNewSecret(http.MethodPut, vs, providerConfig, body)
			break
		case http.MethodDelete:
			responseStatus, responseBody, responseErr = deleteSecret(vs, providerConfig, body)
			break
		}

		if responseErr != nil {
			log.Error(responseErr.Error())
			w.WriteHeader(responseStatus)
			return
		}

		w.WriteHeader(responseStatus)

		if responseBody != nil {
			_, writeErr := w.Write(responseBody)

			if writeErr != nil {
				log.Error("Cannot write body of a response")
				w.WriteHeader(http.StatusInternalServerError)

				return
			}
		}
	}
}

func getSecrets(vs *vault.VaultService, providerConfig types.ProviderConfig, body []byte) (responseStatus int, responseBody []byte, err error) {

	response, respErr := vs.DoRequest("LIST",
		fmt.Sprintf("/v1/secret/%s", providerConfig.Vault.DefaultPolicy), nil)

	if respErr != nil {
		return http.StatusInternalServerError,
			nil,
			fmt.Errorf("Error in request to Vault: %s", respErr)
	}

	// If Vault finds nothing, return StatusOK according to gateway API docs
	if response.StatusCode == http.StatusNotFound {
		return http.StatusOK, []byte(`[]`), nil
	}

	var secretList vapi.Secret
	secretsBody, bodyErr := ioutil.ReadAll(response.Body)
	if bodyErr != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("Error reading response body: %s", bodyErr)
	}

	unmarshalErr := json.Unmarshal(secretsBody, &secretList)
	if unmarshalErr != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("Error in json deserialisation: %s", unmarshalErr)
	}

	secrets := []requests.Secret{}
	for _, k := range secretList.Data["keys"].([]interface{}) {
		secrets = append(secrets, requests.Secret{Name: k.(string)})
	}

	resultsJson, marshalErr := json.Marshal(secrets)
	if marshalErr != nil {
		return http.StatusInternalServerError,
			nil,
			marshalErr
	}

	return http.StatusOK, resultsJson, nil
}

func createNewSecret(method string, vs *vault.VaultService, providerConfig types.ProviderConfig, body []byte) (responseStatus int, responseBody []byte, err error) {

	var secret requests.Secret
	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return http.StatusBadRequest, nil, fmt.Errorf("Error in request json deserialisation: %s", unmarshalErr)
	}

	response, respErr := vs.DoRequest(method,
		fmt.Sprintf("/v1/secret/%s/%s", providerConfig.Vault.DefaultPolicy, secret.Name),
		map[string]interface{}{"value": secret.Value})

	if respErr != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("Error in request to Vault: %s", respErr)
	}

	// Vault only returns 204 type success
	if response.StatusCode != http.StatusNoContent {
		return http.StatusBadRequest, nil, fmt.Errorf("Vault returned unexpected response: %v", response.StatusCode)
	}

	// as per gateway api docs
	if method == http.MethodPost {
		return http.StatusCreated, nil, nil
	} else {
		return http.StatusOK, nil, nil
	}
}

func deleteSecret(vs *vault.VaultService, providerConfig types.ProviderConfig, body []byte) (responseStatus int, responseBody []byte, err error) {

	var secret requests.Secret
	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return http.StatusBadRequest, nil, fmt.Errorf("Error in request json deserialisation: %s", unmarshalErr)
	}

	response, respErr := vs.DoRequest(http.MethodDelete,
		fmt.Sprintf("/v1/secret/%s/%s", providerConfig.Vault.DefaultPolicy, secret.Name), nil)
	if respErr != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("Error in request to Vault: %s", respErr)
	}

	if response.StatusCode != http.StatusNoContent {
		return http.StatusBadRequest, nil, fmt.Errorf("Vault returned unexpected response: %v", response.StatusCode)
	}

	return http.StatusOK, nil, nil

}
