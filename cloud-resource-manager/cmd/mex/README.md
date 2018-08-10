# MEX utility

The `mex` utility serves two purposes:

* testing the infrastructure code
* equivalent in function to `gcloud` and `az` to get to kubernetes cluster based portable API

In addition `mex` utility uses API exposed by mexos_multi.go.  This file uses mexos.go and others.   The mexos_multi.go insulates and negotiates YAML files based APIs and parameter APIs.  The Controller sends ClusterInst and AppInst through cached Inst data updates.  They contain insufficient data and context for complex requirements of the backend.  The mexos_multi.go takes the parameters from Inst cache and builds YAML structures that can be used by the underlying APIs in mexos.go.  Both controller facing CRM server code and `mex` utility use the mexos_multi.go.  It will be helpful to read mexos_multi.go and mexos.go under crmutil directory.  These files under crmutil use APIs from edge-cloud-infra which are lower level.

**Beware that Letsencrypt throttles getting certificates. The platform commands should not run as part of automated test thing all the time :)**

# Environment

Openstack environment requires sourcing `openrc`.  For Bonn cloudlet you can
get a copy of  `https://github.com/mobiledgex/bob-priv/blob/master/mobiledgex.env.tar`

And untar in your HOME directory.

And do,
```
$ source .mobiledgex/openrc
```


## cloudflare

Normally not needed, but if you *have to* test `init` and `clean` methods of `platform` target of `mex` utility -- to create base platform -- then you will need cloudflare API user name and key, to create the DNS entry for the root LB you are creating.  Talk to Bob if you need this.

## registry.mobiledgex.net:5000

It is HTTPS and uses basic auth.

For user `mobiledgex` the password is the name of the street near the office, one word, all lowercase.  Talk to Bob if you need assistance.



## Letsencrypt

They have 20 API calls per week limit. Plus, they limit the number of times to recreate certs for the same FQDN. It seems that limit is 5.

## registry

Registry.mobiledgex.net is a gcp vm. It runs private docker registry at Port 500. Also maven Nexus https at 8081. It is fronted by nginx to terminate TLS at 8081.  Nexus also contains npm, gem and docker repos. They are not used currently.  Only maven.  There's also a seperate nginx file server for static content like qcow2, certificate, key and yaml files at 8080. All theses run as docker instances inside the vm. All these run as docker instances on registry machine.

Registry machine currently also hosts a slackgist server which does basic cicd. 

## Location

Under cloud-resource-manager/cmd/mex.

This CLI utility can be equivalent of `gcloud` and `az` commands.
In fact, `mex` incorporates them. `mex` can front `gke` and `aks`.
Using `mex` you can tell backend cloud platform to construct
kubernetes cluster.

## Usage

```
$ cd mex
$ go install
$ mex help
Usage of /tmp/go-build838444644/b001/exe/main:
  -debug
        debugging
  -help
        help
  -platform string
        platform data
  -quiet
        less verbose
INFO[0000] e.g. mex [-debug] [-platform pl.yaml] platform {init|clean} -manifest my.yaml
INFO[0000] e.g. mex [-debug] [-platform pl.yaml] cluster {create|remove} -manifest my.yaml
INFO[0000] e.g. mex [-debug] [-platform pl.yaml] application {kill|run} -manifest my.yaml
FATA[0000] insufficient args
exit status 1
```

Note that the -platform pl.yaml provides context for each CLI run. When the MEX API is used by the crmserver, the daemon remembers the context. Because CLI doesn't remember from each run, you need to provide platform context.

## creating platform

Platform creation is only relevant to Openstack based backends.  Gcloud and Azure already have platform running, obviously.

