TPL := $(shell find tpl -type f) 
SCRIPT := $(shell find script -type f)

goassets.go: $(TPL) $(SCRIPT) generate.sh
	bash generate.sh > goassets.go

clean:
	rm -f goassets.go
