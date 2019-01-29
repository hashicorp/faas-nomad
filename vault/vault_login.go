package vault

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/vault/api"
)

func NewVaultLoginInfo(vaultClient *api.Client, roleID string, secretID string) (api.Secret, error) {

	var vaultLogin api.Secret
	client := &http.Client{}

	loginRequest := vaultClient.NewRequest("POST", "/v1/auth/approle/login")
	loginRequest.SetJSONBody(map[string]interface{}{"role_id": roleID, "secret_id": secretID})
	lReq, err := loginRequest.ToHTTP()
	if err != nil {
		return vaultLogin, err
	}

	lResp, err := client.Do(lReq)
	if err != nil {
		return vaultLogin, err
	}

	if lResp.StatusCode != http.StatusOK {
		return vaultLogin, fmt.Errorf("Vault response status code %v", lResp.StatusCode)
	}

	lBody, _ := ioutil.ReadAll(lResp.Body)
	parseErr := json.Unmarshal(lBody, &vaultLogin)
	if parseErr != nil {
		return vaultLogin, parseErr
	}

	return vaultLogin, nil
}
