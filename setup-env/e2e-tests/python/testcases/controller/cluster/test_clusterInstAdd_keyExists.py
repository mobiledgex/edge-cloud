#!/usr/bin/python3

#
# create the same cluster twice
# verify error of 'Key already exists' is retruned
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

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)



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
        self.cluster = mex_controller.Cluster(cluster_name=cluster_name,
                                              default_flavor_name=flavor_name)

        self.cluster_instance = mex_controller.ClusterInstance(cluster_name=cluster_name,
                                                             cloudlet_name=cloud_name,
                                                             operator_name=operator_name,
                                                             flavor_name=flavor_name)

    def test_CreateClusterInstFlavorOnly(self):
        # print the existing cluster instances
        clusterinst_pre = self.controller.show_cluster_instances()

        # create a new cluster and cluster instance
        create_cluster_resp = self.controller.create_cluster(self.cluster.cluster)
        self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        #time.sleep(1)

        # create the cluster instance which already exists
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except:
            print('create cluster instance failed')
        # print the cluster instances after error
        clusterinst_post = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Key already exists', 'error details')
        expect_equal(len(clusterinst_pre)+1, len(clusterinst_post), 'same number of cluster')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_cluster_instance(self.cluster_instance.cluster_instance)
        self.controller.delete_cluster(self.cluster.cluster)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

