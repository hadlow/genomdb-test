package endpoints

import (
	"encoding/json"
	"net/http"
)

// Reads a file chunk by key
func ReadChunk(s ServerInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get key from query parameter
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key parameter", http.StatusBadRequest)
			return
		}

		// Read from FSM (can read from any node, not just leader)
		value, ok := s.GetFSM().Get(key)
		if !ok {
			http.NotFound(w, r)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"key":   key,
			"value": value,
		})
	}
}
