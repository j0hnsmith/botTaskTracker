#!/bin/bash
# Script to rebuild, restart, test, and screenshot botTaskTracker

set -e

cd /home/openclaw/.openclaw/workspace/botTaskTracker

echo "==> Generating templ files..."
/usr/local/go/bin/go generate ./...

echo "==> Building application..."
/usr/local/go/bin/go build -o botTaskTracker ./cmd/server

echo "==> Stopping existing server (if running)..."
pkill -f 'botTaskTracker' || true
sleep 1

echo "==> Starting server in background..."
./botTaskTracker &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

echo "==> Waiting for server to start..."
sleep 3

echo "==> Checking server is responding..."
curl -s http://localhost:7002 > /dev/null && echo "✓ Server is up!" || echo "✗ Server failed to start"

echo "==> Taking screenshots..."
# Use Chrome to take screenshots
chromium-browser --headless --disable-gpu --screenshot=/home/openclaw/.openclaw/workspace/botTaskTracker/test-screenshots/full-page.png --window-size=1920,1080 http://localhost:7002

echo "==> Screenshot saved to test-screenshots/full-page.png"

echo ""
echo "Server is running on http://localhost:7002"
echo "To stop: kill $SERVER_PID"
