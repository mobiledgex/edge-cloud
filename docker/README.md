# dockerfiles

These docker files are for producing consistent container images of edge-cloud.

## Dockerfile.protoc

This is for producing a consistent protoc stubs with versioned base tools from known sources. It is not essential. See below for the `build` dockerfile

## Dockerfile.build

This is for container that is used to build edge-cloud. It is used by Dockerfile.edge-cloud.  When the docker image is built
using Dockerfile.build, it will create an image that contains all the necessary bits to compile edge-cloud.  It can be used
by the Docker.edge-cloud as base image.

## Dockerfile.edge-cloud

This is to build a container image that can produce a runnable image containing edge-cloud binaries. The base image from 
Dockerfile.build is used to build the edge-cloud binaries. In the second stage, the base image is switched to alpine, to 
reduce the final artifact size, which will copy in the built binaries from the first stage into runnable `alpine` based
container. The entry point of this docker container image will be from `edge-cloud-entrypoint.sh` which takes first
argument which has to be the name of the binary to run.  All of the binaries such as controller, crmctl, edgectl, crmserver, etc. are available to be run. Subseqent arguments are passed to each program.


For example,

```
 docker run -it --rm mobiledgex/edge-cloud controller -h
Usage of controller:
  -apiAddr string
        API listener address (default "127.0.0.1:55001")
  -d string
        comma separated list of [etcd api notify dmedb dmereq]
  -etcdUrls string
        etcd client listener URLs (default "http://127.0.0.1:2380")
  -httpAddr string
        HTTP listener address (default "127.0.0.1:8091")
  -localEtcd
        set to start local etcd for testing
  -notifyAddr string
        Notify listener address (default "127.0.0.1:50001")
  -r string
        root directory; set for testing
  -region uint
        Region (default 1)
```

And,


```
docker run -it --rm mobiledgex/edge-cloud edgectl -h
Usage:
  edgectl [command]

Available Commands:
  controller
  dme
  crm
  loc-api-sim
  tok-srv-sim
  completion-script Generates bash completion script
  help              Help about any command

Flags:
      --addr string            address to connect to (default "127.0.0.1:55001")
  -h, --help                   help for edgectl
      --output-format string   [yaml json json-compact table] (default "yaml")

Use "edgectl [command] --help" for more information about a command.
```


To build this container run the following in the main directory of `edge-cloud` source tree.
or just run `make build-docker`.


```
docker build -t mobiledgex/edge-cloud -f docker/Dockerfile.edge-cloud .
```
