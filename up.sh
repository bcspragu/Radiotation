#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

go build -o $DIR/cmd/goose/goose $DIR/cmd/goose
$DIR/cmd/goose/goose -dir=$DIR/sqldb/migrations sqlite3 $DIR/cmd/server/radiotation.db up
