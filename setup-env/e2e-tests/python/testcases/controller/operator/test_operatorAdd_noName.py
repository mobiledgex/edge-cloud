#!/usr/local/bin/python3

#
# create operator with empty name and no name
# verify 'Invalid operator name' is received
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'

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

    def test_createOperatorEmptyName(self):
        # print operators before add
        operator_pre = self.controller.show_operators()

        # create operator
        error = None
        self.operator = mex_controller.Operator(operator_name = '')
        try:
            self.controller.create_operator(self.operator.operator)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print operators after add
        operator_post = self.controller.show_operators()
        
        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Invalid operator name', 'error details')
        expect_equal(len(operator_post), len(operator_pre), 'num operator')

        assert_expectations()

    def test_createOperatorNoName(self):
        # print operators before add
        operator_pre = self.controller.show_operators()

        # create operator
        error = None
        self.operator = mex_controller.Operator()
        try:
            self.controller.create_operator(self.operator.operator)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print operators after add
        operator_post = self.controller.show_operators()
        
        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Invalid operator name', 'error details')
        expect_equal(len(operator_post), len(operator_pre), 'num operator')


if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

