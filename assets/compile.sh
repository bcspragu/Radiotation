#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker build -t radiotation-webpack $DIR
docker run -p 8081:8081 --mount type=bind,source=$DIR,destination=/project/assets --rm radiotation-webpack
