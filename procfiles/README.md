# System testing

A procfile is used to describe a set of processes to run. A full MEX edge-cloud will consist of etcd processes, controller processes, distributed matching engine processes, and cloudlet manager processes. Externally there may also be app clients and app servers.

A real test involves each of these processes in their own Docker container, but for testing, it is faster to run each in the same global zone, instead manipulating their config/data directories and ports so that they can interact with each other. The procfile defines these processes and ports.

There are a lot of different tools that can be used to run Procfiles, but the best I've found is ["invoker"](http://invoker.codemancers.com/). On OSX, you can install via:

`sudo gem install invoker`

Then run

`invoker start <Procfile>`

to run all the processes. Processes can also be started/stopped individually. My hope here is the test framework can switch between invoker and docker with only a small translation layer, since they effectively do the same thing (spawn and control processes).
