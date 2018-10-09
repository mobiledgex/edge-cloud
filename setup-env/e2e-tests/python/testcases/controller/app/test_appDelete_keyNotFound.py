#!/usr/local/bin/python3

#
# EDGECLOUD-191 - No error is given when running DeleteApp for an app it cannot find - fixed
#
# delete an app where the key is not found
# verify 'Key not found' error is received
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
image_type = 'ImageTypeDocker'
app_name = 'app' + stamp
app_version = '1.0'
developer_name = 'developer' + stamp
developer_address = 'allen tx'
developer_email = 'dev@dev.com'
flavor = 'x1.small'
cluster_name = 'cluster' + stamp
access_ports = 'tcp:1'

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

        self.developer = mex_controller.Developer(developer_name=developer_name,
                                                  developer_address=developer_address,
                                                  developer_email=developer_email)
        self.cluster = mex_controller.Cluster(cluster_name=cluster_name,
                                              default_flavor_name=flavor)
 
        self.controller.create_developer(self.developer.developer) 
        self.controller.create_cluster(self.cluster.cluster)

    def test_DeleteAppUnknown_noKey(self):
        # print apps before add
        apps_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type=image_type,
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)
        self.controller.create_app(self.app.app)

        # delete app
        error = None
        self.app_delete = mex_controller.App()
        try:
            self.controller.delete_app(self.app_delete.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print developers after add
        apps_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(apps_post)

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Key not found', 'error details')
        expect_equal(len(apps_post), len(apps_pre)+1, 'num developer')
        expect_equal(found_app, True, 'find app')

        assert_expectations()

    def test_DeleteAppUnknown_appNameOnly(self):
        # print apps before add
        apps_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type=image_type,
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # delete app
        error = None
        self.app_delete = mex_controller.App(app_name=app_name)

        try:
            self.controller.delete_app(self.app_delete.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print developers after add
        apps_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(apps_post)

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Key not found', 'error details')
        expect_equal(len(apps_post), len(apps_pre)+1, 'num developer')
        expect_equal(found_app, True, 'find app')

        assert_expectations()


    def test_DeleteAppUnknown_wrongVersion(self):
        # print apps before add
        apps_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type=image_type,
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # delete app
        error = None
        self.app_delete = mex_controller.App(app_name=app_name,
                                             developer_name=developer_name,
                                             app_version="1.1")

        try:
            self.controller.delete_app(self.app_delete.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print developers after add
        apps_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(apps_post)

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Key not found', 'error details')
        expect_equal(len(apps_post), len(apps_pre)+1, 'num developer')
        expect_equal(found_app, True, 'find app')

        assert_expectations()

    def test_DeleteAppUnknown_wrongDeveloperName(self):
        # print apps before add
        apps_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type=image_type,
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # delete app
        error = None
        self.app_delete = mex_controller.App(app_name=app_name,
                                             developer_name=developer_name + 'wrong',
                                             app_version=app_version)

        try:
            self.controller.delete_app(self.app_delete.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print developers after add
        apps_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(apps_post)

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Key not found', 'error details')
        expect_equal(len(apps_post), len(apps_pre)+1, 'num developer')
        expect_equal(found_app, True, 'find app')



    def tearDown(self):
        self.controller.delete_app(self.app.app)
        self.controller.delete_cluster(self.cluster.cluster)
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

