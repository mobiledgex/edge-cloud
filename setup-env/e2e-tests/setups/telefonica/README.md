Overview:

This implements a sample deployment in Open Nenula for Telefonica for demo purposes.  

External load balancing to cluster:

As there is not yet a Mex LB (as in Openstack) nor a public cloud load balancer, a simple
IPVS based L4 load balancer is used.  The setup for this is in add_ipvs.sh.

In the below, 10.95.84.63 is the load balancer IP and 10.95.84.69, 10.95.84.74, 10.95.84.75
are individual servers.  Here we handle traffic on port 80 (sample app) as well as 443 and
50051 (DME) and 55001 (controller)

$ ipvsadm -Ln --stats
IP Virtual Server version 1.2.1 (size=4096)
Prot LocalAddress:Port               Conns   InPkts  OutPkts  InBytes OutBytes
  -> RemoteAddress:Port
TCP  10.95.84.63:80                      6       36       24     2464     2178
  -> 10.95.84.69:80                      2       12        8      825      726
  -> 10.95.84.74:80                      2       12        8      825      726
  -> 10.95.84.75:80                      2       12        8      814      726
TCP  10.95.84.63:443                     0        0        0        0        0
  -> 10.95.84.69:443                     0        0        0        0        0
  -> 10.95.84.74:443                     0        0        0        0        0
  -> 10.95.84.75:443                     0        0        0        0        0
TCP  10.95.84.63:50051                   6       80       60     5426     8750
  -> 10.95.84.69:50051                   1       14       11      935     1516
  -> 10.95.84.74:50051                   2       26       20     1766     2928
  -> 10.95.84.75:50051                   3       40       29     2725     4306
TCP  10.95.84.63:55001                  18      190      145    12635     9778
  -> 10.95.84.69:55001                   4       57       44     3726     3104
  -> 10.95.84.74:55001                   8       77       58     5156     4023
  -> 10.95.84.75:55001                   6       56       43     3753     2651

Deployment of MEX servers:

MEX applications for Controller and DME which need external access are deployed via a DaemonSet.  
This creates one instance per VM and exposes them on their API ports 50051, 55001.  External traffic
can then be passed into any of the instances via the IPVS LB.

Handling of HTTP traffic for L7 Apps:

The nginx ingress controller is used for L7 load balancing internally.  It listens on HTTP port 80 (HTTPS
can be done also but the certs were not added here) and then distributes traffic based on the paths
configured for each app.  Apps can be added without changes to the external IPVS LB.

$ kubectl describe ingress
Name:             ingress-nginx
Namespace:        default
Address:          
Default backend:  default-http-backend:80 (<none>)
Rules:
  Host  Path  Backends
  ----  ----  --------
  *     
        /getdata   mobiledgexsdkdemo-service:80 (<none>)


The application can then be reached from the external LB at 10.95.84.63:80/getdata

The "default-http-backend" pod will return a 404 on any non matched path.

Pod deployment summary:
$ kubectl get pods
NAME                                    READY     STATUS    RESTARTS   AGE
controller-6gp84                        1/1       Running   0          29m
controller-krzmb                        1/1       Running   0          29m
controller-ppzkz                        1/1       Running   0          29m
crmopenneb-69984d89b8-4t5sb             1/1       Running   0          20m
default-http-backend-6586bc58b6-459cs   1/1       Running   0          21h
dme-ggfkh                               1/1       Running   0          23m
dme-h2bzg                               1/1       Running   0          23m
dme-tnqgn                               1/1       Running   0          23m
etcd-operator-649dbdb5cb-whmh9          1/1       Running   14         1d
mex-etcd-cluster-22k42gvdv6             1/1       Running   0          1d
mex-etcd-cluster-lnjvkccsv8             1/1       Running   0          14h
mex-etcd-cluster-p55mnsh46h             1/1       Running   0          1d
mobiledgexsdkdemo-5b86496659-s2c4w      1/1       Running   0          19h
nginx-ingress-controller-4d44f          1/1       Running   0          19h
nginx-ingress-controller-8gn68          1/1       Running   0          19h
nginx-ingress-controller-fd6nc          1/1       Running   2          19h

Note CRM is deployed here as a single instance, although there is nothing automated in this setup
for it to deploy apps.
