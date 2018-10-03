#!/usr/local/bin/python3

#
# create developer on different controller and check the other controller
# verify it is created on both controllers
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

import mex_controller

controller_address_1 = '127.0.0.1:55001'
controller_address_2 = '127.0.0.1:55002'

stamp = str(time.time())
developer_name_1 = 'developer_1 ' + stamp
developer_name_2 = 'developer_2 ' + stamp
developer_address = '502 creekside ln, Allen, TX 75002'
developer_email = 'tester@automation.com'
developer_username = 'username'
developer_passhash = 'sdfasfadfafasfafafafafaeffsdffasfafafafadafafafafdafafafaerqwerqwrasfasfasf'

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

class tc(unittest.TestCase):
    def setUp(self):
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

    def test_createDeveloper(self):
        # print developers before add
        developer_pre_1 = self.controller_1.show_developers()
        developer_pre_2 = self.controller_2.show_developers()

        # create developer
        self.developer_1 = mex_controller.Developer(developer_name = developer_name_1,
                                                    developer_email = developer_email,
                                                    developer_address = developer_address,
                                                    developer_username = developer_email,
                                                    developer_passhash = developer_passhash
        )
        self.developer_2 = mex_controller.Developer(developer_name = developer_name_2,
                                                    developer_email = developer_email,
                                                    developer_address = developer_address,
                                                    developer_username = developer_email,
                                                    developer_passhash = developer_passhash
        )
        self.controller_1.create_developer(self.developer_1.developer)
        self.controller_2.create_developer(self.developer_2.developer)

        # print developers after add
        developer_post_1 = self.controller_1.show_developers()
        developer_post_2 = self.controller_2.show_developers()
        
        # found developer
        found_developer_11 = self.developer_1.exists(developer_post_1)
        found_developer_12 = self.developer_1.exists(developer_post_2)
        found_developer_21 = self.developer_2.exists(developer_post_1)
        found_developer_22 = self.developer_2.exists(developer_post_2)

        # delete developers
        self.controller_1.delete_developer(self.developer_2.developer)
        self.controller_2.delete_developer(self.developer_1.developer)

        # print developers after delete
        developer_post_1_2 = self.controller_1.show_developers()
        developer_post_2_2 = self.controller_2.show_developers()

        # verify developers dont exist after delete
        found_developer_11_2 = self.developer_1.exists(developer_post_1_2)
        found_developer_12_2 = self.developer_1.exists(developer_post_2_2)
        found_developer_21_2 = self.developer_2.exists(developer_post_1_2)
        found_developer_22_2 = self.developer_2.exists(developer_post_2_2)
        
        expect_equal(found_developer_11, True, 'find developer 11')
        expect_equal(found_developer_12, True, 'find developer 12')
        expect_equal(found_developer_21, True, 'find developer 21')
        expect_equal(found_developer_22, True, 'find developer 22')

        expect_equal(found_developer_11_2, False, 'find developer 11 after delete')
        expect_equal(found_developer_12_2, False, 'find developer 12 after delete')
        expect_equal(found_developer_21_2, False, 'find developer 21 after delete')
        expect_equal(found_developer_22_2, False, 'find developer 22 after delete')

        
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

