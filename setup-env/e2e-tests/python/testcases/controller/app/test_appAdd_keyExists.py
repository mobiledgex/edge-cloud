#!/usr/local/bin/python3

#
# create app twice
# verify 'Key already exists' is received
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

stamp = str(time.time())
controller_address = '127.0.0.1:55001'
developer_name = 'developer' + stamp
developer_address = 'allen tx'
developer_email = 'dev@dev.com'
flavor = 'x1.small'
cluster_name = 'cluster' + stamp
app_name = 'app' + stamp
app_version = '1.0'
access_ports = 'tcp:1'

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

        self.developer = mex_controller.Developer(developer_name=developer_name,
                                                  developer_address=developer_address,
                                                  developer_email=developer_email)
        self.cluster = mex_controller.Cluster(cluster_name=cluster_name,
                                              default_flavor_name=flavor)

        self.controller.create_developer(self.developer.developer) 
        self.controller.create_cluster(self.cluster.cluster)

    def test_CreateAppDockerKeyExists(self):
        # print the existing apps 
        app_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # try to add the app again
        err = None
        try:
            resp = self.controller.create_app(self.app.app)
        except grpc.RpcError as e:
            err = e

        expect_equal(err.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(err.details(), 'Key already exists', 'error details')

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(app_post)

        self.controller.delete_app(self.app.app)
        
        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_CreateAppDockerKeyExists_2(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # create the app
        # contains image_type=Docker and no image_path
        app2 = mex_controller.App(image_type='ImageTypeQCOW',
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_ports='tcp:1',
                                      #cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name='x1.tiny')

        # try to add the app again
        err = None
        try:
            resp = self.controller.create_app(app2.app)
        except grpc.RpcError as e:
            err = e

        expect_equal(err.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(err.details(), 'Key already exists', 'error details')

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(app_post)

        self.controller.delete_app(self.app.app)
        
        expect_equal(found_app, True, 'find app')
        assert_expectations()

    @classmethod
    def tearDownClass(self):
        self.controller.delete_cluster(self.cluster.cluster)
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

