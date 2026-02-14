#!/bin/bash
docker-compose -f .devcontainer/docker-compose.yml down
echo "✓ Dev container stopped. Auth is preserved in volumes."
