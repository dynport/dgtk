GEM_PATH = /tmp/gems

default: build

build:
	go get ./...

run: build
	goproxy

clean:
	rm *.test
	rm -Rf tmp

test:
	go test -v
