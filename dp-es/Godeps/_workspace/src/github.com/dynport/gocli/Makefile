default: build test

build:
	go get -t ./...

clean:
	rm -f bin/*

test: clean build
	go test
