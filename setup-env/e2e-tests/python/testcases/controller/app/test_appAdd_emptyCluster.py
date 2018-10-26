#!/usr/local/bin/python3

# EDGECLOUD-208 - now has lower case 'autocluster' - fixed
# EDGECLOUD-240 - creating an autocluster doesnot always pick the same default_flavor
# create app with empty cluster and no cluster parm  
# verify AutoCluster is created in Cluster and has smallest flavor
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
flavor = 'x1.tiny'
cluster_name = 'AutoCluster'
app_name = 'app' + stamp
app_version = '1.0'
cluster_flavor_name = 'c1.medium_2' + stamp
node_flavor_name = 'x1.tiny'
master_flavor_name = 'x1.small'
number_nodes = 1
max_nodes = 1
number_masters = 9
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
        self.cluster_flavor = mex_controller.ClusterFlavor(cluster_flavor_name=cluster_flavor_name,
                                                           node_flavor_name=node_flavor_name,
                                                           master_flavor_name=master_flavor_name,
                                                           number_nodes=number_nodes,
                                                           max_nodes=max_nodes,
                                                           number_masters=number_masters)

        self.controller.create_developer(self.developer.developer) 
        self.controller.create_cluster_flavor(self.cluster_flavor.cluster_flavor)
        self.controller.show_cluster_flavors()
        #time.sleep(1)
    def test_CreateAppNoCluster(self):
        # print the existing apps 
        app_pre = self.controller.show_apps()

        # print the clusters
        cluster_pre = self.controller.show_clusters()

        # create the app
        # contains no cluster and image_type=Docker
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      #cluster_name='',
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)

        resp = self.controller.create_app(self.app.app)

        # print the apps instances after error
        app_post = self.controller.show_apps()

        # print the cluster after add app
        cluster_post = self.controller.show_clusters()

        # find app in list
        apptemp = self.app
        # controller creates cluster with AutoCluster + app_name since cluster is empty
        apptemp.cluster_name = 'autocluster' + app_name  
        found_app = apptemp.exists(app_post)

        # find autocluster in list
        #time.sleep(1)
        cluster = mex_controller.Cluster(cluster_name='autocluster' + app_name,
                                         default_flavor_name=cluster_flavor_name)
        found_cluster = cluster.exists(cluster_post)

        self.controller.delete_app(self.app.app)
        
        expect_equal(found_cluster, True, 'find cluster')
        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_CreateAppEmptyCluster(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # print the clusters
        cluster_pre = self.controller.show_clusters()

        # create the app
        # contains no cluster and image_type=Docker
        self.app = mex_controller.App(image_type='ImageTypeQCOW',
                                      app_name=app_name,
                                      app_version=app_version,
                                      access_ports=access_ports,
                                      cluster_name='',
                                      developer_name=developer_name,
                                      default_flavor_name=flavor)

        resp = self.controller.create_app(self.app.app)

        # print the apps instances after error
        app_post = self.controller.show_apps()

        # print the cluster after add app
        cluster_post = self.controller.show_clusters()

        # find app in list
        apptemp = self.app
        # controller creates cluster with AutoCluster + app_name since cluster is empty
        apptemp.cluster_name = 'autocluster' + app_name
        found_app = apptemp.exists(app_post)

        # find autocluster in list
        #time.sleep(1)
        cluster = mex_controller.Cluster(cluster_name='autocluster' + app_name,
                                         default_flavor_name=cluster_flavor_name)
        found_cluster = cluster.exists(cluster_post)

        self.controller.delete_app(self.app.app)
        
        expect_equal(found_cluster, True, 'find cluster')
        expect_equal(found_app, True, 'find app')
        assert_expectations()

    @classmethod
    def tearDownClass(self):
        self.controller.delete_developer(self.developer.developer)
        self.controller.delete_cluster_flavor(self.cluster_flavor.cluster_flavor)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

