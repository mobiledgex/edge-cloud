#!/usr/local/bin/python3

#
# show controllers by address
# verify controllers is shown
#

import unittest
import sys
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'

controller_address_list = ['0.0.0.0:55001', '0.0.0.0:55002']

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

    def test_showControllersAddr(self):
        # show controllers
        for addr in controller_address_list:
            resp = self.controller.show_controllers(address=addr)

            expect_equal(len(resp), 1, 'number of controllers')
            expect_equal(resp[0].key.addr, addr, 'addr 1')

        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

