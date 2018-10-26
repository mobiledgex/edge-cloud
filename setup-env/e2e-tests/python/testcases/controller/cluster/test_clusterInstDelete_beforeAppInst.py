#!/usr/bin/python3

#
# delte cluster instance before deleting the app instance that is using it
# verify 'ClusterInst in use by Application Instance' is received
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
cloud_name = 'tmocloud-1'
operator_name = 'dmuus'
flavor_name = 'c1.small'
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
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                      app_name=app_name,
                                      app_version=app_version,
                                      ip_access='IpAccessDedicatedOrShared',
                                      access_ports=access_ports,
                                      cluster_name=cluster_name,
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)

        self.controller.create_developer(self.developer.developer) 

        # create the cluster
        self.controller.create_cluster(self.cluster.cluster)

        # create the app
        resp = self.controller.create_app(self.app.app)

    def test_DeleteClusterBeforeApp(self):
        # print the existing cluster instances
        cluster_pre = self.controller.show_cluster_instances()
        

        # create cluster instance
        self.cluster_instance = mex_controller.ClusterInstance(cluster_name=cluster_name,
                                                               cloudlet_name=cloud_name,
                                                               operator_name=operator_name,
                                                               flavor_name=flavor_name)
        self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)

        # create the app instance
        self.app_instance = mex_controller.AppInstance(cloudlet_name=cloud_name,
                                                       app_name=app_name,
                                                       app_version=app_version,
                                                       developer_name=developer_name,
                                                       operator_name=operator_name)
        resp = self.controller.create_app_instance(self.app_instance.app_instance)

        # attempt to delete the cluster instance
        try:
            self.controller.delete_cluster_instance(self.cluster_instance.cluster_instance)
        except grpc.RpcError as e:
            print('error', type(e.code()), e.details())
            expect_equal(e.code(), grpc.StatusCode.UNKNOWN, 'status code')
            expect_equal(e.details(), 'ClusterInst in use by Application Instance', 'error details')
        else:
            print('cluster deleted')

        
        # print the cluster instances after error
        cluster_post = self.controller.show_cluster_instances()

        # find cluster in list
        found_cluster = self.cluster_instance.exists(cluster_post)

        expect_equal(found_cluster, True, 'find cluster instance')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_app_instance(self.app_instance.app_instance)
        #time.sleep(1) # wait till app instance is actually deleted else delete app will fail
        self.controller.delete_app(self.app.app)
        self.controller.delete_cluster_instance(self.cluster_instance.cluster_instance)
        self.controller.delete_cluster(self.cluster.cluster)
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

