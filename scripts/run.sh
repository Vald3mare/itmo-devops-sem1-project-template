#!/bin/bash

set -e

echo "Билдинг приложения..."
cd cmd
cd api
go mod tidy
go build -o app main.go

echo "Приложение запущено успешно."
echo "Logs are being written to app.log"