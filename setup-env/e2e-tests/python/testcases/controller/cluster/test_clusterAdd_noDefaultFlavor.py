#!/usr/bin/python3

#
# create cluster with no default flavor
# verify cluster is created
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

        self.cluster_name = 'cluster' + str(time.time())
        self.cluster = cluster_pb2.Cluster(
                                           key = cluster_pb2.ClusterKey(name = self.cluster_name)
                                          )

    def test_GetCloudlets(self):
        show_cluster_resp = self.cluster_stub.ShowCluster(cluster_pb2.Cluster())
        for c in show_cluster_resp:
            print('clusterBeforeAdd=', c)

        create_cluster_resp = self.cluster_stub.CreateCluster(self.cluster)
        print('x',create_cluster_resp)

        show_cluster_resp = self.cluster_stub.ShowCluster(cluster_pb2.Cluster())
        found_cluster = False
        for c in show_cluster_resp:
            print('clusterAfterAdd=', c)
            if c.key.name == self.cluster_name and c.default_flavor.name == '':
                found_cluster = True
                print('foundkey')  

        expect_equal(found_cluster, True, 'found new cluster')
        assert_expectations()

    def _findFile(self, path):
        for dirname in sys.path:
            candidate = os.path.join(dirname, path)
            if os.path.isfile(candidate):
                return candidate
        raise Error('cant find file {}'.format(path))

    def tearDown(self):
        delete_cluster_resp = self.cluster_stub.DeleteCluster(self.cluster)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

