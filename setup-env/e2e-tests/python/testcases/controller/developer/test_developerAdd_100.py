#!/usr/local/bin/python3

#
# create 100 developers
# verify all 100 are created
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

number_of_developers = 100

controller_address = '127.0.0.1:55001'

developer_name = 'developer' + str(time.time())
developer_address = '502 creekside ln, Allen, TX 75002'
developer_email = 'tester@automation.com'
developer_username = 'username'
developer_passhash = 'sdfasfadfafasfafafafafaeffsdffasfafafafadafafafafdafafafaerqwerqwrasfasfasf'

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

        self.developer_list = []
        for i in range(number_of_developers):
            self.developer_list.append(mex_controller.Developer(developer_name = developer_name + str(i),
                                                                developer_email = developer_email,
                                                                developer_address = developer_address,
                                                                developer_username = developer_email,
                                                                developer_passhash = developer_passhash
            ))
            
    def test_createDeveloper(self):
        # print developers before add
        developer_pre = self.controller.show_developers()

        # create developer
        for i in self.developer_list:
            self.controller.create_developer(i.developer)

        # print developerss after add
        developer_post = self.controller.show_developers()
        
        # find developer in list
        for a in self.developer_list:
            found_op = a.exists(developer_post)
            expect_equal(found_op, True, 'find op' + a.developer_name)

        expect_equal(len(developer_post), len(developer_pre) + number_of_developers, 'number of developers')

        assert_expectations()

    @classmethod
    def tearDownClass(self):
        for a in self.developer_list:
            self.controller.delete_developer(a.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

