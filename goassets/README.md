# goassets: Building Assets

This is a build tool to generate assets and integrate them into a binary.


## Installation

Run
 `go get github.com/dynport/dgtk/goassets`
to generate the `goassets` binary. It should be directly accessible (if $GOPATH/bin is in your PATH environment
variable).


## Requirements

There should be a dedicated directory in your package, that contains all assets (eg. ./assets). Assets can be anything
but go files (having the .go suffix).


## Generation Of Assets

Run the `goassets` tool on your assets folder:
 `goassets ./assets`

This will generate a assets/assets.go file that contains all (zipped) assets and some helper structures to use them.


## Usage Of Assets

Import the assets subpackage and use the provided helper methods. Available methods are:
 * `Get(key string) ([]byte, error)` to fetch a dedicated asset.
 * `Names() []string` to return the list of available asset names.


## Makefile Integration

The following snippet might help a lot in a Makefile:
 `ASSETS_DIR := ./assets
  ASSETS     := $(shell find $(ASSET_DIR) -type f | grep -v ".go$$")
  $(ASSETS_DIR)/assets.go: $(ASSETS)
          @goassets $(ASSETS_DIR) > /dev/null 2>&1`

This will collect all available assets and add them as dependency to the assets.go file. The assets will only be built
when necessary (i.e. let `make` do its magic).


## Troubleshooting

The only real hassle might be that assets have not been compiled when you try to build stuff. Run `goassets` first.


## ToDo

 * Add support for subdirectories in the assets folder.
