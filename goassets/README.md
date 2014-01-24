# goassets: Building Assets

This is a build tool to generate assets and integrate them into a binary. This is an essential requirement for all tools
that should be deployed binary only.


## Installation

Run
	go get github.com/dynport/dgtk/goassets
to generate the `goassets` binary. It should be directly accessible (if `$GOPATH/bin` is in your `PATH` environment
variable).


## Requirements

There should be a dedicated directory in your package, that contains all assets (eg. `./assets`). Assets can be anything
but go files (having the `.go` suffix).


## Generation Of Assets

Run the `goassets` tool on your assets folder:
	`goassets ./assets`

This will generate a `assets.go` file that contains all (zipped) assets and some helper structures to use them.
Additionally a `.goassets` file will be added that can be used with make to automatically generate assets when required.
It contains the source folder and the target file.

The target file name can be modified using the `--target` option.


## Usage Of Assets

The generated go source file contains a map from the assets' names to their respective contents. Retrieval of assets is
done using the following methods:

* `readAsset(key string) ([]byte, error)` to fetch a dedicated asset.
* `mustReadAsset(key string) ([]byte)` panics if the given asset is not found.
* `assetNames() []string` returns the list of names of available assets.


## Makefile Integration

The following snippet might help a lot in a Makefile:

	ASSETS_DIR := ./assets
	ASSETS     := $(shell find $(ASSET_DIR) -type f | grep -v ".go$$")
	assets.go: $(ASSETS)
		@rm -f assets.go
		@goassets $(ASSETS_DIR) > /dev/null 2>&1

This will collect all available assets and add them as dependency to the assets.go file. The assets will only be built
when necessary (i.e. let `make` do its magic).

For libraries we recommend to check in the compiled binary assets into your VCS. The problem of out-of-sync source and
binary assets can be prevented by using make. If assets change the resulting `assets.go` file will have changes and
thereby be shown by the VCS toolchain. Additionally you should consider your CI tool to break in those conditions.


## ToDo

* Add support for subdirectories in the assets folder.
