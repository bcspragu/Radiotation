#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker run \
  -it \
  -u $(id -u):$(id -g) \
  --mount type=bind,source=$DIR/frontend,destination=/project \
  --rm \
  node-env yarn build
