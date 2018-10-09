#!/usr/bin/python3

#
# create operator with same name
# verify it is fails with 'Key already exists'
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'

operator_name = 'operator' + str(time.time())

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

class tc(unittest.TestCase):
    def setUp(self):
        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   )

    def test_createOperator(self):
        # print operators before add
        operator_pre = self.controller.show_operators()

        # create operator
        self.operator = mex_controller.Operator(operator_name = operator_name)
        self.controller.create_operator(self.operator.operator)

        # create operator again
        error = None
        try:
            self.controller.create_operator(self.operator.operator)
        except grpc.RpcError as e:
            logging.info('got exception ' + str(e))
            error = e

        
        # print operators after add
        operator_post = self.controller.show_operators()
        
        # found operator
        found_operator = self.operator.exists(operator_post)

        expect_equal(found_operator, True, 'find operator')
        expect_equal(len(operator_post), len(operator_pre)+1, 'num operator')
        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Key already exists', 'error details')

        assert_expectations()

    def tearDown(self):
        self.controller.delete_operator(self.operator.operator)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

