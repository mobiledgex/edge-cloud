# Cloud Resource Manager CLI

The program `crmctl` allows interacting with `crmserver` program which is a
Cloud Resource Manager server.

## Usage

```
$ crmctl -h
Available commands are:
        add-cloud-resource
        list-cloud-resource
        delete-cloud-resource
        deploy-application
        delete-application
```

Each of the sub-command, such as `delete-application` can have flags.
To list flags per command pass `-h` to the sub-command. For example,

```
$ crmctl delete-cloud-resource -hUsage of delete-cloud-resource:
  -address string
        Address of the cloudlet (required)
  -crm string
        Address of Cloud Resource Manager (required)
  -location string
        Location of the cloudlet (required)
  -opkey string
        Operator Key for the cloudlet (required)
  -opkeyname string
        Operator Key Name for the cloudlet (required)
```


## List cloud resources

```
$ crmctl list-cloud-resource -crm 127.0.0.1:55099
```

## Add a cloud resource

```
$ crmctl add-cloud-resource -crm 127.0.0.1:55099 -address 3.3.3.3:999 -location london -opkey asdf -opkeyname aaa -name abc
```


## Delete a cloud resource

```
$ crmctl delete-cloud-resource -crm 127.0.0.1:55099 -address 3.3.3.3:999 -location london -opkey asdf -opkeyname aaa -name abc
```

## Run application

Applications can be deployed via a manifest file which depends on the type of application.  For Kubernetes applications there are two types supported.  

* k8s-manifest
* k8s-simple

The `k8s-manifest` type requires Kubernets style YAML (or JSON) file which contain Deployments and possibly also Services and other objects specified.

The `k8s-simple` type requires a list of parameters. The Deployment structure is created by the CRM server and passed to the Kubernetes based on the parameters. This is for very simple applications only.

```
```


## Delete Application


```
```
