.PHONY: build check default deps test vet

ALL_DEPS    := $(shell go list ./... | xargs go list -f '{{join .Deps "\n"}}' | grep -e "$github.com\|code.google.com\|launchpad.net" | sort | uniq | grep -v "github.com/dynport/dgtk")
EXTRA_DEPS  := github.com/dynport/dgtk/goassets github.com/smartystreets/goconvey github.com/stretchr/testify/assert github.com/jacobsa/oglematchers
IGN_DEPS    := 
DEPS        := $(filter-out $(IGN_DEPS),$(ALL_DEPS))

ALL_PKGS    := $(shell go list ./...)
IGN_PKGS    := github.com/dynport/dgtk/goassets/script/tpl
PACKAGES    := $(filter-out $(IGN_PKGS),$(ALL_PKGS))
IGN_TEST_PKGS := github.com/dynport/dgtk/es github.com/dynport/dgtk/es/aggregations
TEST_PKGS   := $(filter-out $(IGN_TEST_PKGS),$(PACKAGES))

default: build

build:
	@go install $(PACKAGES)

check:
	@which go > /dev/null || echo "go not installed"

deps:
	@for package in $(EXTRA_DEPS) $(DEPS); do \
		echo "Installing $$package"; \
		go get -u $$package; \
	done

test: build
	@go test $(TEST_PKGS)

vet: build
	@go vet $(PACKAGES)

