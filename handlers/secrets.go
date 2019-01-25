package handlers

import (
	"net/http"

	"github.com/hashicorp/faas-nomad/types"
	hclog "github.com/hashicorp/go-hclog"
	vapi "github.com/hashicorp/vault/api"
)

func MakeSecretHandler(vaultClient *vapi.Client, logger hclog.Logger, providerConfig types.ProviderConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
