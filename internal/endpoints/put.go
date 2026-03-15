package endpoints

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hadlow/genomdb/internal/consensus"
	"github.com/hadlow/genomdb/internal/helpers"
	"github.com/hadlow/genomdb/types"
)

func getHash(contents string) string {
	hasher := fnv.New32a()
	hasher.Write([]byte(contents))

	return strconv.FormatUint(uint64(hasher.Sum32()), 10)
}

func getChunkHash(chunk string) uint32 {
	hasher := fnv.New32a()
	hasher.Write([]byte(chunk))

	return hasher.Sum32()
}

func getNodeFromChunk(nodes []types.RaftNodeAddress, hash uint32) types.RaftNodeAddress {
	index := hash % uint32(len(nodes))
	log.Println("=== Chunk hash:", hash, " ", len(nodes), "assigned to node index:", index, "node address:", nodes[index].RaftAddress)

	return nodes[index]
}

func postChunk(url string, chunk string, hash uint32) (*http.Response, error) {
	body := strings.NewReader(chunk)
	contentType := "text/plain"

	fmt.Printf("Posting chunk with hash %d to node at %s\n", hash, url)

	response, err := http.Post(url+"/write-chunk?hash="+strconv.FormatUint(uint64(hash), 10), contentType, body)
	if err != nil {
		log.Fatalf("Error making POST request: %v", err)
		return nil, err
	}

	defer response.Body.Close()
	return response, nil
}

func Put(s ServerInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// cfg := s.GetConfig()
		raft := s.GetRaft()

		// Writes must go through the leader
		if !consensus.RequireLeader(raft, w, r) {
			return
		}

		// Get key from query parameter
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key parameter", http.StatusBadRequest)
			return
		}

		var body struct {
			Contents string `json:"contents"`
		}

		// Check body decode
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Transform file contents into chunks with metadata
		chunks := helpers.ChunkData(body.Contents, 5)
		chunkData := []types.Chunk{}

		// Get network nodes
		nodes, err := helpers.GetRaftNodeAddresses(raft)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(nodes) == 0 {
			http.Error(w, "no raft nodes available", http.StatusServiceUnavailable)
			return
		}

		// Post chunks to nodes
		for i := range chunks {
			hash := getChunkHash(chunks[i])
			node := getNodeFromChunk(nodes, hash)

			httpAddr, err := helpers.RaftToHttpAddress(node.RaftAddress)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = postChunk("http://"+httpAddr, chunks[i], hash)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			chunkData = append(chunkData, types.Chunk{
				Id:    strconv.FormatUint(uint64(hash), 10),
				Nodes: []string{httpAddr},
			})
		}

		// Once all chunks are posted, store metadata in the FSM
		item := helpers.GetMetadata(key, getHash(body.Contents), len(body.Contents), chunkData)
		metadata := helpers.StringifyMetadata(item)

		if err := consensus.ApplySet(s.GetRaft(), key, metadata); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
