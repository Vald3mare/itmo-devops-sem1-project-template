#!/bin/bash
ls -la
echo "Building the application..."
go build -o server cmd/server/main.go

echo "Starting the application in background..."
nohup ./server > app.log 2>&1 &
echo "Application started successfully in background."
echo "Logs are being written to app.log"