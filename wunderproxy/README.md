# wunderproxy (working title)

wunderproxy is the tool which handles all of our deployments. It runs on all application servers and consists of the following components:

* proxy:    proxy requests to docker containers (running on the docker host sytem)
* launcher: start docker containers for autoscaling instances via userdata

## Proxy

The proxy should be just a simple `switch` which proxies requests to one `backend`. In our case the `backend` is most of the time a docker container, but it can also be e.g. the maintenance page.

The proxy provides an API endpoint with which one can set a new backend to which all new requests should be proxied.

For now, there should be no health check implemented in the proxy.

### Extras

#### 502 handling

When usign using e.g. ELB the proxy could handle 502 by proxying requests to other instances in the ELB (not the status requests from the ELB health check though). Those requests should get an extra header which mark them as `fallback requests`. Those requests should not be proxied again.


#### Status page

The proxy should have a basic status page with e.g. the stats about the number of open connections by states, historic status codes and response times etc.

#### Compression, SSL, etc.

For e.g. the phrase unicorn deployments the Proxy would proxy to an nginx process running in front of unicorn (which acts as a `fast client` - see http://unicorn.bogomips.org/PHILOSOPHY.html). The only benefit of unicorn would be seving of static files and things like compression, ssl termination etc. In theory this could all be done by the proxy which would make the extra nginx disposable.

## Launcher

The launcher is in charge of launching new containers. For starting it just takes an URL to a JSON configuration (prefereable on S3). The JSON config contains the following information.

* Docker image to launch
* Docker configuration for launching the container (e.g. exports, environment variables, etc.)
* Health check URL
* S3 bucket for image artifacts

When triggered (either via CLI or via API request) the launcher creates a new container, waits for the health check to be OK (with some timeout) and then updates the `proxy` with the URL of the new container.

The launcher can be executed from the command line via e.g. `wunderproxy launch https://de-dynport-docker.s3-eu-west-1.amazonaws.com/phrase.config`, then fetches that url and creates a new container, etc. S3 URLs should be detected and handled with gocloud/aws/s3 (with the AWS keys either from env or via IAM instacne profile).

The launcher can be also executed via API call. In that case it would re-fetch the config document, start a new container and update the proxy.

### Extras

#### Rollbacks

The launcher could store the previous configurations and provide an API endpoint to roll back to configuration before the current one.

#### Registry

The docker registry could be the single point of failure. So it would make sense to also provide the registry functionality in `wunderproxy`. In that case the image artifacts must be stored on S3 and the bucket needs to be in the configuration. A basic implementation of the docker registry with S3 as vbackend can be found at `github.com/dynport/dgtk/dpr`.
