#!/bin/bash

# Script to initialize the Raft cluster after Docker containers start
# This adds nodes 2 and 3 to the cluster via the leader (node1)

echo "Waiting for node1 to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  if curl -s -f http://127.0.0.1:8001/ping > /dev/null 2>&1; then
    echo "Node1 is ready!"
    break
  fi
  RETRY_COUNT=$((RETRY_COUNT + 1))
  echo "Waiting for node1... ($RETRY_COUNT/$MAX_RETRIES)"
  sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
  echo "Error: Node1 did not become ready in time"
  exit 1
fi

sleep 2

echo "Adding node2 to cluster..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://127.0.0.1:8001/join \
  -H "Content-Type: application/json" \
  -d '{"node_id": "node2", "node_addr": "node2:9002"}')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "307" ]; then
  echo "✓ Node2 added successfully"
else
  echo "⚠ Failed to add node2 (HTTP $HTTP_CODE) - may already be in cluster"
fi

sleep 2

echo "Adding node3 to cluster..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://127.0.0.1:8001/join \
  -H "Content-Type: application/json" \
  -d '{"node_id": "node3", "node_addr": "node3:9003"}')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "307" ]; then
  echo "✓ Node3 added successfully"
else
  echo "⚠ Failed to add node3 (HTTP $HTTP_CODE) - may already be in cluster"
fi

echo ""
echo "Cluster initialization complete!"
echo ""
echo "Test the cluster:"
echo "  curl -X PUT 'http://127.0.0.1:8001/put?key=test' -H 'Content-Type: application/json' -d '{\"value\": \"hello\"}'"
echo "  curl 'http://127.0.0.1:8001/get?key=test'"
echo ""
echo "View logs: docker-compose logs -f"
