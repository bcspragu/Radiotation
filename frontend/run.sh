#!/bin/bash
sudo rkt run --insecure-options=image --stage1-name=coreos.com/rkt/stage1-coreos:1.27.0 --volume app,kind=host,source=$PWD,readOnly=false ~/Projects/containers/npm/npm-latest.aci --exec=/bin/sh --interactive --net=host
