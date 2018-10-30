export PYTHONPATH=./setup-env/e2e-tests/python/protos:./setup-env/e2e-tests/python/modules:./setup-env/e2e-tests/python/certs
e2e-tests -setupfile ./setup-env/e2e-tests/setups/local_multi_automation.yml -testfile ./setup-env/e2e-tests/testfiles/stop_cleanup.yml
e2e-tests -setupfile ./setup-env/e2e-tests/setups/local_multi_automation.yml -testfile ./setup-env/e2e-tests/testfiles/deploy_start_create_automation.yml
python3 -m unittest discover -s ./setup-env/e2e-tests/python/testcases/$dir
