package endpoints

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hadlow/genomdb/internal/consensus"
)

// Writes a file chunk by key
func WriteChunk(s ServerInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("=== WRITING CHUNK")
		log.Println("=== WRITING CHUNK logger")
		// Writes must go through the leader
		if !consensus.RequireLeader(s.GetRaft(), w, r) {
			return
		}

		// Get hash from query parameter
		hash := r.URL.Query().Get("hash")
		if hash == "" {
			http.Error(w, "missing hash parameter", http.StatusBadRequest)
			return
		}

		// /*
		// 	Example body: (todo)
		// 	{
		// 		"chunks": [
		// 			{
		// 				"id": "chunk1",
		// 				"contents": "chunk1 contents"
		// 			},
		// 		]
		// 	}
		// */
		// var body struct {
		// 	Chunks []struct {
		// 		ID       string `json:"id"`
		// 		Contents string `json:"contents"`
		// 	} `json:"chunks"`
		// }

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("=== Hash: " + hash)
		fmt.Println(string(body))

		w.WriteHeader(http.StatusNoContent)
	}
}
