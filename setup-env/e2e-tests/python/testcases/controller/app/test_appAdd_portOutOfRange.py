#!/usr/local/bin/python3

#
# create app with port out of range
# verify 'Port out of range' is received
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

    def test_CreateAppPortRangeDedicated_1(self):
        # print the existing apps 
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker', 
                                 ip_access='IpAccessDedicated',
                                 access_ports='tcp:0')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 0 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeDedicated_2(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 ip_access='IpAccessDedicated',
                                 access_ports='tcp:-1')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port -1 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeDedicated_3(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 ip_access='IpAccessDedicated',
                                 access_ports='tcp:65536')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 65536 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeDedicated_4(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 ip_access='IpAccessDedicated',
                                 access_ports='tcp:1,tcp:65537,tcp:65535')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 65537 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeDedicatedShared_1(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker', 
                                 ip_access='IpAccessDedicatedOrShared',
                                 access_ports='udp:0')

        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception ' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 0 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeDedicateShared_2(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 ip_access='IpAccessDedicatedOrShared',
                                 access_ports='udp:-1')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port -1 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeDedicatedShared_3(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 ip_access='IpAccessDedicatedOrShared',
                                 access_ports='udp:65536')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 65536 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeDedicatedShared_4(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 ip_access='IpAccessDedicatedOrShared',
                                 access_ports='udp:1,udp:65537,udp:65535')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            logger.info('got exception' + str(e))
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 65537 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()


if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

