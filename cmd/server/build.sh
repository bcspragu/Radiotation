#!/bin/bash
CGO_ENABLED=0 GOOS=linux vgo build -a -installsuffix cgo -o server
docker build -t docker.bsprague.com/radiotation .
