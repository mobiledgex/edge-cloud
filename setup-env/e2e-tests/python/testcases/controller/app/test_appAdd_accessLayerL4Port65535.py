#!/usr/bin/python3

#
# create app with access_layer=AccessLayerL4 with port 65535 and with Docker and QCOW
# verify app is created
# 

import unittest
import grpc
import sys
import time
from delayedassert import expect, expect_equal, assert_expectations

sys.path.append('/root/andy/python/protos')

import mex_controller

stamp = str(time.time())
controller_address = '127.0.0.1:55001'
developer_name = 'developer' + stamp
developer_address = 'allen tx'
developer_email = 'dev@dev.com'
flavor = 'x1.small'
cluster_name = 'cluster' + stamp
app_name = 'app' + stamp
app_version = '1.0'

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

        self.developer = mex_controller.Developer(developer_name=developer_name,
                                                  address=developer_address,
                                                  email=developer_email)
        self.cluster = mex_controller.Cluster(cluster_name=cluster_name,
                                              default_flavor_name=flavor)


        self.controller.create_developer(self.developer.developer) 
        self.controller.create_cluster(self.cluster.cluster)

    def test_CreateAppDockerAccessLayerL4DockerTCP65535(self):
        # print the existing apps 
        app_pre = self.controller.show_apps()

        # create the app
        # contains access_layer=AccessLayerL4 and port=tcp:65535
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                             app_name=app_name,
                                             app_version=app_version,
                                             cluster_name=cluster_name,
                                             developer_name=developer_name,
                                             access_layer = 'AccessLayerL4',
                                             access_ports = 'tcp:65535',
                                             default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # look for app
        found_app = self.app.exists(app_post)

        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_CreateAppDockerAccessLayerL4DockerTCP065535(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app
        # contains access_layer=AccessLayerL4 and port=tcp:01
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                             app_name=app_name,
                                             app_version=app_version,
                                             cluster_name=cluster_name,
                                             developer_name=developer_name,
                                             access_layer = 'AccessLayerL4',
                                             access_ports = 'tcp:065535',
                                             default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # look for app
        found_app = self.app.exists(app_post)

        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_CreateAppDockerAccessLayerL4DockerUDP65535(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app
        # contains access_layer=AccessLayerL4 and port=udp65535:
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                             app_name=app_name,
                                             app_version=app_version,
                                             cluster_name=cluster_name,
                                             developer_name=developer_name,
                                             access_layer = 'AccessLayerL4',
                                             access_ports = 'udp:65535',
                                             default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # look for app
        found_app = self.app.exists(app_post)

        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_CreateAppDockerAccessLayerL4DockerUDP065535(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app
        # contains access_layer=AccessLayerL4 and port=udp:065535
        self.app = mex_controller.App(image_type='ImageTypeDocker',
                                             app_name=app_name,
                                             app_version=app_version,
                                             cluster_name=cluster_name,
                                             developer_name=developer_name,
                                             access_layer = 'AccessLayerL4',
                                             access_ports = 'udp:065535',
                                             default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # look for app
        found_app = self.app.exists(app_post)

        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_CreateAppQCOWAccessLayerL4TCP65535(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app
        # contains access_layer=AccessLayerL4 QCOW tcp:65535
        self.app = mex_controller.App(image_type='ImageTypeQCOW',
                                             app_name=app_name,
                                             app_version=app_version,
                                             cluster_name=cluster_name,
                                             developer_name=developer_name,
                                             access_layer = 'AccessLayerL4',
                                             access_ports = 'tcp:65535',
                                             default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # look for app
        found_app = self.app.exists(app_post)

        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_CreateAppQCOWAccessLayerL4TCP065535(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app
        # contains access_layer=AccessLayerL4 QCOW tcp65535:
        self.app = mex_controller.App(image_type='ImageTypeQCOW',
                                             app_name=app_name,
                                             app_version=app_version,
                                             cluster_name=cluster_name,
                                             developer_name=developer_name,
                                             access_layer = 'AccessLayerL4',
                                             access_ports = 'tcp:065535',
                                             default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # look for app
        found_app = self.app.exists(app_post)

        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_CreateAppQCOWAccessLayerL4UDP65535(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app
        # contains access_layer=AccessLayerL4 QCOW udp:65535
        self.app = mex_controller.App(image_type='ImageTypeQCOW',
                                             app_name=app_name,
                                             app_version=app_version,
                                             cluster_name=cluster_name,
                                             developer_name=developer_name,
                                             access_layer = 'AccessLayerL4',
                                             access_ports = 'udp:65535',
                                             default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # look for app
        found_app = self.app.exists(app_post)

        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def test_CreateAppQCOWAccessLayerL4UDP065535(self):
        # print the existing apps
        app_pre = self.controller.show_apps()

        # create the app
        # contains access_layer=AccessLayerL4 QCOW udp:065535
        self.app = mex_controller.App(image_type='ImageTypeQCOW',
                                             app_name=app_name,
                                             app_version=app_version,
                                             cluster_name=cluster_name,
                                             developer_name=developer_name,
                                             access_layer = 'AccessLayerL4',
                                             access_ports = 'udp:065535',
                                             default_flavor_name=flavor)
        resp = self.controller.create_app(self.app.app)

        # print the cluster instances after error
        app_post = self.controller.show_apps()

        # look for app
        found_app = self.app.exists(app_post)

        expect_equal(found_app, True, 'find app')
        assert_expectations()

    def tearDown(self):
        self.controller.delete_app(self.app.app)
        self.controller.delete_cluster(self.cluster.cluster)
        self.controller.delete_developer(self.developer.developer)

if __name__ == '__main__':
    suite = unittest.TestLoader().loadTestsFromTestCase(tc)
    sys.exit(not unittest.TextTestRunner().run(suite).wasSuccessful())

