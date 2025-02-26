#!/bin/bash

export PATH=$PATH:/usr/local/go/bin
go version

go get github.com/jackc/pgx/v5
go get github.com/joho/godotenv

source database.env

APP_PATH="/cmd/api/main.go"
BINARY_PATH="${GITHUB_WORKSPACE}/main"

echo $BINARY_PATH