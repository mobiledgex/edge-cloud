#!/usr/bin/python3

#
# create an app instance for various ways the cloudlet is not sent or not found
# verify 'Specified cloudlet not found' is returned
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

    def test_CreateAppInstCloudletNotFound_nodata(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()
           
        # create the app instance
        app_instance = mex_controller.AppInstance()

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified cloudlet not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstCloudletNotFound_idonly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(appinst_id=1)

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified cloudlet not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstCloudletNotFound_appnameonly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(app_name='someApplication')

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified cloudlet not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstCloudletNotFound_versiononly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(app_version='1.0')

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified cloudlet not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstCloudletNotFound_developeronly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(developer_name='dev')

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified cloudlet not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstCloudletNotFound_nameIdDeveloperonly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(app_name='someApplication',
                                                  app_version='1.0',
                                                  developer_name='dev')

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified cloudlet not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstCloudletNotFound_cloudletNotFound(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(app_name='someApplication',
                                                  app_version='1.0',
                                                  developer_name='dev',
                                                  cloudlet_name='nocloud',
                                                  operator_name='TMUS')

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified cloudlet not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstCloudletNotFound_cloudletOperatorOnly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(app_name='someApplication',
                                                  app_version='1.0',
                                                  developer_name='dev',
                                                  operator_name='TMUS')

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified cloudlet not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstCloudletNotFound_cloudletNameOnly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(app_name='someApplication',
                                                  app_version='1.0',
                                                  developer_name='dev',
                                                  cloudlet_name='tmocloud-1')

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified cloudlet not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

