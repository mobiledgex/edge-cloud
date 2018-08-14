# dockerfiles

These docker files are for producing consistent container images of edge-cloud.

We use a private registry at `registry.mobiledgex.net:5000`. You need to login to this registry via `docker login`. The login and password are available if you ask around.

## Dockerfile.protoc

This is for producing a consistent protoc stubs with versioned base tools from known sources. It is not essential, as it is seldom done by end users. See below for the `build` dockerfile.  

## Dockerfile.build

This is for container that is used to build edge-cloud. It is used by Dockerfile.edge-cloud.  When the docker image is built
using Dockerfile.build, it will create an image that contains all the necessary bits to compile edge-cloud.  It can be used
by the Docker.edge-cloud as base image.  This is seldom created because it takes a long time. The idea is to create this once, and keep using it to have a consistent build base. It should be available to pull from the private registry at `registry.mobiledgex.net:5000`.  

## Dockerfile.edge-cloud

This is to build a container image that can produce a runnable image containing edge-cloud binaries. The base image from 
Dockerfile.build is used to build the edge-cloud binaries. In the second stage, the base image is switched to alpine, to 
reduce the final artifact size, which will copy in the built binaries from the first stage into runnable custom base
container which has openstack and kubectl, etc. The entry point of this docker container image will be from `edge-cloud-entrypoint.sh` which takes first
argument which has to be the name of the binary to run.  All of the binaries such as controller, crmctl, edgectl, crmserver, etc. are available to be run. Subseqent arguments are passed to each program.
Before building this container you may want to add an entry to your `/etc/hosts` file:

```
$ echo 37.50.143.121 fmbncisrs101.tacn.detemobil.de >> /etc/hosts
```

This is because the domain name used is Bonn openstack cluster which is not registered to public DNS server.  


## building

First make sure your source tree checked out of github builds OK locally.  You should do `dep ensure -v` and `make` locally to make sure.

To build `mobiledgex/edge-cloud` container run the following in the main directory of `edge-cloud` source tree.
or just run `make build-docker`.


```
$ cd ~/src/github.com/mobiledgex/edge-cloud
$ docker build -t mobiledgex/edge-cloud -f docker/Dockerfile.edge-cloud .
$ docker login registry.mobiledgex.net:5000  #if you have not done so already
$ docker tag mobiledgex/edge-cloud registry.mobiledgex.net:5000/mobiledgex/edge-cloud
$ docker push registry.mobiledgex.net:5000/mobiledgex/edge-cloud
```

## Usage

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

## Docker-compose

You can create instances of controller, crmserver, etc. using `docker-compose`.
For the details please read `docker-compose.yml` file.  It creates three docker container
instances.  An etcd server, controller and crmserver.  You will need to install `docker-compose` if you don't already have it.  The `etcd` , `controller` and `crmserver` will be running on the machine you are running `docker-compose`.  This is probably your laptop. The `app` will be launched in a cloudlet kubernetes cluster that `crmserver` creates based on your test scripts.  An example is in `test-edgectl.sh`.

The `docker-compose` proves that the container images can run anywhere as long as the host supports docker runtime. You can create `docker-machine` in multiple cloud hosting providers and launch multiple instances of each docker container.  The example for that will be provided later.

First, get a copy of `mex-docker.env` file from a private repo: https://github.com/mobiledgex/bob-priv/mex-docker.env

Place `mex-docker.env` in the same directory as `docker-compose.yml` file.
You will also need to place `.mobiledgex` directory at your `$HOME`.
Get http://github.com/mobiledgex/bob-priv/mobiledgex.env.tar and untar
into your home directory. You also need to copy .mobiledge dir under this directory before building the docker image for edge-cloud. There is also http://github.com/mobiledgex/bob-priv/edgecloud.tar which must be untarred to this directory before running docker-compose.

You also should make some changes in docker-compose.yaml and test-edgectl.sh files. Change all the `test` to your own name.  For example,
change `testoperator` to `boboperator`, etc.
This is to avoid clashing with other testers in the same backend infrastructure.  When you do this, you will also need to update a testapp.yaml with yourownnameapp.yaml and place it on registry.mobiledgex.net.  This is done manually currently. There will be more automated or UI based methods in the future.

