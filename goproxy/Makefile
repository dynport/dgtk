GEM_PATH = /tmp/gems

default: build

restart: build
	killall goproxy

build:
	go get ./...

run: build
	goproxy

clean:
	rm *.test
	rm -Rf tmp

test:
	go test -v
