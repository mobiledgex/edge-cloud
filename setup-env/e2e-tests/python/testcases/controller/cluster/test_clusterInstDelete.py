#!/usr/bin/python3

#
# create a cluster instance with and without flavor_name
# delete the cluster instance
# verify cluster instance is deleted
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
operator_name = 'dmuus'
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
        self.cluster_instance_flavor = mex_controller.ClusterInstance(cluster_name=self.cluster_name,
                                                                      cloudlet_name=cloud_name,
                                                                      operator_name=operator_name,
                                                                      flavor_name=flavor_name)
        self.cluster_instance_noflavor = mex_controller.ClusterInstance(cluster_name=self.cluster_name,
                                                                        cloudlet_name=cloud_name,
                                                                        operator_name=operator_name,
                                                                       )


    def test_DeleteClusterInstanceFlavor(self):
        # print the existing cluster instances
        clusterinst_before = self.controller.show_cluster_instances()

        # create a new cluster for adding the instance
        create_cluster_resp = self.controller.create_cluster(self.cluster.cluster)

        # create the cluster instance
        self.controller.create_cluster_instance(self.cluster_instance_flavor.cluster_instance)

        # print the cluster instances after adding 
        time.sleep(1)
        clusterinst_after_add = self.controller.show_cluster_instances()

        # look for the cluster
        found_cluster_after_add = self.cluster_instance_flavor.exists(clusterinst_after_add)

        expect_equal(found_cluster_after_add, True, 'found new cluster after add')
        expect_equal(len(clusterinst_after_add), len(clusterinst_before)+1, 'count after add')

        #delete the clusterinst
        self.controller.delete_cluster_instance(self.cluster_instance_flavor.cluster_instance)

        # print the cluster instances after adding
        time.sleep(1)
        clusterinst_after_delete = self.controller.show_cluster_instances()

        # look for the cluster after delete
        found_cluster_after_delete = self.cluster_instance_flavor.exists(clusterinst_after_delete)

        expect_equal(found_cluster_after_delete, False, 'found new cluster after delete')
        expect_equal(len(clusterinst_after_delete), len(clusterinst_before), 'count after delete')

        assert_expectations()

    def test_DeleteClusterInstanceNoFlavor(self):
        # print the existing cluster instances
        clusterinst_before = self.controller.show_cluster_instances()

        # create a new cluster for adding the instance
        create_cluster_resp = self.controller.create_cluster(self.cluster.cluster)

        # create the cluster instance
        self.controller.create_cluster_instance(self.cluster_instance_noflavor.cluster_instance)

        # print the cluster instances after adding
        time.sleep(1)
        clusterinst_after_add = self.controller.show_cluster_instances()

        # look for the cluster
        clusterinst_temp = self.cluster_instance_noflavor
        clusterinst_temp.flavor_name = flavor_name
        found_cluster_after_add = clusterinst_temp.exists(clusterinst_after_add)

        expect_equal(found_cluster_after_add, True, 'found new cluster after add')
        expect_equal(len(clusterinst_after_add), len(clusterinst_before)+1, 'count after add')

        #delete the clusterinst
        self.controller.delete_cluster_instance(self.cluster_instance_noflavor.cluster_instance)

        # print the cluster instances after adding
        time.sleep(1)
        clusterinst_after_delete = self.controller.show_cluster_instances()

        # look for the cluster after delete
        found_cluster_after_delete = clusterinst_temp.exists(clusterinst_after_delete)

        expect_equal(found_cluster_after_delete, False, 'found new cluster after delete')
        expect_equal(len(clusterinst_after_delete), len(clusterinst_before), 'count after delete')

        assert_expectations()

    def tearDown(self):
        self.controller.delete_cluster(self.cluster.cluster)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

