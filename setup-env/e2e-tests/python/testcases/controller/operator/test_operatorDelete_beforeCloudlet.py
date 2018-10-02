#!/usr/local/bin/python3

#
# attempt to delete an operator which is in use by a cloudlet
# verify 'Operator in use by Cloudlet' error is received
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

import mex_controller

controller_address = '127.0.0.1:55001'

stamp = str(time.time())
operator_name = 'operator' + stamp
cloudlet_name = 'cloudlet' + stamp

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
        
        self.operator = mex_controller.Operator(operator_name = operator_name)
        self.cloudlet = mex_controller.Cloudlet(cloudlet_name = cloudlet_name,
                                                operator_name = operator_name,
                                                number_of_dynamic_ips = 254)

        self.controller.create_operator(self.operator.operator)
        self.controller.create_cloudlet(self.cloudlet.cloudlet)

    def test_DeleteOperatorUnknown(self):
        # print operators before add
        operator_pre = self.controller.show_operators()

        # create operator
        error = None
        try:
            self.controller.delete_operator(self.operator.operator)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print operators after add
        operator_post = self.controller.show_operators()
        
        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Operator in use by Cloudlet', 'error details')
        expect_equal(len(operator_post), len(operator_pre), 'num operator')

        assert_expectations()

    def tearDown(self):
        self.controller.delete_cloudlet(self.cloudlet.cloudlet)
        time.sleep(1) # wait for cloudlet to be deleted
        self.controller.delete_operator(self.operator.operator)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

