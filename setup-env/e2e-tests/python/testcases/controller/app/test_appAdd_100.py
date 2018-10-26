#!/usr/local/bin/python3

#
# create app with image_type=ImageTypeDocker and empty/no imagepath
# verify image_path='mobiledgex_' 
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

sys.path.append('/root/andy/python/protos')

import mex_controller

number_of_apps = 100

stamp = str(time.time())
controller_address = '127.0.0.1:55001'
developer_name = 'developer' + stamp
developer_address = 'allen tx'
developer_email = 'dev@dev.com'
flavor = 'x1.small'
cluster_name = 'cluster' + stamp
app_name = 'app' + stamp
app_version = '1.0'
ip_access = 'IpAccessDedicatedOrShared'
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

        self.app_list = []
        for i in range(number_of_apps):
            version = str(i)
            self.app_list.append(mex_controller.App(image_type='ImageTypeDocker',
                                                    app_name=app_name,
                                                    app_version=version,
                                                    ip_access=ip_access,
                                                    access_ports=access_ports,
                                                    cluster_name=cluster_name,
                                                    developer_name=developer_name,
                                                    default_flavor_name=flavor))

    def test_CreateApp100(self):
        # print the existing apps 
        app_pre = self.controller.show_apps()

        # create the app
        for i in self.app_list:
          resp = self.controller.create_app(i.app)

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # find app in list
        for a in self.app_list:
            found_app = a.exists(app_post)
            expect_equal(found_app, True, 'find app' + a.app_name)

        expect_equal(len(app_post), len(app_pre) + number_of_apps, 'number of apps')
        assert_expectations()

    @classmethod
    def tearDownClass(self):
        for a in self.app_list:
            self.controller.delete_app(a.app)
        self.controller.delete_cluster(self.cluster.cluster)
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

