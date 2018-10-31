#!/usr/local/bin/python3

#
# create 100 flavors
# verify all 100 are created
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

number_of_flavors = 100

controller_address = '127.0.0.1:55001'

flavor_name = 'flavor' + str(time.time())
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

        self.flavor_list = []
        for i in range(number_of_flavors):
            self.flavor_list.append(mex_controller.Flavor(flavor_name = flavor_name + str(i),
                                                          ram = ram,
                                                          disk = disk,
                                                          vcpus = vcpus
            ))
            
    def test_createFlavor(self):
        # print flavors before add
        flavor_pre = self.controller.show_flavors()

        # create flavor
        for i in self.flavor_list:
            self.controller.create_flavor(i.flavor)

        # print flavorss after add
        flavor_post = self.controller.show_flavors()
        
        # find flavor in list
        for a in self.flavor_list:
            found_op = a.exists(flavor_post)
            expect_equal(found_op, True, 'find op' + a.flavor_name)

        expect_equal(len(flavor_post), len(flavor_pre) + number_of_flavors, 'number of flavors')

        assert_expectations()

    @classmethod
    def tearDownClass(self):
        for a in self.flavor_list:
            self.controller.delete_flavor(a.flavor)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

