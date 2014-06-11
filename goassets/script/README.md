# goassets

goassets.go is a script to be put in your e.g. go web project for bundling assets into your binary. It is on purpose not a separate application to a) remove external build dependencies (only go and the go stdlib is needed) and b) so that all people building generate the same output.

## Bundling

    go run goassets.go --file assets/assets.go assets

This call will bundle all files found in assets/ into the file assets/assets.go. The recommended interface to access assets is provided vi FileSystem() function which is compatible to `http.FileSystem`.

## Examples

See `examples`
