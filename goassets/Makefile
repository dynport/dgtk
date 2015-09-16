TPL := $(shell find tpl/goassets-tpl -type f) 
GOASSETS_TPL := $(shell find tpl/goassets-script-tpl -type f)

.PHONY: goassets.go

default: goassets.go test vet lint cleanup_tmp

vet: goassets.go
	go vet .

lint: goassets.go
	golint .

goassets.go: $(TPL) $(GOASSETS_TPL) generate.sh
	bash generate.sh > goassets.go

test: goassets.go
	go test

cleanup_tmp:
	rm -Rf tmp

clean: cleanup_tmp
	rm -f goassets.go
