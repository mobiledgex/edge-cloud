#!/usr/local/bin/python3

#
# show developer with name only
# verify only 1 developer is returned
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
developer_name = 'developer' + stamp
developer_address = '502 creekside ln, Allen, TX 75002' + stamp
developer_email = 'tester@automation.com' + stamp
developer_username = 'username' + stamp
developer_passhash = 'sdfasfadfafasfafafafafaeffsdffasfafafafadafafafafdafafafaerqwerqwrasfasfasf' + stamp

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

class tc(unittest.TestCase):
    def setUp(self):
        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   )

    def test_showDeveloper_nameOnly(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        self.developer = mex_controller.Developer(developer_name = developer_name,
                                                  developer_email = developer_email,
                                                  developer_address = developer_address,
                                                  developer_username = developer_email,
                                                  developer_passhash = developer_passhash,
        )
        self.controller.create_developer(self.developer.developer)

        # print developers after add
        developer_post = self.controller.show_developers(mex_controller.Developer(developer_name = developer_name).developer)
        
        # find developer
        found_developer = self.developer.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        expect(len(developer_pre) > 1, 'find developer count pre')
        expect_equal(len(developer_post), 1, 'find single developer count')

        assert_expectations()

    def tearDown(self):
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

