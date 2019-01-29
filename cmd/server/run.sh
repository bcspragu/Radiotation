#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -f $DIR/env.vars ]; then
   source $DIR/env.vars
fi

go build -o $DIR/server github.com/bcspragu/Radiotation/cmd/server
cd $DIR
$DIR/server --addr=:8080
