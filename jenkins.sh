#!/bin/bash -xe

export GOROOT=${GOROOT:-/usr/local/go1.4.3}
if [[ -n $WORKSPACE ]]; then
  export GOPATH=$WORKSPACE
fi

export PATH=$GOPATH/bin:$GOROOT/bin:$PATH/
export DYNPORT_GO_PATH=$GOPATH/src/github.com/dynport
mkdir -p $DYNPORT_GO_PATH $WORKSPACE/bin

pushd $DYNPORT_GO_PATH/dgtk

go get -d ./...
make ego
go test -v ./...
go vet ./...

d=$(mktemp -d /tmp/build-XXXX)

info=$(bash ./build_info.sh)

for os in darwin linux; do
  for name in wunderproxy/wunderproxy wunderproxy/wunderstatus; do
    n=$(basename $name)
    echo "building $n with os $os"
    GOOS=$os go build -ldflags "-X main.BUILD_INFO $info" -o $d/${os}_amd64/$n github.com/dynport/dgtk/$name
  done
done

if [[ -n $GIT_COMMIT && -n $JOB_NAME && -n $BUILDS_BUCKET ]]; then
  aws s3 sync --delete $d/ s3://${BUILDS_BUCKET}/builds/${JOB_NAME}/$GIT_COMMIT/
fi
