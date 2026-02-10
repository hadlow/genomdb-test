# Configuration Files

This directory contains configuration files for running GenomDB nodes.

## Local Development (Non-Docker)

- `config-node1.yml` - First node (bootstraps the Raft cluster)
- `config-node2.yml` - Second node
- `config-node3.yml` - Third node

These configs use `127.0.0.1` for local development.

## Docker

- `config-node1-docker.yml` - First node (bootstraps the Raft cluster)
- `config-node2-docker.yml` - Second node
- `config-node3-docker.yml` - Third node

These configs use Docker service names (`node1`, `node2`, `node3`) for inter-node communication and `0.0.0.0` to bind to all interfaces inside containers.

## Usage

### Local
```bash
./genomdb start configs/config-node1.yml
```

### Docker
```bash
docker-compose up -d
```
