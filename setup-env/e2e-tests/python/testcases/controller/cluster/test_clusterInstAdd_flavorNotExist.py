#!/usr/bin/python3

#
# create a cluster.
# create a cluster instance with a flavor name that doesnt exist in ShowFlavor.
# verify error "Cluster flavor <flavornamr> not found" is returned because flavor does not exist
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
cluster_name = 'cluster' + str(time.time())
flavor_name = 'fakeflavor'

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'


class tc(unittest.TestCase):
    def setUp(self):
        self.controller = mex_controller.Controller(controller_address = controller_address,
                                                    root_cert = mex_root_cert,
                                                    key = mex_key,
                                                    client_cert = mex_cert
                                                   ) 

        # no default flavor
        self.cluster = mex_controller.Cluster(cluster_name=cluster_name)
        self.cluster_instance = mex_controller.ClusterInstance(cluster_name=cluster_name,
                                                               cloudlet_name=cloud_name,
                                                               flavor_name=flavor_name,
                                                               operator_name=operator_name
                                                              )

        # create a new cluster for adding the instance
        create_cluster_resp = self.controller.create_cluster(self.cluster.cluster)

    def test_NoFlavor(self):
        # print the existing cluster instances
        self.controller.show_cluster_instances()

        # create the cluster instance withour the flavor_name
        try:
            resp = self.controller.create_cluster_instance(self.cluster_instance.cluster_instance)
        except:
            print('create cluster instance failed')

        # print the cluster instances after adding 
        time.sleep(1)
        clusterinst_resp = self.controller.show_cluster_instances()

        # verify clusterinst does not exist
        clusterinst_temp = self.cluster_instance
        found_cluster = clusterinst_temp.exists(clusterinst_resp)

        expect_equal(found_cluster, False, 'no flavor found new cluster')
        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Cluster flavor {} not found'.format(flavor_name), 'error details')

        assert_expectations()

    def tearDown(self):
        # delete cluster instance
        self.controller.delete_cluster(self.cluster.cluster)
        time.sleep(1)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

