#!/bin/bash
# Build and run botTaskTracker
set -e

cd "$(dirname "$0")"

echo "Building CSS assets..."
npm run build:linux

echo "Generating templ files..."
/usr/local/go/bin/go generate ./...

echo "Starting botTaskTracker..."
exec /usr/local/go/bin/go run .
