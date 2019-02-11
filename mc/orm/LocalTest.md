# Local MC Testing

There are a bunch of services involved with the master controller. The master controller requires a SQL database and Vault for JWT secrets. If you want to test APIs to the controller, then you need etcd and the controller process running (and optionally a CRM).

MC uses REST APIs and JWT auth. For easy command line testing, install httpie and it's jwt support. On mac osx:

```
brew install httpie
pip install -U httpie-jwt-auth
```

For basic auth local testing, run the mc with options to spawn local SQL and Vault processes automatically (note, take off "-initSql" if you want to keep data you've set up when restarting mc):

```
$ mc -localSql -initSql -d api -sqlAddr 127.0.0.1:5445 -localVault
```

If you want to test the controller APIS as well, in separate windows each, run

```
$ etcd
$ controller -d api,notify
$ crmserver -cloudletKey '{"operator_key":{"name":"bigwaves"},"name":"oceanview"}' -d notify,mexos -fakecloudlet
```

## User and Organization Management

At the start, the only user is the super user. We can login using the superuser's default username and password. This will give us a JWT token which we can use for later requests. Put then token in an env var so it's easy to reference and see which user token we're using.

```
$ http POST 127.0.0.1:9900/api/v1/login username=mexadmin password=mexadmin123
HTTP/1.1 200 OK
Content-Length: 232
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:03:06 GMT

{
    "token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDk2NjY5ODYsImlhdCI6MTU0OTU4MDU4NiwidXNlcm5hbWUiOiJtZXhhZG1pbiIsImlkIjoxLCJraWQiOjJ9.eOB3joK6C7JzYFPvfBR-wksMRkXVpcZm5anTA6gbQBhBVtLePWAxRKHccfPAgI1YCXDiO4Uo59REg0ApP6uqXg"
}

$ export SUPERPASS=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDk2NjY5ODYsImlhdCI6MTU0OTU4MDU4NiwidXNlcm5hbWUiOiJtZXhhZG1pbiIsImlkIjoxLCJraWQiOjJ9.eOB3joK6C7JzYFPvfBR-wksMRkXVpcZm5anTA6gbQBhBVtLePWAxRKHccfPAgI1YCXDiO4Uo59REg0ApP6uqXg
```

We can now explore a few things as the super user.

```
$ http --auth-type=jwt --auth=$SUPERPASS POST 127.0.0.1:9900/api/v1/auth/user/current
HTTP/1.1 200 OK
Content-Length: 290
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:04:43 GMT

{
    "CreatedAt": "2019-02-07T15:02:57.617502-08:00",
    "Email": "mexadmin@mobiledgex.net",
    "EmailVerified": true,
    "FamilyName": "mexadmin",
    "GivenName": "mexadmin",
    "ID": 1,
    "Iter": 0,
    "Name": "mexadmin",
    "Nickname": "mexadmin",
    "Passhash": "",
    "Picture": "",
    "Salt": "",
    "UpdatedAt": "2019-02-07T15:02:57.617502-08:00"
}

$ http --auth-type=jwt --auth=$SUPERPASS POST 127.0.0.1:9900/api/v1/auth/role/show
HTTP/1.1 200 OK
Content-Length: 166
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:04:50 GMT

[
    "AdminContributor",
    "AdminManager",
    "AdminViewer",
    "DeveloperContributor",
    "DeveloperManager",
    "DeveloperViewer",
    "OperatorContributor",
    "OperatorManager",
    "OperatorViewer"
]

$ http --auth-type=jwt --auth=$SUPERPASS POST 127.0.0.1:9900/api/v1/auth/org/show
HTTP/1.1 200 OK
Content-Length: 2
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:04:54 GMT

[]

$ http --auth-type=jwt --auth=$SUPERPASS POST 127.0.0.1:9900/api/v1/auth/controller/show
HTTP/1.1 200 OK
Content-Length: 2
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:05:01 GMT

[]

```

Note there are no Organizations or Controllers created yet. The Controllers should be populated by MEX admins, so we can do that now. We can create one that points to the local controller we ran. Organizations however, should be created by the Organization's owner. So we'll do that later.

```
$ http --auth-type=jwt --auth=$SUPERPASS POST 127.0.0.1:9900/api/v1/auth/controller/create region=local address=127.0.0.1:55001
HTTP/1.1 200 OK
Content-Length: 32
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:08:00 GMT

{
    "message": "Controller created"
}
```

Now we're at the point where a Customer can get involved. Each customer can create Organizations, and each Customer can belong to multiple Organizations. The customer that creates the Organization has permission to manage the users in that Organization. So there will primarily be two types of Customers.

Customers who manage the users for their organization will:
1. Create a new user account for themselves
2. Login with their username and password
3. Create an Organization (or two) that they manage users for
4. Add other users to their organization

Customers who work in an organization will:
1. Create a new user account for themselves
2. Request (out of band) to get added to an Organization
3. Login with their username and password
4. Contribute/View the Organization's data (Cloudlets or Clusters/Apps/etc)

Let's start with the Organization Manager (first one above):

```
$ http POST 127.0.0.1:9900/api/v1/usercreate name=orgman passhash=pointyears email="orgman@bigorg.com"
HTTP/1.1 200 OK
Content-Length: 33
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:16:43 GMT

{
    "id": 2,
    "message": "user created"
}

$ http POST 127.0.0.1:9900/api/v1/login username=orgman password=pointyears
HTTP/1.1 200 OK
Content-Length: 230
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:17:12 GMT

{
    "token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDk2Njc4MzIsImlhdCI6MTU0OTU4MTQzMiwidXNlcm5hbWUiOiJvcmdtYW4iLCJpZCI6Miwia2lkIjoyfQ.2tcogi2HHL6UZ1Us6PJuPHWUzoG00581l6n5nR4jN5StLNo8yhXmFspr9c2d5i_V2LKV3A6OvjgjIVyZ0HZI1g"
}

$ export ORGMANTOKEN=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDk2Njc4MzIsImlhdCI6MTU0OTU4MTQzMiwidXNlcm5hbWUiOiJvcmdtYW4iLCJpZCI6Miwia2lkIjoyfQ.2tcogi2HHL6UZ1Us6PJuPHWUzoG00581l6n5nR4jN5StLNo8yhXmFspr9c2d5i_V2LKV3A6OvjgjIVyZ0HZI1g

$ http --auth-type=jwt --auth=$ORGMANTOKEN POST 127.0.0.1:9900/api/v1/auth/org/create name=bigorg type=developer address="123 abc st" phone="123-456-1234"
HTTP/1.1 200 OK
Content-Length: 50
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:18:55 GMT

{
    "message": "Organization created",
    "name": "bigorg"
}
```

Now another user in BigOrg can create an account, and the Manager can add them to BigOrg in a given role. Right now there are only pre-built roles. Managers can do everything - manage users in the organization, read and write all data. Contributors can read/write all data except user membership. Viewers can only read data, and like Contributors have no access to user membership.

```
$ http POST 127.0.0.1:9900/api/v1/usercreate name=worker1 passhash=workinghard email="worker1@bigorg.com"
HTTP/1.1 200 OK
Content-Length: 33
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:22:13 GMT

{
    "id": 3,
    "message": "user created"
}

$ http POST 127.0.0.1:9900/api/v1/login username=worker1 password=workinghard
HTTP/1.1 200 OK
Content-Length: 231
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:22:28 GMT

{
    "token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDk2NjgxNDgsImlhdCI6MTU0OTU4MTc0OCwidXNlcm5hbWUiOiJ3b3JrZXIxIiwiaWQiOjMsImtpZCI6Mn0.Hdl5Imw8NGsCPJ5wfAjrK3XJhoikv807ILZ72Gt64EHjcVWf_xS4jaIin81QFjA6GgAwcX7-Ik6eJRWim2-_xQ"
}

$ export WORKER1TOKEN=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDk2NjgxNDgsImlhdCI6MTU0OTU4MTc0OCwidXNlcm5hbWUiOiJ3b3JrZXIxIiwiaWQiOjMsImtpZCI6Mn0.Hdl5Imw8NGsCPJ5wfAjrK3XJhoikv807ILZ72Gt64EHjcVWf_xS4jaIin81QFjA6GgAwcX7-Ik6eJRWim2-_xQ

$ http --auth-type=jwt --auth=$WORKER1TOKEN POST 127.0.0.1:9900/api/v1/auth/user/current
HTTP/1.1 200 OK
Content-Length: 261
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:25:50 GMT

{
    "CreatedAt": "2019-02-07T15:22:13.169825-08:00",
    "Email": "worker1@bigorg.com",
    "EmailVerified": false,
    "FamilyName": "",
    "GivenName": "",
    "ID": 3,
    "Iter": 0,
    "Name": "worker1",
    "Nickname": "",
    "Passhash": "",
    "Picture": "",
    "Salt": "",
    "UpdatedAt": "2019-02-07T15:22:13.169825-08:00"
}
```

Now the manager can add the new user to the organization. The request takes in the organization name, role name, and user id of the user to add. Because the user id is an integer in the request, we need to pass real json to the http tool.

```
$ http --auth-type=jwt --auth=$ORGMANTOKEN POST 127.0.0.1:9900/api/v1/auth/role/adduser <<< '{"org":"bigorg","userid":3,"role":"DeveloperContributor"}'

HTTP/1.1 200 OK
Content-Length: 32
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:29:59 GMT

{
    "message": "Role added to user"
}
```

Let's create another user with their own org.

```
$ http POST 127.0.0.1:9900/api/v1/usercreate name=beachboy passhash=beachboy email="beachboy@ocean.com"
HTTP/1.1 200 OK
Content-Length: 33
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:33:55 GMT

{
    "id": 4,
    "message": "user created"
}

$ http POST 127.0.0.1:9900/api/v1/login username=beachboy password=beachboy
HTTP/1.1 200 OK
Content-Length: 232
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:34:12 GMT

{
    "token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDk2Njg4NTIsImlhdCI6MTU0OTU4MjQ1MiwidXNlcm5hbWUiOiJiZWFjaGJveSIsImlkIjo0LCJraWQiOjJ9.XfEJ6Fjxlgvg2Hr8iR9IXKfNiL3HqNQFhmZMz7wYKqyS23YwKEP2WTl5zn8hqe2ptpoHdsip_vVhmlz5PlquKg"
}

$ export BBTOKEN=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDk2Njg4NTIsImlhdCI6MTU0OTU4MjQ1MiwidXNlcm5hbWUiOiJiZWFjaGJveSIsImlkIjo0LCJraWQiOjJ9.XfEJ6Fjxlgvg2Hr8iR9IXKfNiL3HqNQFhmZMz7wYKqyS23YwKEP2WTl5zn8hqe2ptpoHdsip_vVhmlz5PlquKg

$ http --auth-type=jwt --auth=$BBTOKEN POST 127.0.0.1:9900/api/v1/auth/org/create name=bigwaves type=operator address="120 ocean st" phone="808-456-1234"
HTTP/1.1 200 OK
Content-Length: 52
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:35:14 GMT

{
    "message": "Organization created",
    "name": "bigwaves"
}
```

Now we can compare what visibility each user has.

```
$ http --auth-type=jwt --auth=$SUPERPASS POST 127.0.0.1:9900/api/v1/auth/org/show
HTTP/1.1 200 OK
Content-Length: 389
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:36:01 GMT

[
    {
        "Address": "123 abc st",
        "AdminUserID": 2,
        "CreatedAt": "2019-02-07T15:18:55.132651-08:00",
        "Name": "bigorg",
        "Phone": "123-456-1234",
        "Type": "developer",
        "UpdatedAt": "2019-02-07T15:18:55.132651-08:00"
    },
    {
        "Address": "120 ocean st",
        "AdminUserID": 4,
        "CreatedAt": "2019-02-07T15:35:14.81412-08:00",
        "Name": "bigwaves",
        "Phone": "808-456-1234",
        "Type": "operator",
        "UpdatedAt": "2019-02-07T15:35:14.81412-08:00"
    }
]

$ http --auth-type=jwt --auth=$ORGMANTOKEN POST 127.0.0.1:9900/api/v1/auth/org/show
HTTP/1.1 200 OK
Content-Length: 194
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:36:10 GMT

[
    {
        "Address": "123 abc st",
        "AdminUserID": 2,
        "CreatedAt": "2019-02-07T15:18:55.132651-08:00",
        "Name": "bigorg",
        "Phone": "123-456-1234",
        "Type": "developer",
        "UpdatedAt": "2019-02-07T15:18:55.132651-08:00"
    }
]

$ http --auth-type=jwt --auth=$WORKER1TOKEN POST 127.0.0.1:9900/api/v1/auth/org/show
HTTP/1.1 200 OK
Content-Length: 194
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:36:18 GMT

[
    {
        "Address": "123 abc st",
        "AdminUserID": 2,
        "CreatedAt": "2019-02-07T15:18:55.132651-08:00",
        "Name": "bigorg",
        "Phone": "123-456-1234",
        "Type": "developer",
        "UpdatedAt": "2019-02-07T15:18:55.132651-08:00"
    }
]

$ http --auth-type=jwt --auth=$BBTOKEN POST 127.0.0.1:9900/api/v1/auth/org/show
HTTP/1.1 200 OK
Content-Length: 196
Content-Type: application/json; charset=UTF-8
Date: Thu, 07 Feb 2019 23:36:23 GMT

[
    {
        "Address": "120 ocean st",
        "AdminUserID": 4,
        "CreatedAt": "2019-02-07T15:35:14.81412-08:00",
        "Name": "bigwaves",
        "Phone": "808-456-1234",
        "Type": "operator",
        "UpdatedAt": "2019-02-07T15:35:14.81412-08:00"
    }
]

## Controller APIs

Controller APIs on the MC are the same as controller REST APIs on the Controller, except that they are wrapped by a region. The region tells the MC which controller address to use when making a request.

Currently we only have one controller tied to the "local" region. Let's create our usual set of data. For this experiment, we'll treat bigwaves as the operator, and bigorg as the developer.

```
http --auth-type=jwt --auth=$SUPERPASS POST 127.0.0.1:9900/api/v1/auth/ctrl/CreateFlavor <<< '{"region":"local","flavor":{"key":{"name":"x1.medium"},"ram":2048,"vcpus":2,"disk":1}}'
http --auth-type=jwt --auth=$SUPERPASS POST 127.0.0.1:9900/api/v1/auth/ctrl/CreateClusterFlavor <<< '{"region":"local","clusterflavor":{"key":{"name":"x1.medium"},"node_flavor":{"name":"x1.medium"},"master_flavor":{"name":"x1.medium"},"num_nodes":2,"max_nodes":2,"num_masters":1}}'
http --auth-type=jwt --auth=$BBTOKEN POST 127.0.0.1:9900/api/v1/auth/ctrl/CreateCloudlet <<< '{"region":"local","cloudlet":{"key":{"operator_key":{"name":"bigwaves"},"name":"oceanview"},"location":{"latitude":1,"longitude":1,"timestamp":{}},"ip_support":2,"num_dynamic_ips":30}}'
http --auth-type=jwt --auth=$ORGMANTOKEN POST 127.0.0.1:9900/api/v1/auth/ctrl/CreateClusterInst <<< '{"region":"local","clusterinst":{"key":{"cluster_key":{"name":"bigclust"},"cloudlet_key":{"operator_key":{"name":"bigwaves"},"name":"oceanview"},"developer":"bigorg"},"flavor":{"name":"x1.medium"}}}'
http --auth-type=jwt --auth=$WORKER1TOKEN POST 127.0.0.1:9900/api/v1/auth/ctrl/CreateApp <<< '{"region":"local","app":{"key":{"developer_key":{"name":"bigorg"},"name":"myapp","version":"1.0.0"},"image_path":"registry.mobiledgex.net:5000/mobiledgex/simapp","image_type":1,"access_ports":"udp:12001,tcp:80,http:7777","default_flavor":{"name":"x1.medium"},"cluster":{"name":"bigclust"},"command":"simapp -port 7777"}}'
http --auth-type=jwt --auth=$WORKER1TOKEN POST 127.0.0.1:9900/api/v1/auth/ctrl/CreateAppInst <<< '{"region":"local","appinst":{"key":{"app_key":{"developer_key":{"name":"bigorg"},"name":"myapp","version":"1.0.0"},"cloudlet_key":{"operator_key":{"name":"bigwaves"},"name":"oceanview"}},"cluster_inst_key":{"cluster_key":{"name":"bigclust"},"cloudlet_key":{"operator_key":{"name":"bigwaves"},"name":"oceanview"}}}}'
```

Notes:
edgectl commands to populate controller directly:
```
edgectl controller CreateCloudlet --key-name oceanview --key-operatorkey-name bigwaves --numdynamicips 30 --location-longitude 1 --location-latitude 1
edgectl controller CreateFlavor --key-name x1.medium --vcpus 2 --ram 2048 --disk 1
edgectl controller CreateClusterFlavor --key-name x1.medium --masterflavor-name x1.medium --maxnodes 2 --nodeflavor-name x1.medium --nummasters 1 --numnodes 2
edgectl controller CreateClusterInst --key-cloudletkey-name oceanview --key-cloudletkey-operatorkey-name bigwaves --key-clusterkey-name bigclust --flavor-name x1.medium
edgectl controller CreateApp --key-developerkey-name bigorg --key-name myapp --key-version 1.0.0 --cluster-name bigclust --defaultflavor-name x1.medium --imagetype ImageTypeDocker --accessports "udp:12001,tcp:80,http:7777" --imagepath "registry.mobiledgex.net:5000/mobiledgex/simapp" --command "simapp -port 7777"
edgectl controller CreateAppInst --key-appkey-developerkey-name bigorg --key-appkey-name myapp --key-appkey-version 1.0.0 --key-cloudletkey-name oceanview --key-cloudletkey-operatorkey-name bigwaves
```
