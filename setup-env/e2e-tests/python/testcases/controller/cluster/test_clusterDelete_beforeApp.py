#!/usr/bin/python3

#
# delte cluster before deleting the app that is using it
# verify 'Cluster in use by Application' is received
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


    def test_DeleteClusterBeforeApp(self):
        # print the existing apps 
        cluster_pre = self.controller.show_clusters()

        # create the cluster
        self.controller.create_cluster(self.cluster.cluster)
        
        # create the app
        # contains image_type=Docker and no image_path
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                      app_name=app_name,
                                      app_version=app_version,
                                      ip_access='IpAccessShared',
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # attempt to delete the cluster
        try:
            self.controller.delete_cluster(self.cluster.cluster)
        except grpc.RpcError as e:
            print('error', type(e.code()), e.details())
            expect_equal(e.code(), grpc.StatusCode.UNKNOWN, 'status code')
            expect_equal(e.details(), 'Cluster in use by Application', 'error details')
        else:
            print('cluster deleted')

        
        # print the cluster instances after error
        cluster_post = self.controller.show_clusters()

        # find cluster in list
        found_cluster = self.cluster.exists(cluster_post)

        expect_equal(found_cluster, True, 'find cluster')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_app(self.app.app)
        self.controller.delete_cluster(self.cluster.cluster)
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

