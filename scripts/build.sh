#!/bin/bash
set -e

echo "Building pm..."
mkdir -p bin
go build -o bin/pm ./cmd/pm

echo "✓ Built binary: bin/pm"
echo ""
echo "To use:"
echo "  ./bin/pm init sandbox"
echo "  ./bin/pm --help"
