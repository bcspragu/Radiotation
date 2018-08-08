#!/bin/bash
CC=gcc vgo build github.com/bcspragu/Radiotation/cmd/server
./server --addr=:8080
