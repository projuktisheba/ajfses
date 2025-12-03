#!/bin/bash

# -----------------------------
# Configuration
# -----------------------------
VPS_HOST="203.161.48.179"
REMOTE_PATH="/home/samiul/apps/bin/ajfses-api"
SERVICE_NAME="ajfsesapi.service"
PING_URL="https://ajfses-api.pssoft.xyz/api/v1/ping"

# -----------------------------
# Step 1: Remove old binary locally
# -----------------------------
echo "Removing old binary..."
rm -f app

# -----------------------------
# Step 2: Build the Go app
# -----------------------------
echo "Building app for Linux..."

GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o app

if [[ $? -ne 0 ]]; then
    echo "Build failed. Exiting."
    exit 1
fi
echo "Build successful: app-linux"


# -----------------------------
# Step 3: Stop the service on VPS
# -----------------------------
echo "Stopping remote service..."
ssh samiul@"$VPS_HOST" "sudo systemctl stop $SERVICE_NAME"
if [[ $? -ne 0 ]]; then
    echo "Failed to stop service. Exiting."
    exit 1
fi

# -----------------------------
# Step 4: Copy the new binary to VPS
# -----------------------------
echo "Uploading new binary..."
scp app samiul@"$VPS_HOST":"$REMOTE_PATH"
if [[ $? -ne 0 ]]; then
    echo "SCP failed. Exiting."
    exit 1
fi

# -----------------------------
# Step 5: Restart the service
# -----------------------------
echo "Restarting remote service..."
ssh samiul@"$VPS_HOST" "sudo systemctl restart $SERVICE_NAME && sudo systemctl status $SERVICE_NAME --no-pager"
if [[ $? -ne 0 ]]; then
    echo "Failed to restart service."
    exit 1
fi

# -----------------------------
# Step 6: Ping the endpoint
# # -----------------------------
echo "Pinging API..."
curl -s -o /dev/null -w "%{http_code}\n" "$PING_URL"
