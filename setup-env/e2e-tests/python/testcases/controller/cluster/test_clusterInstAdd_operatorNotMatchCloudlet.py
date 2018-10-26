#!/usr/bin/python3

#
# create cluster
# create cloudinst with cloudletname that exists in cloudlets but does not match operator_name 
# verify Specified Cloudlet not found 
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'
operator_name = 'ATT'
cloud_name = 'tmocloud-1'
flavor_name = 'c1.small'
cluster_name = 'cluster' + str(time.time())

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

        self.cluster = mex_controller.Cluster(cluster_name=cluster_name,
                                         default_flavor_name=flavor_name)

        self.cluster_instance = mex_controller.ClusterInstance(cluster_name=cluster_name,
                                                             cloudlet_name=cloud_name,
                                                             operator_name=operator_name,
                                                             flavor_name=flavor_name)

    def test_OperatorNotMatchCloudlet(self):
        # print the existing cluster instances
        clusterinst_pre = self.controller.show_cluster_instances()

        # create a new cluster for adding the instance
        create_cluster_resp = self.controller.create_cluster(self.cluster.cluster)

        # create the cluster instance with operator name that does not match cloudlet operator name 
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except Exception as e:
            print('got exception', e)

        # print the cluster instances after error
        clusterinst_post = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Cloudlet operator_key:<name:"' + operator_name + '" > name:"' + cloud_name + '"  not ready, state is CloudletStateNotPresent', 'error details')
        expect_equal(len(clusterinst_pre), len(clusterinst_post), 'same number of cluster')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_cluster(self.cluster.cluster)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

