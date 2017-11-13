#!/bin/bash
set -e

docker run -it -p 8081:8081 --mount type=bind,source=$PWD,destination=/project/assets --rm radiotation-webpack /bin/sh

