  
Runs e2e tests by running test-mex multiple times based on a config file and accumlating the output
Usage of e2e-tests:
  -datadir string
        directory where app data files exist (default "$GOPATH/src/github.com/mobiledgex/edge-cloud/setup-env/e2e-tests/data")
  -outputdir string
        output directory, timestamp will be appended (default "/tmp/e2e_test_out")
  -setupfile string
        network config setup file
  -stop
        stop on failures
  -testfile string
        input file with tests
 
Examples:
## individual testfiles
e2e-tests -testfile testfiles/deploy_start.yml -setupfile setups/local_multi.yml
e2e-tests -testfile testfiles/ctrl_restart.yml -setupfile setups/local_multi.yml`
e2e-tests -testfile testfiles/add_show.yml -setupfile setups/local_multi.yml 
e2e-tests -testfile testfiles/find_cloudlet.yml -setupfile setups/local_multi.yml 
e2e-tests -testfile testfiles/verify_loc.yml -setupfile setups/local_multi.yml
e2e-tests -testfile testfiles/fetchlogs.yml -setupfile setups/local_multi.yml 
e2e-tests -testfile testfiles/stop_cleanup.yml -setupfile setups/local_multi.yml 

## test group files
# regression test
e2e-tests -testfile testfiles/regression_group.yml -setupfile setups/local_multi.yml 
# remote deploy and leave things set up
e2e-tests -testfile testfiles/deploy_group.yml -setupfile setups/local_multi.yml

## Garner's SDK test setup
e2e-tests -testfile testfiles/start_add_sdk.yml -setupfile setups/buckhorn_vm1.yml
e2e-tests -testfile testfiles/start_add_sdk.yml -setupfile setups/buckhorn_vm2.yml
