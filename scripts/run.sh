#!/bin/bash

echo "Building the application..."
cd сmd/api
go mod tidy
go build -o app main.go
./cmd/api/app

echo "Starting the application in background..."
nohup ./app > app.log 2>&1 &

echo "Application started successfully in background."
echo "Logs are being written to app.log"