For Openstack, as in Bonn installation, we have to carve out and establish proper settings for hosting kubernetes.  This involves setting up VM instance(s), networking and getting DNS name registered with Cloudflare, getting properly signed certificates from Letsencrypt, running a `mex` agent on the VM instance. The agent is called `mexosagent` and is dockerized.  The agent is built from the directory `edge-cloud-infra/openstack-tenant/agent`.  Doing a `make` in that directory will compile, create docker image, tag it and upload to the private docker registry at `registry.mobiledgex.net:5000`.  Running the `mexosagent` is done by doing a `docker run` referencing the docker image from the registry on the target system.  `mexosagent` performs reverse proxy functions at Layer 7 and this function is programmable via GRPC and REST API.  This is important since our users' apps will be instantiated and the reverse proxy maps have to be updated in realtime.   `mexosagent` also terminates TLS and allows backend apps to work safely without encryption within internal networks.  It also reduces the number of public reserved IP addresses.  Other features exist but not explained here.

Normally you will not create platform often. 

```
$ mex -platform platform-openstack-kubernetes.yaml platform init  -manifest platform-openstack-kubernetes.yaml

```

This takes a while.


## creating kubernetes cluster

On gcloud and azure, the process is simple. They have done all the work. We just use `gcloud` and `az` utilities.  To use `mex` to run relevant `gcloud` and `az` commands you will need to first initialize them.  

* https://cloud.google.com/sdk/install
* https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest

The `mex` utility will talk to backend platform instance running on Bonn cloud (for now) and allow you to create a kubernetes cluster there. 

The idea is to get to a point of kubernetes cluster based API which is portable on all available cloud backend platforms.  Once we have kubernetes we can use kubectl and deployment / service manifests.




## create azure kubernetes cluster

```
$ mex  cluster create -manifest cluster-azure-aks.yaml
#wait for 13.5 minutes
```

## create gcloud kubernetes cluster

```
$ mex cluster create -manifest cluster-gcloud-gke.yaml
#wait for 5 minutes
```

## create openstack kubernetes cluster at Bonn cloudlet
```
$ mex -platform platform-openstack-kubernetes.yaml cluster create -manifest cluster-kubernetes.yaml
#wait for 10 minutes
```


After creating the clusters on each backend platforms, you can use standard `kubectl` tool to launch apps. The kubeconfig credentials retrieved from the cluster and made available. In gke and aks cases, they are in ~/.kube/config.  In the case of openstack kubernetes, they are in ~/.mobiledgex/nameofthecluster-guid.kubeconfig and ~/.mobiledgex/nameofthecluster-guid.kubeconfig-proxy.


## Launching apps

Using kubectl you can launch apps.  But you should use `mex` instead.  There are things provided by `mex` beyond kubectl.  Inside `mex`,  kubectl is used, obviously.  But it performs additional actions. On gcloud and azure, for example, do not have builtin support for pulling from a private registry with proper credentials in a programmatic way.  Nor do they provide different ways of network topology with shared public IP external load balancer.  They cater to certain assumptions and most steps are manually done, even with their APIs.  The `mex` is different because it had to respond to needs of controller which simply indicates desire to run an app and nothing more.  Bunch of things, normally done manually are done automatically and in a way to allow for multiple clusters.

## Launch app into kubernetes via mex

```
$ mex -platform platform-openstack-kubernetes.yaml application run -manifest application-kubernetes.yaml
```

The application-kubernetes.yaml file tells Mex tool details required for the app contexts. Within it, a reference is made to kubernetes manifest that reside on a registry. The kubemainfest  has two parts. First it declares deployment. And service manifest.  The app image is mobiledgex/mexexample docker image from registry.mobiledgex.net:5000.  The code is under cloud-resource-manager/example/mexexample. Doing a make there will build, create docker image, tag it and upload to the registry.  The image is then referenced in the application-kubernetes.yaml file.  Running `mex` to run this manifest file is like doing it with kubectl.  Besides normal kubectl there are a bunch of things being done, such as making sure the secret is created in the kubernetes etcd to be referenced in the application-kubernetes.yaml to allow fetching from private repo.  Setting up proper routing and reverse proxy rules based on the service declarations of ports. The `mexexample` is an app that supports GRPC, REST and HTTP endpoint APIs.  It uses three different ports which are declared and honored by the kubernetes and reverse proxy automatically.  

