package types

type VaultConfig struct {
	Addr             string `json:"Addr"`
	Enabled          bool   `json:"Enabled"`
	DefaultPolicy    string
	SecretPathPrefix string
	AppRoleID        string
	AppSecretID      string
}
