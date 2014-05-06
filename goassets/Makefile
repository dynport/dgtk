build:
	go get ./...

test:
	go test -v

testold: build
	echo "running tests"
	go test -v
	cd goassets-test && rm -f assets.go; goassets -pkg main ../fixtures && go run *.go
