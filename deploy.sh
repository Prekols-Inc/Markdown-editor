#!/bin/bash

REMOTE_HOST="87.228.113.72"
REMOTE_USER="root"
SSH_KEY="~/.ssh/id_rsa"
REMOTE_PATH="./"
SERVER_ADDR="$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

scp -i "$SSH_KEY" "./docker-compose.yml" "$SERVER_ADDR"
scp -i "$SSH_KEY" "./.env" "$SERVER_ADDR"
scp -i "$SSH_KEY" -r "./prometheus" "$SERVER_ADDR"
scp -i "$SSH_KEY" -r "./grafana" "$SERVER_ADDR"

ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" << EOF
cd $REMOTE_PATH || exit 1
docker compose pull
docker compose down
docker compose up -d --no-build --remove-orphans
EOF