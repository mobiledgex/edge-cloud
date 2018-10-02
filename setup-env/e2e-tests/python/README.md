### Setup the environment
Install python3 and needed modules:
```sh
$ brew install python
$ pip3 install grpcio grpcio-tools googleapis-common-protos
```

Export environment variable to find protos, modules, and certificates
```sh
$ export PYTHONPATH=~/go/src/github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/python/protos:~/go/src/github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/python/modules:~/go/src/github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/python/certs
```

Start the edge processes
```sh
$ e2e-tests -testfile deploy_start_create.yml  -setupfile local_multi.yml -stop
```

### Run the testcases
There is a seperate directory for each testcase type. This is the structure
    
    |--controller  - contains controller testcases
    |  |--cluster  - cluster testcases
    |  |--app      - app testcases
    |--dme         - dme testcases
       |--cloudlet - cloudlet testcases 
       |--location - location testcases

Execute 'python3 -m unittest' from any of theses directories to run the tests in that directory and its subdirectories.
* execute 'python3 -m unittest' from the toplevel directory to execute all tests in controller and dme
* execute 'python3 -m unittest' from the controller directory to execute all tests in cluster/app/location
* execute 'python3 -m unittest' from cluster directory to execute only cluster testcases 
* execute 'python3 -m unittest <filename>' or 'python3 <filename>' to execute all tests in that file
* execute 'python3 -m unittest <filename(without extension)>.tc.<testname> to run the single testcase from the file\
          'python3 -m unittest test_clusterInstAdd.tc.test_AddClusterInstance' - where test_AddClusterInstance exists in test_clusterInstAdd.py
