#!/bin/bash
set -e

CC=gcc vgo build github.com/bcspragu/Radiotation/cmd/server
./server --addr=:8080
