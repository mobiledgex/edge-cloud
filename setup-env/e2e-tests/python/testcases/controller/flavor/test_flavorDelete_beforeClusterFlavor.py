#!/usr/local/bin/python3

#
# attempt to delete flavor before the cluster flavor
# verify 'Flavor in use by Cluster Flavor' error is received
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
flavor_name = 'flavor' + stamp
ram = 1024
vcpus = 1
disk = 1
cluster_flavor_name = 'clusterflavor' + stamp

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

        self.flavor = mex_controller.Flavor(flavor_name=flavor_name,
                                            ram=ram,
                                            disk=disk,
                                            vcpus=vcpus)
        self.cluster_flavor = mex_controller.ClusterFlavor(cluster_flavor_name=cluster_flavor_name,
                                                    node_flavor_name=flavor_name,
                                                    master_flavor_name=flavor_name)

        self.controller.create_flavor(self.flavor.flavor) 
        self.controller.create_cluster_flavor(self.cluster_flavor.cluster_flavor)

    def test_DeleteFlavorBeforeClusterFlavor(self):
        # print flavors before delete
        flavor_pre = self.controller.show_flavors()

        # delete flavor
        error = None
        try:
            self.controller.delete_flavor(self.flavor.flavor)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print flavors after delete
        flavor_post = self.controller.show_flavors()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Flavor in use by Cluster Flavor', 'error details')
        expect_equal(len(flavor_post), len(flavor_pre), 'num flavor')

        assert_expectations()

    @classmethod
    def tearDownClass(self):
        self.controller.delete_cluster_flavor(self.cluster_flavor.cluster_flavor)
        self.controller.delete_flavor(self.flavor.flavor)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

