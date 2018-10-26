#!/usr/local/bin/python3

#
# create develpor with name only
# verify it is created
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'

developer_name = 'developer' + str(time.time())

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

    def test_createDeveloper_nameOnly(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        self.developer = mex_controller.Developer(developer_name = developer_name)
        self.controller.create_developer(self.developer.developer)

        # print developers after add
        developer_post = self.controller.show_developers()
        
        # find developer
        found_developer = self.developer.exists(developer_post)

        self.controller.delete_developer(self.developer.developer)

        expect_equal(found_developer, True, 'find developer')
        assert_expectations()

    def test_createDeveloper_nameEmptyParms(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        self.developer = mex_controller.Developer(developer_name = developer_name,
                                                  developer_address = '',
                                                  developer_email = '',
                                                  developer_passhash = '',
                                                  developer_username = '')
        self.controller.create_developer(self.developer.developer)

        # print developers after add
        developer_post = self.controller.show_developers()
        
        # find developer
        found_developer = self.developer.exists(developer_post)

        self.controller.delete_developer(self.developer.developer)

        expect_equal(found_developer, True, 'find developer')
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

