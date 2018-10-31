#!/usr/bin/python3

# EDGECLOUD-141 - fixed
#
# send VerifyLocation with no GPS location
# verify 'Missing GpsLocation' is received
#

import grpc
import unittest
from delayedassert import expect, expect_equal, assert_expectations
import sys
import os

import app_client_pb2
import app_client_pb2_grpc
import loc_pb2

dme_address = '127.0.0.1:50051'

mex_root_cert = 'mex-ca.crt'
mex_cert = 'localserver.crt'
mex_key = 'localserver.key'

class tc(unittest.TestCase):
    def setUp(self):
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
        req = app_client_pb2.RegisterClientRequest(
                                                     DevName = 'AcmeAppCo',
                                                     AppName = 'someapplication',
                                                     AppVers = '1.0'
        ) 
        regResp = self.stub.RegisterClient(req)
        print('cookie',regResp.SessionCookie)
        print('tokenuri',regResp.TokenServerURI)
        print(regResp)
        self.locreq = app_client_pb2.VerifyLocationRequest(CarrierName = 'TMUS',
                                                          #DevName = 'AcmeAppCo',
                                                          #AppName = 'someApplication',
                                                          #AppVers = '1.0',
#                                                          GpsLocation = loc_pb2.Loc(lat=49.8614, long=8.5676),
                                                          #VerifyLocToken = "0000000001",
                                                          SessionCookie = regResp.SessionCookie
                                                         )

    def test_verifyLocationNoGPS(self):

        try:
            locResp = self.stub.VerifyLocation(self.locreq)
        except grpc.RpcError as e:
            print('VerifyLocatoin failed')
            expect_equal(e.code(), grpc.StatusCode.UNKNOWN, 'status code')
            expect_equal(e.details(), 'Missing GpsLocation', 'error details')
        else:
            print('VerifyLocation passed')
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

