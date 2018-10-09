#!/usr/local/bin/python3

#EDGECLOUD-192 - fixed
#create an app with app name this is not docker compliant
#verify imagename is converted to docker compliant  - compiancy is checked in the mex_controller module

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

    def test_CreateNameSpace(self):
        # print the existing apps 
        apps_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type=image_type,
                                      app_name='andy dandy',
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)

        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        apps_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(apps_post)

        expect_equal(found_app, True, 'find app')
        expect_equal(len(apps_post), len(apps_pre)+1, 'num developer')
                
        assert_expectations()

    def test_CreateAndSymbol(self):
        # print the existing apps 
        apps_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type=image_type,
                                      app_name='andy&dandy',
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)

        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        apps_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(apps_post)

        expect_equal(found_app, True, 'find app')
        expect_equal(len(apps_post), len(apps_pre)+1, 'num developer')
                
        assert_expectations()

    def test_CreateComma(self):
        # print the existing apps 
        apps_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type=image_type,
                                      app_name='andy,dandy',
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)

        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        apps_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(apps_post)

        expect_equal(found_app, True, 'find app')
        expect_equal(len(apps_post), len(apps_pre)+1, 'num developer')
                
        assert_expectations()

    def test_CreateBang(self):
        # print the existing apps 
        apps_pre = self.controller.show_apps()

        # create the app
        self.app = mex_controller.App(image_type=image_type,
                                      app_name='andy!dandy',
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)

        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        apps_post = self.controller.show_apps()

        # find app in list
        found_app = self.app.exists(apps_post)

        expect_equal(found_app, True, 'find app')
        expect_equal(len(apps_post), len(apps_pre)+1, 'num developer')
                
        assert_expectations()

    def tearDown(self):
        self.controller.delete_app(self.app.app)
        self.controller.delete_cluster(self.cluster.cluster)
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

