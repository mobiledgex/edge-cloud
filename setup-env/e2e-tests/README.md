  
Runs e2e tests by running setup-mex multiple times based on a config file and accumlating the output

Usage of e2e-tests:
 -outputdir string
        output directory, timestamp will be appended
  -stop
        stop on failures
  -testfile string
        input file with tests
  -vars string
        optional vars with key=value, separated by comma, e.g. -vars setupdir=setups/test2.yml,var2=somevalue

Variables in -vars are used to substitute entries like {{setupdir}} in the testfile.  This is to make the tests reuseable across setups.
 
Examples:
e2e-tests -testfile setup-env/e2e-tests/testfiles/deploy_add_show.yml -outputdir /tmp/test_out --vars setupdir=setup-env/setups/local_simplex
e2e-tests -testfile setup-env/e2e-tests/testfiles/delete_add.yml -outputdir /tmp/test_out --vars setupdir=setup-env/setups/local_simplex
