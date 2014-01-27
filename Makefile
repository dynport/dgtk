.PHONY: build check clean default deps test vet

ASSET_DIRS  := $(shell find . -type f -name .goassets)
ASSET_FILES := $(addsuffix assets.go,$(dir $(ASSET_DIRS)))

ALL_DEPS    := $(shell go list ./... | xargs go list -f '{{join .Deps "\n"}}' | grep -e "$github.com\|code.google.com\|launchpad.net" | sort | uniq | grep -v "github.com/dynport/dgtk")
EXTRA_DEPS  := github.com/dynport/dgtk/goassets github.com/smartystreets/goconvey github.com/stretchr/testify/assert
IGN_DEPS    := 
DEPS        := $(filter-out $(IGN_DEPS),$(ALL_DEPS))

ALL_PKGS    := $(shell go list ./...)
IGN_PKGS    := 
PACKAGES    := $(filter-out $(IGN_PKGS),$(ALL_PKGS))

default: build

build: $(ASSET_FILES)
	@go install $(PACKAGES)

check:
	@which go > /dev/null || echo "go not installed"
	@which goassets > /dev/null || echo "go assets missing, call 'go get github.com/dynport/dgtk/goassets'"

clean:
	@rm -f $(ASSET_FILES)

%/assets.go:
	@rm -f $@
	@cd $* && goassets assets

deps:
	@for package in $(EXTRA_DEPS) $(DEPS); do \
		echo "Installing $$package"; \
		go get -u $$package; \
	done

test: build
	@go test -v $(PACKAGES)

vet: build
	@go vet $(PACKAGES)

