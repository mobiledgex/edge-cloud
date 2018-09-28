#!/usr/bin/python3

#
# create app with image_type=ImageTypeDocker and empty/no imagepath
# verify image_path='mobiledgex_' 
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
developer_name = 'developer' + stamp
developer_address = 'allen tx'
developer_email = 'dev@dev.com'
flavor = 'x1.small'
cluster_name = 'cluster' + stamp
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
        self.cluster = mex_controller.Cluster(cluster_name=cluster_name,
                                              default_flavor_name=flavor)

        # create the app
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_layer='AccessLayerL7',
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)

        self.controller.create_developer(self.developer.developer) 
        self.controller.create_cluster(self.cluster.cluster)
        self.controller.create_app(self.app.app)


    def test_QueryAppName(self):
        # print the existing apps 
        app = mex_controller.App(app_name=app_name)
        app_show = self.controller.show_apps(app.app)

        # find app in list
        found_app = self.app.exists(app_show)

        expect_equal(len(app_show), 1, 'number of apps')
        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_QueryDeveloperName(self):
        # print the existing apps
        app = mex_controller.App(developer_name=developer_name)
        app_show = self.controller.show_apps(app.app)

        # find app in list
        found_app = self.app.exists(app_show)

        expect_equal(len(app_show), 1, 'number of apps')
        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_QueryAppNameVersion(self):
        # print the existing apps
        app = mex_controller.App(app_name=app_name, app_version=app_version)
        app_show = self.controller.show_apps(app.app)

        # find app in list
        found_app = self.app.exists(app_show)

        expect_equal(len(app_show), 1, 'number of apps')
        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_app(self.app.app)
        self.controller.delete_cluster(self.cluster.cluster)
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

