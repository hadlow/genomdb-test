package types

/*
Exmaple item
{
	Id: "123",
	Hash: "abc",
	Size: 4092,
	Chunks: [
		{
			Id: "456",
			Nodes: [
				"127.0.0.1:8001"
				"127.0.0.1:8003"
			]
		}
	]
}
*/

type Item struct {
	Id     string
	Hash   string
	Size   int
	Chunks []Chunk
}

type Chunk struct {
	Id    string
	Nodes []string
}
