#!/bin/bash

REMOTE_HOST="$1"
REMOTE_USER="root"
SSH_KEY="$2"
REMOTE_PATH="./"
SERVER_ADDR_BASE="$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH"

ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" << EOF
mkdir -p $REMOTE_PATH/grafana
mkdir -p $REMOTE_PATH/auth/db
EOF

scp -i "$SSH_KEY" "./docker-compose.yml" "$SERVER_ADDR_BASE/docker-compose.yml"
scp -i "$SSH_KEY" "./.env" "$SERVER_ADDR_BASE/.env"
scp -i "$SSH_KEY" -r "./prometheus" "$SERVER_ADDR_BASE/prometheus"
scp -i "$SSH_KEY" "./grafana/grafana.ini" "$SERVER_ADDR_BASE/grafana/grafana.ini"
scp -i "$SSH_KEY" -r "./grafana/provisioning" "$SERVER_ADDR_BASE/grafana/provisioning"
scp -i "$SSH_KEY" "./auth/db/init.sql" "$SERVER_ADDR_BASE/auth/db/init.sql"

ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" << EOF
cd $REMOTE_PATH || exit 1
docker compose pull
docker compose down
docker compose up -d --no-build --remove-orphans
EOF
