#!/usr/local/bin/python3

#
# create flavor with values set to greater than largest int size 
# verify flavor is not added
# 

import unittest
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'

flavor_name = 'flavor' + str(int(time.time()))
ram = 18446744073709551615  # unit64 64-bit integers (0 to 18446744073709551615) 
disk = 18446744073709551615
vcpus = 18446744073709551615 

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

    def test_createFlavorRamTooLarge(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor
        error = None
        try:
            self.flavor = mex_controller.Flavor(flavor_name=flavor_name, ram=ram+1, vcpus=vcpus, disk=disk)
        except ValueError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print flavors after add
        flavor_post = self.controller.show_flavors()

        expect_equal(len(flavor_post), len(flavor_pre), 'num flavor')
        expect_equal(str(error), 'Value out of range: ' + str(ram+1), 'error code')
        assert_expectations()

    def test_createFlavorVcpusTooLarge(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor
        error = None
        try:
            self.flavor = mex_controller.Flavor(flavor_name=flavor_name, ram=ram, vcpus=vcpus+1, disk=disk)
        except ValueError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print flavors after add
        flavor_post = self.controller.show_flavors()

        expect_equal(len(flavor_post), len(flavor_pre), 'num flavor')
        expect_equal(str(error), 'Value out of range: ' + str(vcpus+1), 'error code')
        assert_expectations()

    def test_createFlavorDiskTooLarge(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor
        error = None
        try:
            self.flavor = mex_controller.Flavor(flavor_name=flavor_name, ram=ram, vcpus=vcpus, disk=disk+1)
        except ValueError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print flavors after add
        flavor_post = self.controller.show_flavors()

        expect_equal(len(flavor_post), len(flavor_pre), 'num flavor')
        expect_equal(str(error), 'Value out of range: ' + str(disk+1), 'error code')
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

