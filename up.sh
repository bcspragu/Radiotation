#!/bin/bash
cd ./cmd/goose
go build
cd ../../
./cmd/goose/goose -dir=sql/migrations sqlite3 ./radiotation.db up
