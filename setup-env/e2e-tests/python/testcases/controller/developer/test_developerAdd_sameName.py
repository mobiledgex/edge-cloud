#!/usr/local/bin/python3

#
# attempt to create developer with same name
# verify 'Key already exists' is reveived
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

import mex_controller

controller_address = '127.0.0.1:55001'

developer_name = 'developer' + str(time.time())
developer_address = '502 creekside ln, Allen, TX 75002'
developer_email = 'tester@automation.com'
developer_username = 'username'
developer_passhash = 'sdfasfadfafasfafafafafaeffsdffasfafafafadafafafafdafafafaerqwerqwrasfasfasf'

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

    def test_createDeveloper_sameName_allOptional(self):
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

        # create same developer
        error = None
        try:
            self.controller.create_developer(self.developer.developer)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print developers after add
        developer_post = self.controller.show_developers()
        
        # find developer
        found_developer = self.developer.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        expect_equal(len(developer_post), len(developer_pre)+1, 'num developer')
        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Key already exists', 'error details')

        assert_expectations()

    def test_createDeveloper_sameName_nameOnly(self):
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

        # create same developer
        error = None
        try:
            self.controller.create_developer(mex_controller.Developer(developer_name = developer_name).developer)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print developers after add
        developer_post = self.controller.show_developers()
        
        # find developer
        found_developer = self.developer.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        expect_equal(len(developer_post), len(developer_pre)+1, 'num developer')
        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Key already exists', 'error details')

        assert_expectations()

    def tearDown(self):
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

