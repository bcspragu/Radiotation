#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -f $DIR/env.vars ]; then
   source $DIR/env.vars
fi

CC=gcc vgo build github.com/bcspragu/Radiotation/cmd/server
./server --addr=:8080
