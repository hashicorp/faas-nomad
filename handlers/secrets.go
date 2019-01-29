package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/faas-nomad/types"
	hclog "github.com/hashicorp/go-hclog"
	vapi "github.com/hashicorp/vault/api"
	"github.com/openfaas/faas/gateway/requests"
)

func MakeSecretHandler(vaultClient *vapi.Client, log hclog.Logger, providerConfig types.ProviderConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		body, readBodyErr := ioutil.ReadAll(r.Body)
		if readBodyErr != nil {
			log.Error("couldn't read body of a request: %s", readBodyErr)

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
			responseStatus, responseBody, responseErr = getSecrets(vaultClient, providerConfig, body)
			break
		case http.MethodPost:
			responseStatus, responseBody, responseErr = createNewSecret(vaultClient, providerConfig, body)
			break
		case http.MethodPut:
			responseStatus, responseBody, responseErr = createNewSecret(vaultClient, providerConfig, body)
			break
		case http.MethodDelete:
			responseStatus, responseBody, responseErr = deleteSecret(vaultClient, providerConfig, body)
			break
		}

		if responseErr != nil {
			log.Error("Vault error response", responseErr)

			w.WriteHeader(responseStatus)

			return
		}

		if responseBody != nil {
			_, writeErr := w.Write(responseBody)

			if writeErr != nil {
				log.Error("Cannot write body of a response")
				w.WriteHeader(http.StatusInternalServerError)

				return
			}
		}

		w.WriteHeader(responseStatus)
	}
}

func getSecrets(vaultClient *vapi.Client, providerConfig types.ProviderConfig, body []byte) (responseStatus int, responseBody []byte, err error) {

	response, err := getSecretResponse(vaultClient, "LIST", fmt.Sprintf("/v1/secret/%s", providerConfig.Vault.DefaultPolicy), nil)

	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	if response.StatusCode != http.StatusOK {
		return response.StatusCode, nil, err
	}

	var secretList vapi.Secret
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	unmarshalErr := json.Unmarshal(body, &secretList)
	if unmarshalErr != nil {
		return http.StatusInternalServerError, nil, err
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

func createNewSecret(vaultClient *vapi.Client, providerConfig types.ProviderConfig, body []byte) (responseStatus int, responseBody []byte, err error) {

	var secret requests.Secret

	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return http.StatusBadRequest, nil, unmarshalErr
	}

	response, err := getSecretResponse(vaultClient, http.MethodPost, fmt.Sprintf("/v1/secret/%s/%s", providerConfig.Vault.DefaultPolicy, secret.Name), map[string]interface{}{"value": secret.Value})
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	if response.StatusCode != http.StatusNoContent {
		return response.StatusCode, nil, err
	}

	return http.StatusCreated, nil, nil
}

func deleteSecret(vaultClient *vapi.Client, providerConfig types.ProviderConfig, body []byte) (responseStatus int, responseBody []byte, err error) {

	var secret requests.Secret

	unmarshalErr := json.Unmarshal(body, &secret)
	if unmarshalErr != nil {
		return http.StatusBadRequest, nil, unmarshalErr
	}

	response, err := getSecretResponse(vaultClient, http.MethodDelete, fmt.Sprintf("/v1/secret/%s/%s", providerConfig.Vault.DefaultPolicy, secret.Name), nil)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	if response.StatusCode != http.StatusNoContent {
		return response.StatusCode, nil, err
	}

	return http.StatusOK, nil, nil

}

func getSecretResponse(vaultClient *vapi.Client, method string, path string, body interface{}) (*http.Response, error) {

	client := &http.Client{}
	createRequest := vaultClient.NewRequest(method, path)
	createRequest.SetJSONBody(body)

	request, _ := createRequest.ToHTTP()
	return client.Do(request)
}
