#!/usr/bin/python3

#
# create an app instance for various ways the app is not sent or not found
# verify 'Specified app not found' is returned
#
 
import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

sys.path.append('/root/andy/python/protos')

import mex_controller

controller_address = '127.0.0.1:55001'

stamp = str(time.time())
cloud_name = 'cloud' + stamp
operator_name = 'operator' + stamp

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

        self.operator = mex_controller.Operator(operator_name = operator_name)
        self.cloudlet = mex_controller.Cloudlet(cloudlet_name = cloud_name,
                                                operator_name = operator_name,
                                                number_of_dynamic_ips = 254)

        self.controller.create_operator(self.operator.operator)
        self.controller.create_cloudlet(self.cloudlet.cloudlet)

    def test_CreateAppInstAppNotFound_nodata(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()
           
        # create the app instance
        app_instance = mex_controller.AppInstance(cloudlet_name=cloud_name,
                                                  operator_name=operator_name)

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified app not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstAppNotFound_idonly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(appinst_id=1,
                                                  cloudlet_name=cloud_name,
                                                  operator_name=operator_name)

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified app not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstAppNotFound_appnameonly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(app_name='smeApplication',
                                                  cloudlet_name=cloud_name,
                                                  operator_name=operator_name)

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified app not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstAppNotFound_versiononly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(app_version='1.0',
                                                  cloudlet_name=cloud_name,
                                                  operator_name=operator_name)

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified app not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstAppNotFound_developeronly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(developer_name='dev',
                                                  cloudlet_name=cloud_name,
                                                  operator_name=operator_name)

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified app not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def test_CreateAppInstAppNotFound_nameIdDeveloperonly(self):
        # print the existing app instances
        appinst_pre = self.controller.show_app_instances()

        # create the app instance
        app_instance = mex_controller.AppInstance(app_name='smeApplication',
                                                  app_version='1.0',
                                                  developer_name='dev',
                                                  cloudlet_name=cloud_name,
                                                  operator_name=operator_name)

        resp = None
        try:
            resp = self.controller.create_app_instance(app_instance.app_instance)
        except:
            print('create app instance failed')

        # print the cluster instances after error
        appinst_post = self.controller.show_app_instances()

        expect_equal(self.controller.response.code(), grpc.StatusCode.UNKNOWN, 'status code')
        expect_equal(self.controller.response.details(), 'Specified app not found', 'error details')
        expect_equal(len(appinst_pre), len(appinst_post), 'same number of app ainst')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_cloudlet(self.cloudlet.cloudlet)
        time.sleep(1)
        self.controller.delete_operator(self.operator.operator)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

