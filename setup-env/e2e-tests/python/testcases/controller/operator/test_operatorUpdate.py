#!/usr/local/bin/python3

#
# update operator
# verify it is updated - doesnt really do anything. maybe for future
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

import mex_controller

controller_address = '127.0.0.1:55001'

operator_name = 'operator' + str(time.time())

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

class tc(unittest.TestCase):
    def setUp(self):
        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   )

    def test_updateOperator(self):
        # print operators before add
        operator_pre = self.controller.show_operators()

        # create operator
        self.operator = mex_controller.Operator(operator_name = operator_name)
        self.controller.create_operator(self.operator.operator)

        # update the operator
        self.controller.update_operator(self.operator.operator)

        # print operators after add
        operator_post = self.controller.show_operators()
        
        # found operator
        found_operator = self.operator.exists(operator_post)

        expect_equal(found_operator, True, 'find operator')
        assert_expectations()

    def test_updateOperatorSpace(self):
        # print operators before add
        operator_pre = self.controller.show_operators()

        # create operator
        self.operator = mex_controller.Operator(operator_name = operator_name + ' operator')
        self.controller.create_operator(self.operator.operator)

        # update the operator
        self.controller.update_operator(self.operator.operator)
        
        # print operators after add
        operator_post = self.controller.show_operators()
        
        # found operator
        found_operator = self.operator.exists(operator_post)

        expect_equal(found_operator, True, 'find operator')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_operator(self.operator.operator)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

