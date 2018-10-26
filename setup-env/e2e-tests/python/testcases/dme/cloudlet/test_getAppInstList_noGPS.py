#!/usr/bin/python3

# EDGECLOUD-144 - fixed
#
# send GetAppInstList without GPS location
# verify 'missing GPS location' is returned
#

import unittest
import grpc
import sys
import os
from delayedassert import expect, expect_equal, assert_expectations

import app_client_pb2
import app_client_pb2_grpc
import loc_pb2

dme_address = '127.0.0.1:50051'

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

class tc(unittest.TestCase):
    def setUp(self):
        #creds = grpc.ssl_channel_credentials(open('/root/go/src/github.com/mobiledgex/edge-cloud/tls/out/mex-ca.crt').read())
        #print('creds',type(creds))
        #channel = grpc.insecure_channel(dme_address)
        #channel = grpc.secure_channel(dme_address, creds)
        #print('channel',channel)

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
        channel = grpc.secure_channel(dme_address, credentials)

        self.stub = app_client_pb2_grpc.Match_Engine_ApiStub(channel)
        self.req = app_client_pb2.RegisterClientRequest(
                                                  DevName = 'AcmeAppCo'
                                                 )
        self.regResp = self.stub.RegisterClient(self.req)
        print('reg=',self.regResp)
        print('reqtpe=',type(self.regResp))
        print("cookie=",self.regResp.SessionCookie)
        print("turl=",self.regResp.TokenServerURI)

    def test_GetCloudletsNoGPS(self):
        get_cloudlets_req = app_client_pb2.AppInstListRequest(
                                                               )

        print(get_cloudlets_req)
        get_cloudlets_req.SessionCookie = self.regResp.SessionCookie

        try:
            get_cloudlets_resp = self.stub.GetAppInstList(get_cloudlets_req)
        except grpc.RpcError as e:
            print('GetCloudlets failed')
            expect_equal(e.code(), grpc.StatusCode.UNKNOWN, 'status code')
            expect_equal(e.details(), 'missing GPS location', 'error details')
        else:
            print('GetCloudlets passed')
            sys.exit(1)

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

