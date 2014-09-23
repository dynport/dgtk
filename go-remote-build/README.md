# go-remote-build

Builds a go package on a *real* linux host (e.g. a local VM) and (optionally) uploads the binary to a specific host (via ssh) or S3 bucket.

All local dependencies are uploaded to the linux box before building (from the current GOPATH) so no extra dependencies need to be fetched on the linux host.

    Usage of go-remote-build:
      -bucket="": Upload binary to s3 bucket after building
      -deploy="": Deploy to host after building. Example: ubuntu@127.0.0.1
      -dir="": Dir to build. Default: current directory
      -host="ubuntu@172.16.223.140": Host to build on. Example: ubuntu@127.0.0.1
      -verbose=false: Build using -v flag
