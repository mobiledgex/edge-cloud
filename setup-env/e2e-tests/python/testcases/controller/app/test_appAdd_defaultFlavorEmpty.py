#!/usr/bin/python3

#
# create app with default empty and missing 
# verify '"Specified default flavor not found"' is received
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

sys.path.append('/root/andy/python/protos')

import mex_controller

stamp = str(time.time())

controller_address = '127.0.0.1:55001'
developer_name = 'mr. developer' + stamp
developer_address = 'allen tx'
developer_email = 'dev@dev.com'
app_name = 'app' + stamp
app_version = '1.0'

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

        self.developer = mex_controller.Developer(developer_name=developer_name,
                                                  address=developer_address,
                                                  email=developer_email)
        self.controller.create_developer(self.developer.developer) 
    def test_CreateAppDefaultFlavorEmpty(self):
        # print the existing apps 
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 app_name=app_name,
                                 app_version=app_version,
                                 cluster_name='dummyCluster',
                                 developer_name=developer_name,
                                 default_flavor_name='')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Specified default flavor not found', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppDefaultFlavorEmpty_2(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeQCOW',
                                 app_name=app_name,
                                 app_version=app_version,
                                 cluster_name='dummyCluster',
                                 developer_name=developer_name,
                                 default_flavor_name='')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Specified default flavor not found', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppDefaultFlavorNotExist(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 app_name=app_name,
                                 app_version=app_version,
                                 cluster_name='dummyCluster',
                                 developer_name=developer_name,
                                 )
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Specified default flavor not found', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppDefaultFlavorNotExist_2(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeQCOW',
                                 app_name=app_name,
                                 app_version=app_version,
                                 cluster_name='dummyCluster',
                                 developer_name=developer_name,
                                 )
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Specified default flavor not found', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

