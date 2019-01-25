package types

type VaultConfig struct {
	Addr    string `json:"Addr"`
	Enabled bool   `json:"Enabled"`
}

type VaultLoginInfo struct {
	RequestID string    `json:"request_id"`
	LeaseID   string    `json:"lease_id"`
	Renewable bool      `json:"renewable"`
	Auth      VaultAuth `json:"auth"`
}

type VaultAuth struct {
	ClientToken string                 `json:"client_token"`
	Accessor    string                 `json:"accessor"`
	Policies    []string               `json:"policies"`
	Metadata    map[string]interface{} `json:"metadata"`
}
