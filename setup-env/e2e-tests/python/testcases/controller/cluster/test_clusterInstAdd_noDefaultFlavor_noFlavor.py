#!/usr/bin/python3

# EDGECLOUD-171 - fixed

#
# create a cluster without a default flavor.
# create a cluster instance without a flavor name.
# verify error 'No ClusterFlavor specified and no default ClusterFlavor for Cluster' is returned because there is no flavor to add
#
import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'
operator_name = 'tmus'
cloud_name = 'tmocloud-1'
flavor_name = 'c1.tiny'
cluster_name = 'cluster' + str(time.time())

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

        # no default flavor
        self.cluster = mex_controller.Cluster(cluster_name=cluster_name)
        self.cluster_instance_noFlavor = mex_controller.ClusterInstance(cluster_name=cluster_name,
                                                                        cloudlet_name=cloud_name,
                                                                        operator_name=operator_name
                                                                       )
        self.cluster_instance_emptyFlavor = mex_controller.ClusterInstance(cluster_name=cluster_name,
                                                                           cloudlet_name=cloud_name,
                                                                           flavor_name='',
                                                                           operator_name=operator_name
                                                                          )

        # create a new cluster for adding the instance
        create_cluster_resp = self.controller.create_cluster(self.cluster.cluster)

    def test_NoFlavor(self):
        # print the existing cluster instances
        self.controller.show_cluster_instances()

        # create the cluster instance withour the flavor_name
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance_noFlavor.cluster_instance)
        except:
            print('create cluster instance failed')

        # print the cluster instances after adding 
        time.sleep(1)
        clusterinst_resp = self.controller.show_cluster_instances()

        # verify ci.tiny is picked up from the default_flavor_name
        clusterinst_temp = self.cluster_instance_noFlavor
        clusterinst_temp.flavor_name = flavor_name
        found_cluster = clusterinst_temp.exists(clusterinst_resp)

        expect_equal(found_cluster, False, 'no flavor found new cluster')
        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        # EDGECLOUD-171
        expect_equal(self.controller.response.details(), 'No ClusterFlavor specified and no default ClusterFlavor for Cluster', 'error details')

        assert_expectations()

    def test_EmptyFlavor(self):
        # print the existing cluster instances
        self.controller.show_cluster_instances()

        # create the cluster instance withour the flavor_name
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance_emptyFlavor.cluster_instance)
        except:
            print('create controller instance failed')

        # print the cluster instances after adding
        time.sleep(1)
        clusterinst_resp = self.controller.show_cluster_instances()

        # verify ci.tiny is picked up from the default_flavor_name
        clusterinst_temp = self.cluster_instance_emptyFlavor
        clusterinst_temp.flavor_name = flavor_name
        found_cluster = clusterinst_temp.exists(clusterinst_resp)

        expect_equal(found_cluster, False, 'empty flavor found new cluster')
        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        # EDGECLOUD-171
        expect_equal(self.controller.response.details(), 'No ClusterFlavor specified and no default ClusterFlavor for Cluster', 'error details')

        assert_expectations()

    def tearDown(self):
        # delete cluster instance
        self.controller.delete_cluster(self.cluster.cluster)
        time.sleep(1)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

