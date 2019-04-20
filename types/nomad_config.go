package types

type NomadConfig struct {
	TLSEnabled    bool
	Address       string
	ACLToken      string
	TLSCA         string
	TLSCert       string
	TLSPrivateKey string
	TLSSkipVerify bool
}
