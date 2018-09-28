#!/usr/bin/python3

#
# create 100 clusters and 100 cluster instances
# verify all cluster instance is created
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

number_of_clusterInsts = 100 

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

        self.cluster_list = []
        self.clusterinst_list = []

        self.stamp = str(time.time())
        for i in range(number_of_clusterInsts):
            cluster_name = 'cluster' + str(i) + '-' + self.stamp
            self.cluster_list.append(mex_controller.Cluster(cluster_name=cluster_name,
                                                            default_flavor_name=flavor_name))
            self.clusterinst_list.append(mex_controller.ClusterInstance(cluster_name=cluster_name,
                                                                        cloudlet_name=cloud_name,
                                                                        operator_name=operator_name,
                                                                        flavor_name=flavor_name))

    def test_AddClusterInstance(self):
        # print the existing cluster and cluster instances
        clusters_before = self.controller.show_clusters()
        cluster_instances_before = self.controller.show_cluster_instances()

        # create a new cluster for adding the instance
        for i in self.cluster_list:
            self.controller.create_cluster(i.cluster)

        # create the cluster instance
        for i in self.clusterinst_list:
            self.controller.create_cluster_instance(i.cluster_instance)

        # print the cluster instances after adding 
        time.sleep(1)
        clusters_after = self.controller.show_clusters()
        cluster_instances_after = self.controller.show_cluster_instances()

        # look for the cluster
        for c in self.clusterinst_list:
            found_cluster = c.exists(cluster_instances_after)
            expect_equal(found_cluster, True, 'find new cluster' + c.cluster_name)
        
        assert_expectations()

    def tearDown(self):
        for c in self.clusterinst_list:
            self.controller.delete_cluster_instance(c.cluster_instance)
        for c in self.cluster_list:
            self.controller.delete_cluster(c.cluster)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

