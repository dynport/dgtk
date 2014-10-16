TPL := $(shell find tpl -type f) 
SCRIPT := $(shell find script -type f)

.PHONY: goassets.go

vet: goassets.go
	go vet .

lint: goassets.go
	golint .

goassets.go: $(TPL) $(SCRIPT) generate.sh
	bash generate.sh > goassets.go

test: goassets.go
	go test

clean:
	rm -f goassets.go
