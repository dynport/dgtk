GEM_PATH = /tmp/gems

default: build

build:
	go get ./...

run: build
	goproxy

test:
	go test -v
