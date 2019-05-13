# Upgrade test files

The upgrade test files are raw output of etcdctl with alternating lines of keys and values. To generate one of these files with preupgrade data you can use the following command:

$ ETCDCTL_API=3 etcdctl get "" --prefix --endpoints=[127.0.0.1:30001,127.0.0.1:30002,127.0.0.1:30003]

where array of endpoints is where your etcd is running. In the above example it's a local set of etcds running after e2e test setup.

Postupgrade file can be generated in the same way.