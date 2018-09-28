#!/usr/bin/python3

#
# create app with unsupported port protocol AccessLayerL4 and AccessLayerL4L7
# verify '%s is not a supported L4 Protocol' is received
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

sys.path.append('/root/andy/python/protos')

import mex_controller

controller_address = '127.0.0.1:55001'

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

    def test_CreateAppPortRangeL4_1(self):
        # print the existing apps 
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker', 
                                 access_layer='AccessLayerL4',
                                 access_ports='tcp:0')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 0 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeL4_2(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 access_layer='AccessLayerL4',
                                 access_ports='tcp:-1')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port -1 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeL4_3(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 access_layer='AccessLayerL4',
                                 access_ports='tcp:65536')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 65536 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeL4_4(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 access_layer='AccessLayerL4',
                                 access_ports='tcp:1,tcp:65537,tcp:65535')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 65537 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeL4L7_1(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker', 
                                 access_layer='AccessLayerL4L7',
                                 access_ports='udp:0')

        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 0 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeL4L7_2(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 access_layer='AccessLayerL4L7',
                                 access_ports='udp:-1')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port -1 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeL4L7_3(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 access_layer='AccessLayerL4L7',
                                 access_ports='udp:65536')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
            error = e

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        expect_equal(error.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(error.details(), 'Port 65536 out of range', 'error details')
        expect_equal(len(app_pre), len(app_post), 'same number of apps')
        assert_expectations()

    def test_CreateAppPortRangeL4L7_4(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app with no parms
        error = None
        app = mex_controller.App(image_type='ImageTypeDocker',
                                 access_layer='AccessLayerL4L7',
                                 access_ports='udp:1,udp:65537,udp:65535')
        try:
            resp = self.controller.create_app(app.app)
        except grpc.RpcError as e:
            print('got exception', e)
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

