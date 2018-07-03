  
Runs e2e tests by running test-mex multiple times based on a config file and accumlating the output

Usage of e2e-tests:
 -outputdir string
        output directory, timestamp will be appended
  -stop
        stop on failures
  -testfile string
        input file with tests
  -outputdir string
        directory to store the results and logs
  -setupfile string
        file used to define the network setup
  -datadir string
        directory with data files such as those used for APIs

 
Examples:
e2e-tests -testfile testfiles/deploy_start.yml -outputdir /tmp/test_results -setupfile setups/local_multi.yml -datadir data
e2e-tests -testfile testfiles/ctrl_restart.yml -outputdir /tmp/test_results -setupfile setups/local_multi.yml -datadir data
e2e-tests -testfile testfiles/add_show.yml -outputdir /tmp/test_results -setupfile setups/local_multi.yml -datadir data
e2e-tests -testfile testfiles/find_cloudlet.yml -outputdir /tmp/test_results -setupfile setups/local_multi.yml -datadir data
e2e-tests -testfile testfiles/verify_loc.yml -outputdir /tmp/test_results -setupfile setups/local_multi.yml -datadir data
e2e-tests -testfile testfiles/fetchlogs.yml -outputdir /tmp/test_results -setupfile setups/local_multi.yml -datadir data
e2e-tests -testfile testfiles/stop_cleanup.yml -outputdir /tmp/test_results -setupfile setups/local_multi.yml -datadir data
