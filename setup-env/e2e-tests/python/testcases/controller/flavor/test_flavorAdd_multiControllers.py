#!/usr/local/bin/python3

#
# create flavor on different controller and check the other controller
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
flavor_name_1 = 'flavor_1 ' + stamp
flavor_name_2 = 'flavor_2 ' + stamp
ram = 1
vcpus = 2
disk = 3

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

class tc(unittest.TestCase):
    @classmethod
    def setUpClass(self):
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

    def test_createFlavor(self):
        # print flavors before add
        flavor_pre_1 = self.controller_1.show_flavors()
        flavor_pre_2 = self.controller_2.show_flavors()

        # create flavor
        self.flavor_1 = mex_controller.Flavor(flavor_name = flavor_name_1,
                                              ram = ram,
                                              vcpus = vcpus,
                                              disk = disk,
        )
        self.flavor_2 = mex_controller.Flavor(flavor_name = flavor_name_2,
                                              ram = ram,
                                              vcpus = vcpus,
                                              disk = disk,

        )
        self.controller_1.create_flavor(self.flavor_1.flavor)
        self.controller_2.create_flavor(self.flavor_2.flavor)

        # print flavors after add
        flavor_post_1 = self.controller_1.show_flavors()
        flavor_post_2 = self.controller_2.show_flavors()
        
        # found flavor
        found_flavor_11 = self.flavor_1.exists(flavor_post_1)
        found_flavor_12 = self.flavor_1.exists(flavor_post_2)
        found_flavor_21 = self.flavor_2.exists(flavor_post_1)
        found_flavor_22 = self.flavor_2.exists(flavor_post_2)

        # delete flavors
        self.controller_1.delete_flavor(self.flavor_2.flavor)
        self.controller_2.delete_flavor(self.flavor_1.flavor)

        # print flavors after delete
        flavor_post_1_2 = self.controller_1.show_flavors()
        flavor_post_2_2 = self.controller_2.show_flavors()

        # verify flavors dont exist after delete
        found_flavor_11_2 = self.flavor_1.exists(flavor_post_1_2)
        found_flavor_12_2 = self.flavor_1.exists(flavor_post_2_2)
        found_flavor_21_2 = self.flavor_2.exists(flavor_post_1_2)
        found_flavor_22_2 = self.flavor_2.exists(flavor_post_2_2)
        
        expect_equal(found_flavor_11, True, 'find flavor 11')
        expect_equal(found_flavor_12, True, 'find flavor 12')
        expect_equal(found_flavor_21, True, 'find flavor 21')
        expect_equal(found_flavor_22, True, 'find flavor 22')

        expect_equal(found_flavor_11_2, False, 'find flavor 11 after delete')
        expect_equal(found_flavor_12_2, False, 'find flavor 12 after delete')
        expect_equal(found_flavor_21_2, False, 'find flavor 21 after delete')
        expect_equal(found_flavor_22_2, False, 'find flavor 22 after delete')

        
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

