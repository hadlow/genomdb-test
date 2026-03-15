# GenomDB

GenomDB is a distributed storage system for genomic data. GenomDB can store both SAM and BAM files.

## How it works

Each read of the SAM file will be split across multiple nodes in the network.

## GenomDB Query language (GQL)

GQL will be translated by the index in order to figure out the get calls that need to be made to the data.

## Features

- **Raft Consensus**: Ensures all nodes have consistent data
- **Distributed Storage**: Run multiple nodes locally or across networks
- **HTTP API**: Simple REST API for get/put operations
- **Automatic Leader Election**: Raft handles leader election and failover

## Quick Start

### Using Docker (Recommended)

The easiest way to run the cluster is using Docker Compose:

1. **Build and start all nodes:**
   ```bash
   docker-compose up -d
   ```

2. **Initialize the cluster (add nodes 2 and 3):**
   ```bash
   ./docker-init.sh
   ```
   
   Or manually:
   ```bash
   # Add node2
   curl -X POST http://127.0.0.1:8001/join \
     -H "Content-Type: application/json" \
       -d '{"node_id": "node2", "node_addr": "node2:9026"}'
   
   # Add node3
   curl -X POST http://127.0.0.1:8001/join \
     -H "Content-Type: application/json" \
       -d '{"node_id": "node3", "node_addr": "node3:9027"}'
   ```

3. **View logs:**
   ```bash
   docker-compose logs -f
   ```

4. **Open the monitoring dashboard:**
   - http://127.0.0.1:8080

5. **Stop the cluster:**
   ```bash
   docker-compose down
   ```

6. **Stop and remove volumes (clean slate):**
   ```bash
   docker-compose down -v
   ```

### Docker Hot Reload (Development)

To rebuild/restart node processes automatically when `.go` files change, run:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

Or with Make:

```bash
make docker-dev-up
```

Detached mode:

```bash
make docker-dev-up-d
```

`make docker-dev-up-d` now starts containers and runs cluster init (`/join` for node2/node3).

If Raft gets stuck after address/config changes (for example heartbeats to `127.0.0.1:9025` inside containers), reset dev volumes and re-init:

```bash
make docker-dev-reset
```

This uses `air` inside each node container and bind-mounts your workspace, so saving Go files triggers recompilation and process restart.

Stop hot-reload stack:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml down
```

## API Endpoints

### PUT - Store a value
```bash
curl -X PUT "http://127.0.0.1:8001/put?key=mykey" \
  -H "Content-Type: application/json" \
  -d '{"value": "myvalue"}'
```
**Note:** Writes must go through the leader. If you hit a follower, it will redirect you to the leader.

### GET - Retrieve a value
```bash
curl "http://127.0.0.1:8001/get?key=mykey"
```
**Note:** Reads can be performed on any node (leader or follower).

### PING - Health check
```bash
curl http://127.0.0.1:8001/ping
```

### JOIN - Add a node to the cluster
```bash
curl -X POST http://127.0.0.1:8001/join \
  -H "Content-Type: application/json" \
   -d '{"node_id": "node2", "node_addr": "127.0.0.1:9026"}'
```
**Note:** Only the leader can add nodes.

### STATUS - Cluster monitoring metadata
```bash
curl http://127.0.0.1:8001/status
```
Returns node information including raft state, leader, peer list, keys, and in-memory store values.

## Configuration

Each node requires a configuration file with:

- `database`: Path to the BoltDB database file
- `server.host`: HTTP server host
- `server.port`: HTTP server port
- `raft.node_id`: Unique node identifier
- `raft.bind_addr`: Raft bind host (Raft port is derived as `server.port + 1024`)
- `raft.advertise_addr`: Optional Raft advertise host (port also derived as `server.port + 1024`)
- `raft.data_dir`: Directory for Raft logs and snapshots
- `raft.peers`: List of peer Raft addresses (empty for first node)

Example config files are provided in the `configs/` directory:
- `config-node1.yml` / `config-node1-docker.yml` - First node (bootstraps cluster)
- `config-node2.yml` / `config-node2-docker.yml` - Second node
- `config-node3.yml` / `config-node3-docker.yml` - Third node

**Note:** The `-docker.yml` variants are configured for Docker networking (using service names instead of localhost).

## How It Works

1. **First Node**: Bootstraps the Raft cluster
2. **Additional Nodes**: Start and wait to be added to the cluster
3. **Leader Election**: Raft automatically elects a leader
4. **Consensus**: All writes go through the leader and are replicated to followers
5. **Reads**: Can be performed on any node (eventually consistent)

## Data Persistence

- Raft logs and snapshots are stored in `data/<node_id>/`
- Key-value data is stored in the FSM (in-memory, persisted via snapshots)
- BoltDB database files store additional metadata
