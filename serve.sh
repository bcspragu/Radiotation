#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

CMD=serve
if [[ $# -eq 1 ]] ; then
  CMD=$1
fi

docker run \
  -it \
  -u $(id -u):$(id -g) \
  --net=host \
  --mount type=bind,source=$DIR/frontend,destination=/project \
  --rm \
  node-env yarn $CMD
