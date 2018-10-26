#!/usr/local/bin/python3

#
# create app with invalid digits for AccessLayerL4 and AccessLayerL4L7
# verify 'Failed to convert port A80 to integer: strconv.ParseInt: parsing "A80": invalid syntax' is received
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations
import logging

import mex_controller

controller_address = '127.0.0.1:55001'
stamp = str(int(time.time()))
controller_address = '127.0.0.1:55001'
app_name = 'appname' + stamp
app_version = '1.0'
developer_name = 'developer' + stamp

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
    def test_CreateAppPortInvalidUnknown(self):
        # print the existing apps 
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 developer_name=developer_name,
                                 app_name=app_name,
                                 app_version=app_version,
                                 ip_access='IpAccessUnknown',
                                 access_ports='tcp:A80')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Failed to convert port A80 to integer: strconv.ParseInt: parsing "A80": invalid syntax', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortInvalidDedicated(self):
        # print the existing apps 
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 developer_name=developer_name,
                                 app_name=app_name,
                                 app_version=app_version,
                                 ip_access='IpAccessDedicated',
                                 access_ports='tcp:A80')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Failed to convert port A80 to integer: strconv.ParseInt: parsing "A80": invalid syntax', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortInvalidDedicatedShared(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 developer_name=developer_name,
                                 app_name=app_name,
                                 app_version=app_version,
                                 ip_access='IpAccessDedicatedOrShared',
                                 access_ports='udp:xx')

        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Failed to convert port xx to integer: strconv.ParseInt: parsing "xx": invalid syntax', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortInvalidShared(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 developer_name=developer_name,
                                 app_name=app_name,
                                 app_version=app_version,
                                 ip_access='IpAccessShared',
                                 access_ports='udp:xx')

        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Failed to convert port xx to integer: strconv.ParseInt: parsing "xx": invalid syntax', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

