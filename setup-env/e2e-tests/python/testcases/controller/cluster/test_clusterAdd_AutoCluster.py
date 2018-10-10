#!/usr/bin/python3

#
# create a cluster with name=AutoCluster
# verify it fails with 'Cluster name prefix "AutoCluster" is reserved'
#

import unittest
import grpc
import sys
import time
import os
from delayedassert import expect, expect_equal, assert_expectations
import logging

import cluster_pb2
import cluster_pb2_grpc
import clusterflavor_pb2

controller_address = '127.0.0.1:55001'

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)

class tc(unittest.TestCase):
    def setUp(self):
        #controller_channel = grpc.insecure_channel(controller_address)
        self.mex_root_cert = self._findFile(mex_root_cert)
        self.mex_key = self._findFile(mex_key)
        self.mex_cert = self._findFile(mex_cert)

        with open(self.mex_root_cert, 'rb') as f:
            print('using root_cert =',mex_root_cert)
            #trusted_certs = f.read().encode()
            trusted_certs = f.read()
        with open(self.mex_key,'rb') as f:
            print('using key =',mex_key)
            trusted_key = f.read()
        with open(self.mex_cert, 'rb') as f:
            print('using client cert =', mex_cert)
            cert = f.read()

        credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs, private_key=trusted_key, certificate_chain=cert)
        controller_channel = grpc.secure_channel(controller_address, credentials)

        self.cluster_stub = cluster_pb2_grpc.ClusterApiStub(controller_channel)

        self.cluster = cluster_pb2.Cluster(
                                           key = cluster_pb2.ClusterKey(name = 'AutoClustername')
                                          )

    def test_AddAutoCluster(self):
        show_cluster_resp = self.cluster_stub.ShowCluster(cluster_pb2.Cluster())
        num_clusters_before = 0
        for c in show_cluster_resp:
            print('clusterBeforeAdd=', c)
            num_clusters_before += 1

        try:
            create_cluster_resp = self.cluster_stub.CreateCluster(self.cluster)
        except grpc.RpcError as e:
            print('error', type(e.code()), e.details())
            expect_equal(e.code(), grpc.StatusCode.UNKNOWN, 'status code')
            expect_equal(e.details(), 'Cluster name prefix "AutoCluster" is reserved', 'error details')
        else:
            print('cluster added',create_cluster_resp)

        show_cluster_resp = self.cluster_stub.ShowCluster(cluster_pb2.Cluster())
        num_clusters_after = 0
        for c in show_cluster_resp:
            print('clusterAfterAdd=', c)
            num_clusters_after += 1

        expect_equal(num_clusters_before, num_clusters_after, 'same number of cluster')
        assert_expectations()

    def _findFile(self, path):
        for dirname in sys.path:
            candidate = os.path.join(dirname, path)
            if os.path.isfile(candidate):
                return candidate
        raise Error('cant find file {}'.format(path))

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

