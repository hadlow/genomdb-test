package helpers

import (
	"strconv"
	"strings"

	"github.com/hadlow/genomdb/types"
)

func GetMetadata(key string, hash string, length int, chunkData []types.Chunk) types.Item {
	return types.Item{
		Id:     key,
		Hash:   hash,
		Size:   length,
		Chunks: chunkData,
	}
}

/*
We know where the chunks start and end as it will always
be 1 (the ID) + the number of nodes data is replicated to.

CHUNK: 345,127.0.0.1:8001,127.0.0.1:8003
123,qwe,1000,<CHUNK>,<CHUNK>
*/
func StringifyMetadata(item types.Item) string {
	chunks := ""

	for i := range item.Chunks {
		nodes := strings.Join(item.Chunks[i].Nodes, ",")
		chunks += item.Chunks[i].Id + "," + nodes
	}

	return item.Id + "," + item.Hash + "," + strconv.Itoa(item.Size) + "," + chunks
}

func ChunkData(contents string, chunkLines int) []string {
	var chunks []string

	// Split contents into lines
	lines := strings.Split(contents, "\n")

	// Create chunks of specified line count
	for i := 0; i < len(lines); i += chunkLines {
		end := i + chunkLines
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, strings.Join(lines[i:end], "\n"))
	}

	return chunks
}
