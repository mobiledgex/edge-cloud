#!/usr/bin/python3

#
# create cloudinst with flavor_name only
# create cloudinst with operator_name only
# create cloudinst with cloudlet_name only
# create cloudinst with cluster_name only
# create cloudinst with no parms
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
    def setUp(self):
        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   )

        #self.cluster_instance = mex_controller.ClusterInstance(flavor_name='flavor_name')

    def test_CreateClusterInstFlavorOnly(self):
        # print the existing cluster instances
        clusterinst_pre = self.controller.show_cluster_instances()

        # create the cluster instance with flavor_name only
        self.cluster_instance = mex_controller.ClusterInstance(flavor_name='flavor_name')
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except Exception as e:
            print('got exception', e)

        # print the cluster instances after error
        clusterinst_post = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Invalid cluster name', 'error details')
        expect_equal(len(clusterinst_pre), len(clusterinst_post), 'same number of cluster')
        assert_expectations()

    def test_CreateClusterInstOperatorOnly(self):
        # print the existing cluster instances
        clusterinst_pre = self.controller.show_cluster_instances()

        # create the cluster instance with flavor_name only
        self.cluster_instance = mex_controller.ClusterInstance(operator_name='dmuus')
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except Exception as e:
            print('got exception', e)

        # print the cluster instances after error
        clusterinst_post = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Invalid cluster name', 'error details')
        expect_equal(len(clusterinst_pre), len(clusterinst_post), 'same number of cluster')
        assert_expectations()

    def test_CreateClusterInstCloudletNameOnly(self):
        # print the existing cluster instances
        clusterinst_pre = self.controller.show_cluster_instances()

        # create the cluster instance with flavor_name only
        self.cluster_instance = mex_controller.ClusterInstance(cloudlet_name='tmocloud-1')
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except Exception as e:
            print('got exception', e)

        # print the cluster instances after error
        clusterinst_post = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Invalid cluster name', 'error details')
        expect_equal(len(clusterinst_pre), len(clusterinst_post), 'same number of cluster')
        assert_expectations()

    def test_CreateClusterInstClusterNameOnly(self):
        # print the existing cluster instances
        clusterinst_pre = self.controller.show_cluster_instances()

        # create the cluster instance with flavor_name only
        self.cluster_instance = mex_controller.ClusterInstance(cluster_name='SmallCluster')
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except Exception as e:
            print('got exception', e)

        # print the cluster instances after error
        clusterinst_post = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Invalid operator name', 'error details')
        expect_equal(len(clusterinst_pre), len(clusterinst_post), 'same number of cluster')
        assert_expectations()

    def test_CreateClusterInstNoParms(self):
        # print the existing cluster instances
        clusterinst_pre = self.controller.show_cluster_instances()

        # create the cluster instance with flavor_name only
        self.cluster_instance = mex_controller.ClusterInstance()
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except Exception as e:
            print('got exception', e)

        # print the cluster instances after error
        clusterinst_post = self.controller.show_cluster_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Invalid cluster name', 'error details')
        expect_equal(len(clusterinst_pre), len(clusterinst_post), 'same number of cluster')
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

