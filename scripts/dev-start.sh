#!/bin/bash
docker-compose -f .devcontainer/docker-compose.yml up -d
echo "✓ Dev container started. Use './scripts/dev-copilot.sh' to interact with Copilot CLI."
echo "  - Use './scripts/dev-copilot.sh' to interact with Copilot CLI."
echo "  - Use './scripts/dev-shell.sh' to enter container shell."
echo "  - Use './scripts/dev-stop.sh' to stop container (preserves auth in volumes)."
