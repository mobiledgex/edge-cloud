#!/usr/bin/python3

#
# create a cluster instance with  liveness=1(LivenessStatic)
# verify cluster instance is created
#

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

sys.path.append('/root/andy/python/protos')
print(sys.path)

import mex_controller

controller_address = '127.0.0.1:55001'
operator_name = 'tmus'
cloud_name = 'tmocloud-1'
flavor_name = 'c1.small'

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

class tc(unittest.TestCase):
    def setUp(self):
        self.cluster_name = 'cluster' + str(time.time())

        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   ) 

        self.cluster = mex_controller.Cluster(cluster_name=self.cluster_name,
                                         default_flavor_name=flavor_name)
        self.cluster_instance = mex_controller.ClusterInstance(cluster_name=self.cluster_name,
                                                             cloudlet_name=cloud_name,
                                                             operator_name=operator_name,
                                                             flavor_name=flavor_name,
                                                             liveness=1)

    def test_AddClusterInstance(self):
        # print the existing cluster instances
        self.controller.show_cluster_instances()

        # create a new cluster for adding the instance
        create_cluster_resp = self.controller.create_cluster(self.cluster.cluster)

        # create the cluster instance
        self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)

        # print the cluster instances after adding 
        time.sleep(1)
        clusterinst_resp = self.controller.show_cluster_instances()

        # look for the cluster
        found_cluster = self.cluster_instance.exists(clusterinst_resp)

        expect_equal(found_cluster, True, 'found new cluster')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_cluster_instance(self.cluster_instance.cluster_instance)
        self.controller.delete_cluster(self.cluster.cluster)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

