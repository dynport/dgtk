# wunderproxy

wunderproxy is a tool to simply manage docker based deployments. It consists of two components:

* proxy:   Forward requests on a specified port to the currently deployed docker container.
* manager: Manage docker containers via an API.


## The Manager

The container manager is used to provide a simple API for handling docker containers and manage the proxy. It has the following actions:
 
* `/status`: This action returns status information on the currently active container, like the number of requests served.
* `/launch`: Will start a new container. The configuration identifier is read from the request body. The actual config file is fetched from S3. Using this information the container will be launched.
* `/switch`: Switches the proxy port to the container given.

The `launch` action fetches the actual configuration from S3 using the requested identifier (it's the hash of the configuration file) as last part of the S3 key. The JSON config expected under this key contains the following information:

* Docker image to launch
* Docker configuration for launching the container (e.g. exports, environment variables, etc.)
* Health check URL
* S3 bucket for image artifacts

The manager expects the currently running configuration to also be available with the `current.json` key suffix on S3. This is the container configuration that will be launched on startup.


## The Proxy

Using a reverse proxy approach the proxy will send incoming requests to a modifiable port. Some minor statistics are collected along the way, that can be retrieved using the manager's API. 


## Usage

The following options can be given to determine what happens when starting the wunderproxy:

* `ProxyAddress`: The listen address of the proxy. The default is to use `0.0.0.0:80`. This is the port that will be forwarded to the docker containers. 
* `ApiAddress`: The listen address of the manager API. The default is `0.0.0.0:8001`.
* `RegistryPort`: The current version of the wunderproxy relies on a local docker registry/distribution (see [here](https://github.com/docker/distribution/blob/master/docs/deploying.md)). The default is to expect the registry on port 8080.
* `ConfigFile`: This is a simple `key=value` based file that can be used to give static environment variables to containers. This is optional of course.

Besides these options there are three arguments required:

* The S3 bucket used to handle container launch configurations.
* The path prefix to be used in the S3 bucket.
* The name of the application handled. This is used as prefix for the containers images.


## Possible Improvements (TODO)

### 502 Handling

If the load is distributed over multiple machines the local wunderproxy instance could try to defer handling a certain requests to one of the other instances if a 502 status code was returned (i.e. if the container wasn't started or doesn't run any more).
 
* Those requests should get an extra header which mark them as `fallback requests`.
* Those requests should not be proxied again.
* The status action (used for health checks for example) shouldn't be forwarded of course.


### Status Page

The manager's status action could be improved to contain:

* the number of open connections by states
* historic status codes
* response times etc.


### Compression, SSL, etc.

Currently the wunderproxy just does the default buffering provided by the Go http library (we measured that to be about 1MB). For unicorn deployments this isn't sufficient to handle `slow clients` (see [unicorn philosophy](http://unicorn.bogomips.org/PHILOSOPHY.html)). We decided to simply use a nginx instance in front of wunderproxy, what of course adds another layer but is sufficient for our case. The nginx layer could be made disposable with an improved request handling in wunderproxy.


### Improved Container Switch

The old container should be kept around until all open connections are handled (or a timeout occurs).s