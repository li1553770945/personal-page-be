#!/usr/bin/env bash
RUN_NAME="personal-page-be"
mkdir -p output/bin
cp run.sh output
go env -w GOPROXY=https://goproxy.cn
go build -o output/bin/${RUN_NAME}
