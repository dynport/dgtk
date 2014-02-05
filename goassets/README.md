# goassets: Building Assets

This is a build tool to generate assets and integrate them into a binary. This is an essential requirement for all tools
that should be deployed binary only.


## Installation

Installation follows the golang scheme. Run

	go get github.com/dynport/dgtk/goassets

to generate the `goassets` binary. It should be directly accessible (if `$GOPATH/bin` is in your `PATH` environment
variable).


## Requirements

There should be a dedicated directory in your package, that contains all assets (eg. `./assets`). Assets can be anything
but go files (having the `.go` suffix).


## Generation Of Assets

Run the `goassets` tool on your assets folder:

	goassets ./assets

This will generate a `assets.go` file that contains all (zipped) assets and some helper structures to use them. The
default target file name `assets.go` can be modified using the `--target` option.


## Usage Of Assets

The generated go source file contains a map from the assets' names to their respective contents and functions to access
those files. To get a list of all available assets use `assetNames` function with the following signature:

	assetNames() []string

Reading a single asset is possible using the `readAsset` function with this signature:

	readAsset(key string) ([]byte, error)

If handling these errors is not required and a panic is tolerated, then `mustReadAsset` is available:

	mustReadAsset(key string) []byte

Besides these helper methods the assets are handled using a fake filesystem approach (idea taken from [these slides by
Andrew Gerrand](http://nf.wh3rd.net/10things/#8)), i.e. you can use the `Open` method of the `assets` value. That will
return a value that implements the `ReadCloser` interface.

	asset, e := assets.Read("some_asset")
	if e != nil {
		log.Fatal(e)
	}
	defer asset.Close()

	c, _ := ioutil.ReadAll(asset)
	log.Print(string(c))


## Usage Of Assets In Development Mode

Sometimes it might not be appropriate to rebuild the assets each time something changes, i.e. when assets are used by a
web server it would be necessary to restart the server after updating the assets. Therefore there is the option to
specify the `GOASSETS_<pkg_name>_PATH` environment variable, that will trigger usage of this path as source for the
assets (this is where the fake filesystem is used actually). The part with `pkg_name` must be replaced by the name of
the respective package. This allows for using multiple packages with assets simultaneously and decided on a per package
basis which assets to use from the binary or not.

In most cases it should be sufficient to have a decent makefile at hands that will trigger an update if necessary.


## Makefile Integration

The following snippet might help a lot in a Makefile:

	ASSETS_DIR := ./assets
	ASSETS     := $(shell find $(ASSET_DIR) -type f | grep -v ".go$$")
	
	build: assets.go
		# whatever you need to do
	
	assets.go: $(ASSETS)
		@rm -f assets.go
		@goassets $(ASSETS_DIR) > /dev/null 2>&1

This will collect all available assets and add them as dependency to the assets.go file. The assets will only be built
when necessary (i.e. let `make` do its magic).

If there are multiple sub-packages with assets in the top level package the following Makefile snippet might help:

	ASSET_DIRS  := $(patsubst %/.goassets,%,$(shell find . -type f -name .goassets))
	ASSET_FILES := $(addsuffix /assets.go,$(ASSET_DIRS))
	
	build: $(ASSET_FILES)
		# whatever you need to do
	
	%/assets.go: %*/assets/*
		@rm -f $@
		@cd $* && goassets assets

This searches for directories containing the `.goassets` file (a marker for assets so to say) and compiles those when
required.


## Recommendation

For libraries we recommend to check in the compiled binary assets into your VCS. This makes sure it possible to build
and install using `go get` as usually done. Otherwise users see error messages (because of the missing `assets.go` file
in a fresh clone of the repository) and can't easily add the library.

The problem of out-of-sync source and binary assets can be prevented using make (see previous section). If assets change
the resulting `assets.go` file will have changes and thereby be shown by the VCS toolchain. Additionally you should
consider your CI tool to break in those conditions.


## ToDo

* Add support for subdirectories in the assets folder.
