#!/usr/bin/python3

#
# create a cloudlet and cluster
# create a cluster instance for a cloudlet that does not exist in CloudletInfo
# verify 'No resource information found for Cloudlet' is returned
#
 
import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

sys.path.append('/root/andy/python/protos')

import mex_controller

controller_address = '127.0.0.1:55001'

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

class tc(unittest.TestCase):
    def setUp(self):
        stamp = str(time.time())
        cluster_name = 'cluster' + stamp
        operator_name = 'dmuus'
        cloud_name = 'cloudlet' + stamp 
        flavor_name = 'c1.small'

        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   )

        self.cloudlet = mex_controller.Cloudlet(cloudlet_name = cloud_name,
                                                operator_name = operator_name,
                                                number_of_dynamic_ips = 254)
        self.cluster = mex_controller.Cluster(cluster_name=cluster_name,
                                              default_flavor_name=flavor_name)
        self.cluster_instance = mex_controller.ClusterInstance(cluster_name=cluster_name,
                                                             cloudlet_name=cloud_name,
                                                             operator_name=operator_name,
                                                             flavor_name=flavor_name)
        
    def test_CreateClusterInstCloudletNotFound(self):
        # create a cloudlet
        self.controller.create_cloudlet(self.cloudlet.cloudlet)

        # create a new cluster for adding the instance
        create_cluster_resp = self.controller.create_cluster(self.cluster.cluster)

        # print the existing cluster instances
        clusterinst_pre = self.controller.show_cluster_instances()

        # create the cluster instance
        resp = None
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except:
            print('create cluster instance failed')

        # print the cluster instances after error
        clusterinst_post = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'No resource information found for Cloudlet', 'error details')
        expect_equal(len(clusterinst_pre), len(clusterinst_post), 'same number of cluster')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_cluster(self.cluster.cluster)
        self.controller.delete_cloudlet(self.cloudlet.cloudlet)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