The application-kubernetes.yaml is mostly custom yaml structures needed by mex. There is a .spec.kubemanifest: pointer which points to a kubernetes manifest called http://registry.mobiledgex.net:8080/mexexample.yaml.  This file contains the  standard kubernetes declarations for deployment and service.  This is served by a server on registry.mobiledgex.net:8080.   It is pulled by kubectl at the target kubernetes cluster.  Before  and after that happens, bunch of things have to happen, and the rest of the yaml content in application-kubernetes.yaml describes details.


Once `mexexample` app is running in the cluster, you can talk to it via proxy.

```
$ curl -XGET https://mexlb.tdg.mobiledgex.net/mexexamplehttp/index.html
hostname mexexample-deployment-56dd4f7d44-9sl5koutbound ip 10.36.0.1{1 65536 lo  up|loopback} [127.0.0.1/8] {12 1376 eth0 d6:bb:12:b6:5f:89 up|broadcast|multicast} [10.36.0.1/12]
```

You can POST

```
$ curl -XPOST https://mexlb.tdg.mobiledgex.net/mexexamplerest/info -d '{"message":"info"}'
{"message":"info","outbound":"10.36.0.1","hostname":"mexexample-deployment-56dd4f7d44-9sl5k","interfaces":[{"name":"lo","addresses":"[127.0.0.1/8]"},{"name":"eth0","addresses":"[10.36.0.1/12]"}]}
```

The `mexexamplerest` and `mexexamplehttp` are at different ports. Another port is proxied with `mexexamplegrpc`.

To see the proxied paths at the reverse proxy `mexosagent` docker app on the rootLB:

```
$ curl http://mexlb.tdg.mobiledgex.net:18889/v1/proxy -H "Content-Type: application/json" -d '{ "message": "list"}'
{"message":"list","status":"map[/mexexamplehttp/*catchall:http://10.101.102.2:27274 /mexexamplerest/*catchall:http://10.101.102.2:27273]"}
```

## using kubectl

```
$ KUBECONFIG=/home/bob/.mobiledgex/mex-k8s-master-1-testPoke-bdgla9hvu8hvb1v74tug.kubeconfig-proxy kubectl get nodes
NAME                                             STATUS    ROLES     AGE       VERSION
mex-k8s-master-1-testpoke-bdgla9hvu8hvb1v74tug   Ready     master    26m       v1.11.1
mex-k8s-node-1-testpoke-bdgla9hvu8hvb1v74tug     Ready     <none>    25m       v1.11.1
mex-k8s-node-2-testpoke-bdgla9hvu8hvb1v74tug     Ready     <none>    25m       v1.11.1
```

```
$ KUBECONFIG=/home/bob/.mobiledgex/mex-k8s-master-1-testPoke-bdgla9hvu8hvb1v74tug.kubeconfig-proxy kubectl get pods
NAME                                     READY     STATUS    RESTARTS   AGE
mexexample-deployment-56dd4f7d44-768nt   1/1       Running   0          20m
mexexample-deployment-56dd4f7d44-9sl5k   1/1       Running   0          20m
```

## launching qcow KVM app

A VM based app can be launched as well

```
$ go run main.go -debug -platform platform-openstack-kubernetes.yaml application run -manifest application-qcow2.yaml
```

The details of this workflow still need to be worked out when apps are made available to us.  As is, not enough information available to see what customization has to be done for a given KVM.


## The -platform flag

The platform flag given before cluster and application sub commands to mex establish baseline platform specifics that are required by cluster and application commands such as create/remove or run/kill.  

When running the MEX API in a CRM talking to controller, the context of the platform created is available.  When running as a CLI, the context is from the yaml file via -platform flag.  




## cleaning up

```
$ source ~/mex.env
$ mex -d mexos -platform platform-openstack-kubernetes.yaml cluster remove -manifest cluster-kubernetes.yaml
$ mex -d mexos -platform platform-openstack-kubernetes.yaml platform clean -manifest platform-openstack-kubernetes.yaml
$ openstack server list
```

