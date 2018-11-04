#!/usr/local/bin/python3

#
# show flavor with name only
# verify only 1 flavor is returned
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'

stamp = str(time.time())
flavor_name = 'flavor' + stamp
ram = 1024
disk = 1
vcpus = 1

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

    def test_showFlavor_nameOnly(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor
        self.flavor = mex_controller.Flavor(flavor_name = flavor_name,
                                            ram=ram,
                                            disk=disk,
                                            vcpus=vcpus
        )

        self.controller.create_flavor(self.flavor.flavor)

        # print flavors after add
        flavor_post = self.controller.show_flavors(mex_controller.Flavor(flavor_name = flavor_name).flavor)
        
        # find flavor
        found_flavor = self.flavor.exists(flavor_post)

        expect_equal(found_flavor, True, 'find flavor')
        expect(len(flavor_pre) > 1, 'find flavor count pre')
        expect_equal(len(flavor_post), 1, 'find single flavor count')

        assert_expectations()

    def tearDown(self):
        self.controller.delete_flavor(self.flavor.flavor)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

