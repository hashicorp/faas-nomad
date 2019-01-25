package vault

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/faas-nomad/types"
	"github.com/hashicorp/vault/api"
)

func NewVaultLoginInfo(vaultClient *api.Client, vaultAppRoleID string, vaultAppRoleSecret string) (types.VaultLoginInfo, error) {

	var vaultLogin types.VaultLoginInfo
	loginRequest := vaultClient.NewRequest("POST", "/v1/auth/approle/login")
	loginRequest.SetJSONBody(map[string]interface{}{"role_id": vaultAppRoleID, "secret_id": vaultAppRoleSecret})
	req, err := loginRequest.ToHTTP()
	if err != nil {
		return vaultLogin, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return vaultLogin, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	parseErr := json.Unmarshal(body, &vaultLogin)
	if parseErr != nil {
		return vaultLogin, parseErr
	}

	return vaultLogin, nil
}
