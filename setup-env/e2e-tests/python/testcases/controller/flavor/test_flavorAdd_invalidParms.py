#!/usr/local/bin/python3

#
# create flavor with invalid disk/ram/vcpus 
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

    def test_createFlavorRamInvalid(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor
        error = None
        try:
            self.flavor = mex_controller.Flavor(flavor_name=flavor_name, ram='a', vcpus=1, disk=1)
        except TypeError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print flavors after add
        flavor_post = self.controller.show_flavors()

        expect_equal(len(flavor_post), len(flavor_pre), 'num flavor')
        expect_equal(str(error), "'a' has type <class 'str'>, but expected one of: (<class 'int'>,) for field Flavor.ram", 'error code')
        assert_expectations()

    def test_createFlavorVcpusInvalid(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor
        error = None
        try:
            self.flavor = mex_controller.Flavor(flavor_name=flavor_name, ram=1, vcpus='vcpus', disk=2)
        except TypeError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print flavors after add
        flavor_post = self.controller.show_flavors()

        expect_equal(len(flavor_post), len(flavor_pre), 'num flavor')
        expect_equal(str(error), "'vcpus' has type <class 'str'>, but expected one of: (<class 'int'>,) for field Flavor.vcpus", 'error code')
        assert_expectations()

    def test_createFlavorDiskInvalid(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor
        error = None
        try:
            self.flavor = mex_controller.Flavor(flavor_name=flavor_name, ram=1, vcpus=1, disk='disk')
        except TypeError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print flavors after add
        flavor_post = self.controller.show_flavors()

        expect_equal(len(flavor_post), len(flavor_pre), 'num flavor')
        expect_equal(str(error), "'disk' has type <class 'str'>, but expected one of: (<class 'int'>,) for field Flavor.disk", 'error code')
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

