#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker build -t npm-env $DIR
docker run -it -p 8081:8081 --mount type=bind,source=$DIR,destination=/project --rm npm-env /bin/sh
