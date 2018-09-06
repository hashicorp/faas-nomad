package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hashicorp/faas-nomad/metrics"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/openfaas/faas-provider/types"
)

const nomadIdentifier = "nomad"

// MakeInfo creates handler for /system/info endpoint
func MakeInfo(logger hclog.Logger, stats metrics.StatsD, version string) http.HandlerFunc {
	log := logger.Named("info_handler")

	return func(rw http.ResponseWriter, r *http.Request) {
		stats.Incr("info.called", nil, 1)
		log.Info("Querying info")

		if r.Body != nil {
			defer r.Body.Close()
		}

		infoRequest := types.InfoRequest{
			Orchestration: nomadIdentifier,
			Version: types.ProviderVersion{
				Release: version,
				SHA:     "", // Note: this is an optional field.
			},
		}

		jsonOut, marshalErr := json.Marshal(infoRequest)
		if marshalErr != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			log.Warn("Unable to marshal system info", marshalErr.Error())
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(jsonOut)
	}
}