```
$ cd ~/src/github.com/mobiledgex/edge-cloud/docker
$ # place the mex-docker.env file, .mobiledgex directory, and edgecloud directory here.
$ docker-compose up -d
$ docker ps
CONTAINER ID        IMAGE                                                COMMAND                  CREATED             STATUS              PORTS               NAMES
c2add5a4ad36        gcr.io/etcd-development/etcd:v3.3.9                  "/usr/local/bin/etcd…"   13 minutes ago      Up 13 minutes                           docker_etcd-gcr-v3.3.9_1
8139edf89734        registry.mobiledgex.net:5000/mobiledgex/edge-cloud   "edge-cloud-entrypoi…"   13 minutes ago      Up 13 minutes                           docker_controller_1
1d4cf3d9dc08        registry.mobiledgex.net:5000/mobiledgex/edge-cloud   "edge-cloud-entrypoi…"   13 minutes ago      Up 13 minutes                           docker_crmserver_1
```

You can then run `edgectl` to tell controller to send messages to `crmserver`.
Some examples are in `test-edgectl.sh` file.

After running `edgectl` commands you should have a testapp running the the created kubernetes cluster in a cloudlet. 
To verify you can do:

```
$ docker exec -e KUBECONFIG=/root/.mobiledgex/mex-k8s-master-1-testcluster.kubeconfig-proxy -it docker_crmserver_1   kubectl get  nodes
NAME                           STATUS    ROLES     AGE       VERSION
mex-k8s-master-1-testcluster   Ready     master    12m       v1.11.2
mex-k8s-node-1-testcluster     Ready     <none>    11m       v1.11.2
mex-k8s-node-2-testcluster     Ready     <none>    11m       v1.11.2
$ docker exec -e KUBECONFIG=/root/.mobiledgex/mex-k8s-master-1-testcluster.kubeconfig-proxy -it docker_crmserver_1   kubectl get  pods
NAME                                  READY     STATUS    RESTARTS   AGE
testapp-deployment-74db74fc9b-mxs8p   1/1       Running   0          6m
testapp-deployment-74db74fc9b-szk97   1/1       Running   0          6m
$ docker exec -e KUBECONFIG=/root/.mobiledgex/mex-k8s-master-1-testcluster.kubeconfig-proxy -it docker_crmserver_1   kubectl get deploy
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
testapp-deployment   2         2         2            2           7m
$  docker exec -e KUBECONFIG=/root/.mobiledgex/mex-k8s-master-1-testcluster.kubeconfig-proxy -it docker_crmserver_1   kubectl get svc
NAME              TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                                           AGE
kubernetes        ClusterIP      10.96.0.1       <none>         443/TCP                                           13m
testapp-service   LoadBalancer   10.106.17.184   10.101.103.2   27272:31248/TCP,27273:31928/TCP,27274:30246/TCP   7m
```

The `testapp` is deployed in 2 different pods.  The service for the deployment is at 10.101.103.2 which is the kubernetes
cluster master node in the internet network.  From the kubernetes cluster point of view this is the `external address`
assigned by the external load balancer.  In our case, the `mexosagent` handles the assignment.  The reverse proxy
inside `mexosagent` also has paths registered and terminates TLS.  From the Internet, the testapp can be accessed.

To list the proxy assignment of paths:

```
$ curl -s -XPOST  http://testcloudlet.testoperator.mobiledgex.net:18889/v1/proxy -d '{"message":"list"}' | jq .
{
  "message": "list",
  "status": "map[/testappgrpc/*catchall:http://10.101.103.2:27272 /testapprest/*catchall:http://10.101.103.2:27273 /testapphttp/*catchall:http://10.101.103.2:27274]"
}
```

To talk to `testapp` at its HTTP via GET:

```
$ curl https://testcloudlet.testoperator.mobiledgex.net/testapphttp/
hostname testapp-deployment-74db74fc9b-mxs8poutbound ip 10.36.0.1{1 65536 lo  up|loopback} [127.0.0.1/8] {12 1376 eth0 c2:8c:7e:bd:54:66 up|broadcast|multicast} [10.36.0.1/12]
```

To POST to REST endpoint of `testapp`:

```
$ curl -XPOST https://testcloudlet.testoperator.mobiledgex.net/testapprest/info -d '{"message": "info"}'
{"message":"info","outbound":"10.44.0.1","hostname":"testapp-deployment-74db74fc9b-szk97","interfaces":[{"name":"lo","addresses":"[127.0.0.1/8]"},{"name":"eth0","addresses":"[10.44.0.1/12]"}]}
```

Both `testapp` and `mexosagent` also have GRPC endpoints.
