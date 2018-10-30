#!/usr/local/bin/python3

#
# create 100 operators
# verify all 100 are created
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

number_of_operators = 100

controller_address = '127.0.0.1:55001'

operator_name = 'operator' + str(time.time())

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

class tc(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   )

        self.operator_list = []
        for i in range(number_of_operators):
            self.operator_list.append(mex_controller.Operator(operator_name = 'operator ' + str(i)))
            
    def test_createOperator(self):
        # print operators before add
        operator_pre = self.controller.show_operators()

        # create operator
        for i in self.operator_list:
            self.controller.create_operator(i.operator)

        # print operators after add
        operator_post = self.controller.show_operators()
        
        # find operator in list
        for a in self.operator_list:
            found_op = a.exists(operator_post)
            expect_equal(found_op, True, 'find op' + a.operator_name)

        expect_equal(len(operator_post), len(operator_pre) + number_of_operators, 'number of operators')

        assert_expectations()

    @classmethod
    def tearDownClass(self):
        for a in self.operator_list:
            self.controller.delete_operator(a.operator)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

