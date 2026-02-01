#!/bin/bash
set -e

# Build the binary
./scripts/build.sh

# Create a clean test environment
echo ""
echo "Setting up test environment in sandbox/..."
rm -rf sandbox
mkdir -p sandbox

cd sandbox

# Initialize and test
echo ""
echo "Running: pm init ."
../bin/pm init .

echo ""
echo "Running: pm new 'Test ticket'"
../bin/pm new "Test ticket"

echo ""
echo "Running: pm list"
../bin/pm list

echo ""
echo "✓ Local test complete"
echo ""
echo "Test directory: sandbox/"
echo "To continue testing:"
echo "  cd sandbox"
echo "  ../bin/pm <command>"
