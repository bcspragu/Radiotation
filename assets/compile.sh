#!/bin/bash
set -e

docker build -t radiotation-webpack .
docker run -p 8081:8081 --mount type=bind,source=$PWD,destination=/project/assets --rm radiotation-webpack
