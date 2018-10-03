#!/usr/local/bin/python3

#
# update developer by changing the values
# verify it is updated
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

    def test_updateDeveloperNameOnly(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        self.developer = mex_controller.Developer(developer_name = developer_name)
        self.controller.create_developer(self.developer.developer)

        # update developer
        self.controller.update_developer(self.developer.developer)
        
        # print developers after add
        developer_post = self.controller.show_developers()
        
        # found developer
        found_developer = self.developer.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        assert_expectations()

    def test_updateDeveloperSameParms(self):
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

        # update the developer
        self.controller.update_developer(self.developer.developer)
        
        # print developers after add
        developer_post = self.controller.show_developers()
        
        # found developer
        found_developer = self.developer.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        assert_expectations()

    def test_updateDeveloperAllParms(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        self.developer = mex_controller.Developer(developer_name = developer_name,
                                                  developer_email = developer_email,
                                                  developer_address = developer_address,
                                                  developer_username = developer_username,
                                                  developer_passhash = developer_passhash,
        )
        self.controller.create_developer(self.developer.developer)

        # update the developer
        self.developer_new = mex_controller.Developer(developer_name = developer_name,
                                                      developer_email = developer_email + 'new',
                                                      developer_address = developer_address + 'new',
                                                      developer_username = developer_username + 'new',
                                                      developer_passhash = developer_passhash + 'new',
                                                      include_fields = True
        )        
        self.controller.update_developer(self.developer_new.developer)
        
        # print developers after add
        developer_post = self.controller.show_developers()
        
        # found developer
        found_developer = self.developer_new.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        assert_expectations()

    def test_updateDeveloperEmail(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        self.developer = mex_controller.Developer(developer_name = developer_name,
                                                  developer_email = developer_email,
                                                  developer_address = developer_address,
                                                  developer_username = developer_username,
                                                  developer_passhash = developer_passhash,
        )
        self.controller.create_developer(self.developer.developer)

        # update the developer
        self.developer_new = mex_controller.Developer(developer_name = developer_name,
                                                      developer_email = developer_email + 'new',
                                                      include_fields = True
        )        
        self.controller.update_developer(self.developer_new.developer)
        
        # print developers after add
        developer_post = self.controller.show_developers()
        
        # found developer
        self.developer_search = self.developer_new
        self.developer_search.developer_address = developer_address
        self.developer_search.developer_username = developer_username
        self.developer_search.developer_passhash = developer_passhash
        found_developer = self.developer_search.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        assert_expectations()

    def test_updateDeveloperAddress(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        self.developer = mex_controller.Developer(developer_name = developer_name,
                                                  developer_email = developer_email,
                                                  developer_address = developer_address,
                                                  developer_username = developer_username,
                                                  developer_passhash = developer_passhash,
        )
        self.controller.create_developer(self.developer.developer)

        # update the developer
        self.developer_new = mex_controller.Developer(developer_name = developer_name,
                                                      developer_address = developer_address + 'new',
                                                      include_fields = True
        )        
        self.controller.update_developer(self.developer_new.developer)
        
        # print developers after add
        developer_post = self.controller.show_developers()
        
        # found developer
        self.developer_search = self.developer_new
        self.developer_search.developer_email = developer_email
        self.developer_search.developer_username = developer_username
        self.developer_search.developer_passhash = developer_passhash
        found_developer = self.developer_search.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        assert_expectations()

    def test_updateDeveloperUsername(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        self.developer = mex_controller.Developer(developer_name = developer_name,
                                                  developer_email = developer_email,
                                                  developer_address = developer_address,
                                                  developer_username = developer_username,
                                                  developer_passhash = developer_passhash,
        )
        self.controller.create_developer(self.developer.developer)

        # update the developer
        self.developer_new = mex_controller.Developer(developer_name = developer_name,
                                                      developer_username = developer_username + 'new',
                                                      include_fields = True
        )        
        self.controller.update_developer(self.developer_new.developer)
        
        # print developers after add
        developer_post = self.controller.show_developers()
        
        # found developer
        self.developer_search = self.developer_new
        self.developer_search.developer_address = developer_address
        self.developer_search.developer_email = developer_email
        self.developer_search.developer_passhash = developer_passhash
        found_developer = self.developer_search.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        assert_expectations()

    def test_updateDeveloperPasshash(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        self.developer = mex_controller.Developer(developer_name = developer_name,
                                                  developer_email = developer_email,
                                                  developer_address = developer_address,
                                                  developer_username = developer_username,
                                                  developer_passhash = developer_passhash,
        )
        self.controller.create_developer(self.developer.developer)

        # update the developer
        self.developer_new = mex_controller.Developer(developer_name = developer_name,
                                                      developer_passhash = developer_passhash + 'new',
                                                      include_fields = True
        )        
        self.controller.update_developer(self.developer_new.developer)
        
        # print developers after add
        developer_post = self.controller.show_developers()
        
        # found developer
        self.developer_search = self.developer_new
        self.developer_search.developer_address = developer_address
        self.developer_search.developer_username = developer_username
        self.developer_search.developer_email = developer_email
        found_developer = self.developer_search.exists(developer_post)

        expect_equal(found_developer, True, 'find developer')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

