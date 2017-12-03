#!/bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Building a binary..."
GOOS=linux go build -a -ldflags '-extldflags "-static" -s -w' -o radiotation .

echo "Building the assests..."
docker build -t npm-env $DIR/frontend
docker run --mount type=bind,source=$DIR/frontend,destination=/project --rm npm-env /bin/sh -c "npm run build"

echo "Cleaning old assets..."
ssh prod "rm -rf ~/rkt/Radiotation/frontend/dist/*"

echo "Uploading new assets..."
scp -r frontend/dist/* prod:~/rkt/Radiotation/frontend/dist/

echo "Uploading new binary..."
scp radiotation prod:~/rkt/Radiotation

read -p 'Password: ' -s password

ssh prod << EOF
  cd ~/rkt/Radiotation/

  echo "Building new image..."
  echo "$password" | sudo -S ./build-image.sh

  echo "Killing old service..."
  echo "$password" | sudo -S ./kill-service.sh

  echo "Starting new service..."
  echo "$password" | sudo -S ./run-image.sh
EOF
