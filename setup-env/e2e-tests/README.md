  
Runs e2e tests by running setup-mex multiple times based on a config file and accumlating the output

Usage of e2e-tests:
  -outputdir string
        output directory, timestamp will be appended
  -stop
        stop on failures
  -testfile string
        input file with tests

Example:
e2e-tests -testfile testfiles/gcp_e2e_test.yml  -outputdir /tmp/test_out

