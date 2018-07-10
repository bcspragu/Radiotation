#!/bin/bash
gin \
  --port=8080 \
  --bin=data/gin-bin \
  --appPort=8000 \
  --path=../.. \
  --build=. \
  --excludeDir=frontend \
  run

