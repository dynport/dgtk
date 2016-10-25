#!/bin/bash
set -e

export GOROOT=${GOROOT:-/usr/local/go1.4.3}
if [[ -n $WORKSPACE ]]; then
  export GOPATH=$WORKSPACE
fi

export PATH=$GOPATH/bin:$GOROOT/bin:$PATH/
export DYNPORT_GO_PATH=$GOPATH/src/github.com/dynport
mkdir -p $DYNPORT_GO_PATH $WORKSPACE/bin

pushd $DYNPORT_GO_PATH/dgtk

go get -t -d ./...
make ego
go test ./...
go vet ./...

d=$(mktemp -d /tmp/build-XXXX)

info=$(bash ./build_info.sh)

if [[ -z $info ]]; then
  echo "info must not be blank"
  exit 1
fi

for os in darwin linux; do
  for name in github/gh wunderproxy/wunderproxy wunderproxy/wunderstatus dpr/dpr dp-es; do
    echo "building $name with os $os"
    GOOS=$os go build -ldflags "-X main.BUILD_INFO $info" -o $d/${os}_amd64/$(basename $name) github.com/dynport/dgtk/$name
  done
done

if [[ -n $GIT_COMMIT && -n $JOB_NAME && -n $BUILDS_BUCKET ]]; then
  aws s3 sync --delete $d/ s3://${BUILDS_BUCKET}/builds/${JOB_NAME}/$GIT_COMMIT/
  echo $info > $d/current.json
  aws s3 cp $d/current.json s3://${BUILDS_BUCKET}/builds/${JOB_NAME}/current.json
fi
