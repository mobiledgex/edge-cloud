#!/usr/local/bin/python3

#
# create flavor with ram=0 
# verify 'Ram cannot be 0' is received
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
    @classmethod
    def setUpClass(self):
        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   )

    def test_createFlavorRamOnly(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor 
        error = None
        self.flavor = mex_controller.Flavor(flavor_name = 'myflavor', ram = 0)
        try:
            self.controller.create_flavor(self.flavor.flavor)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print flavors after add
        flavor_post = self.controller.show_flavors()
        
        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Ram cannot be 0', 'error details')
        expect_equal(len(flavor_post), len(flavor_pre), 'num flavor')

        assert_expectations()

    def test_createFlavorRam0OtherParms(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor
        error = None
        self.flavor = mex_controller.Flavor(flavor_name='myflavor', ram=0, vcpus=1, disk=1)
        try:
            self.controller.create_flavor(self.flavor.flavor)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print flavors after add
        flavor_post = self.controller.show_flavors()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Ram cannot be 0', 'error details')
        expect_equal(len(flavor_post), len(flavor_pre), 'num flavor')

        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

