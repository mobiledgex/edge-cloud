#!/usr/bin/python3

#
# create a cluster instance with flavor_name
# delete the cluster instance with key not found error
# verify  proper error is generated
#

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'
operator_name = 'dmuus'
cloud_name = 'tmocloud-1'
flavor_name = 'c1.small'

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

class tc(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        self.cluster_name = 'cluster' + str(time.time())

        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   ) 

        self.cluster_instance_clusterNameOnly = mex_controller.ClusterInstance(cluster_name=self.cluster_name)
                                                                      
        self.cluster_instance_noflavor = mex_controller.ClusterInstance(cluster_name=self.cluster_name,
                                                                        cloudlet_name=cloud_name,
                                                                        operator_name=operator_name,
                                                                       )

        self.cluster_instance_noName = mex_controller.ClusterInstance(
                                                                        cloudlet_name=cloud_name,
                                                                        operator_name=operator_name,
                                                                       )

    def test_DeleteClusterNameOnly(self):
        # print the existing cluster instances
        clusterinst_before = self.controller.show_cluster_instances()

        # create the cluster instance
        try:
            self.controller.delete_cluster_instance(self.cluster_instance_clusterNameOnly.cluster_instance)
        except:
            print('delete cluster failed')

        # print the cluster instances after adding 
        #time.sleep(1)
        clusterinst_after_add = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Cloudlet operator_key:<>  not ready, state is CloudletStateNotPresent', 'error details')

        expect_equal(len(clusterinst_after_add), len(clusterinst_before), 'count after add')

        assert_expectations()

    def test_DeleteClusterNoFlavor(self):
        # print the existing cluster instances
        clusterinst_before = self.controller.show_cluster_instances()

        # create the cluster instance
        try:
            self.controller.delete_cluster_instance(self.cluster_instance_noflavor.cluster_instance)
        except:
            print('delete cluster failed')

        # print the cluster instances after adding
        #time.sleep(1)
        clusterinst_after_add = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Key not found', 'error details')

        expect_equal(len(clusterinst_after_add), len(clusterinst_before), 'count after add')

        assert_expectations()

    def test_DeleteClusterNoName(self):
        # print the existing cluster instances
        clusterinst_before = self.controller.show_cluster_instances()

        # create the cluster instance
        try:
            self.controller.delete_cluster_instance(self.cluster_instance_noName.cluster_instance)
        except:
            print('delete cluster failed')

        # print the cluster instances after adding
        #time.sleep(1)
        clusterinst_after_add = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Key not found', 'error details')

        expect_equal(len(clusterinst_after_add), len(clusterinst_before), 'count after add')

        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

