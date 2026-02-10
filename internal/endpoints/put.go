package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/hadlow/genomdb/internal/consensus"
)

func Put(s ServerInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Writes must go through the leader
		if !consensus.RequireLeader(s.GetRaft(), w, r) {
			return
		}

		// Get key from query parameter
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key parameter", http.StatusBadRequest)
			return
		}

		var body struct {
			Value string `json:"value"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := consensus.ApplySet(s.GetRaft(), key, body.Value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
