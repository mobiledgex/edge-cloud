#!/usr/local/bin/python3

#
# create operator on different controller and check the other controller
# verify it is created on both controllers
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address_1 = '127.0.0.1:55001'
controller_address_2 = '127.0.0.1:55002'

stamp = str(time.time())
operator_name_1 = 'operator_1 ' + stamp
operator_name_2 = 'operator_2 ' + stamp

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

class tc(unittest.TestCase):
    def setUp(self):
        self.controller_1 = mex_controller.Controller(controller_address = controller_address_1,
                                                      root_cert = mex_root_cert,
                                                      key = mex_key,
                                                      client_cert = mex_cert
        )

        self.controller_2 = mex_controller.Controller(controller_address = controller_address_2,
                                                      root_cert = mex_root_cert,
                                                      key = mex_key,
                                                      client_cert = mex_cert
        )

    def test_createOperator(self):
        # print operators before add
        operator_pre_1 = self.controller_1.show_operators()
        operator_pre_2 = self.controller_2.show_operators()

        # create operator
        self.operator_1 = mex_controller.Operator(operator_name = operator_name_1)
        self.operator_2 = mex_controller.Operator(operator_name = operator_name_2)
        self.controller_1.create_operator(self.operator_1.operator)
        self.controller_2.create_operator(self.operator_2.operator)

        # print operators after add
        operator_post_1 = self.controller_1.show_operators()
        operator_post_2 = self.controller_2.show_operators()
        
        # found operator
        found_operator_11 = self.operator_1.exists(operator_post_1)
        found_operator_12 = self.operator_1.exists(operator_post_2)
        found_operator_21 = self.operator_2.exists(operator_post_1)
        found_operator_22 = self.operator_2.exists(operator_post_2)

        # delete operators
        #time.sleep(1)
        self.controller_1.delete_operator(self.operator_2.operator)
        self.controller_2.delete_operator(self.operator_1.operator)

        # print operators after delete
        operator_post_1_2 = self.controller_1.show_operators()
        operator_post_2_2 = self.controller_2.show_operators()

        # verify operators dont exist after delete
        found_operator_11_2 = self.operator_1.exists(operator_post_1_2)
        found_operator_12_2 = self.operator_1.exists(operator_post_2_2)
        found_operator_21_2 = self.operator_2.exists(operator_post_1_2)
        found_operator_22_2 = self.operator_2.exists(operator_post_2_2)
        
        expect_equal(found_operator_11, True, 'find operator 11')
        expect_equal(found_operator_12, True, 'find operator 12')
        expect_equal(found_operator_21, True, 'find operator 21')
        expect_equal(found_operator_22, True, 'find operator 22')

        expect_equal(found_operator_11_2, False, 'find operator 11 after delete')
        expect_equal(found_operator_12_2, False, 'find operator 12 after delete')
        expect_equal(found_operator_21_2, False, 'find operator 21 after delete')
        expect_equal(found_operator_22_2, False, 'find operator 22 after delete')

        
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

