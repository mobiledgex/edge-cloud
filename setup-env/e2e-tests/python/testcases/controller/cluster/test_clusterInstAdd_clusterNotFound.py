#!/usr/bin/python3

#
# create a cluster instance for a cluster that does not exist
# verify 'Specified Cluster not found' is returned
#
 
import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'


class tc(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        cluster_name = 'cluster' + str(time.time())
        operator_name = 'dmuus'
        cloud_name = 'tmocloud-1'
        flavor_name = 'c1.small'

        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   )

        self.cluster_instance = mex_controller.ClusterInstance(cluster_name=cluster_name,
                                                             cloudlet_name=cloud_name,
                                                             operator_name=operator_name,
                                                             flavor_name=flavor_name)

    def test_CreateClusterInstFlavorOnly(self):
        # print the existing cluster instances
        clusterinst_pre = self.controller.show_cluster_instances()

        # create the cluster instance
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except:
            print('error creating cluster instance')

        # print the cluster instances after error
        clusterinst_post = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified Cluster not found', 'error details')
        expect_equal(len(clusterinst_pre), len(clusterinst_post), 'same number of cluster')
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

