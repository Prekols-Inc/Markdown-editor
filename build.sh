#!/bin/bash

BACKEND_SERVICE_NAME="backend"
AUTH_SERVICE_NAME="auth"
FRONTEND_PORT=8081

echo "Starting auth service..."
cd auth
go mod tidy
go build -o ${AUTH_SERVICE_NAME} .
./${AUTH_SERVICE_NAME} >> ${AUTH_SERVICE_NAME}.log 2>&1 & 

echo "Starting storage service..."
cd ../backend
go mod tidy
go build -o ${BACKEND_SERVICE_NAME} .
./${BACKEND_SERVICE_NAME} >> ${BACKEND_SERVICE_NAME}.log 2>&1 & 

echo "Starting frontend service..."
cd ../frontend
npm install
npm run build
nohup npm run preview -- --port ${FRONTEND_PORT} >> frontend.log 2>&1 &
