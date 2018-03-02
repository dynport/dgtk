#!/bin/bash
set -e

DIR=/go/src/github.com/dynport/dgtk
docker run --volume $(pwd):${DIR} --interactive --workdir ${DIR} --rm golang:1.10 <<EOF
go get -v -t -d ./...
make ego
go test -v ./...
go vet -v ./...
EOF
