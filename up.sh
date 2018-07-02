#!/bin/bash
cd ./cmd/goose
go build
cd ../../
./cmd/goose/goose -dir=sqldb/migrations sqlite3 ./radiotation.db up
