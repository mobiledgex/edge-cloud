### App Testcases
* Add 100 apps - test_appAdd_100.py
* Add app with empty cluster to create autocluster - test_appAdd_emptyCluster.py
* Add app with config - test_appAdd_config.py
* Add app with image path - test_appAdd_imagePath.py
* Add app with image type Docker and no/empty image path - test_appAdd_docker.py
* Add app that is not docker compliant but is converted to docker compliant - test_appAdd_dockerCompliant.py
* Add app with no ip_access - test_appAdd_noIpAccess.py
* Add app with IpAccessDedicatedOrShared and port 1 - test_appAdd_IpAccessDedicatedOrSharedPort1.py
* Add app with IpAccessDedicatedOrShared and port 65535 - test_appAdd_IpAccessDedicatedOrSharedPort65535.py
* Add app with IpAccessDedicatedOrShared and multiple ports - test_appAdd_IpAccessDedicatedOrSharedMulti.py
* Add app with IpAccessDedicated and port 1 - test_appAdd_IpAccessDedicatedPort1.py
* Add app with IpAccessDedicated and port 65535 - test_appAdd_IpAccessDedicated65535.py
* Add app with IpAccessDedicated and multiple ports - test_appAdd_IpAccessDedicatedPortMulti.py
* Add app with IpAccessShared and port 1 - test_appAdd_IpAccessSharedPort1.py
* Add app with IpAccessShared and port 65535 - test_appAdd_IpAccessSharedPort65535.py
* Add app with IpAccessShared and multiple ports - test_appAdd_IpAccessSharedMulti.py
* Add app fails with empty/missing name - test_appAdd_appNameEmpty.py
* Add app fails with cluster not found - test_appAdd_clusterNotFound.py
* Add app fails with no default flavor - test_appAdd_defaultFlavorEmpty.py
* Add app fails with default flavor not found - test_appAdd_defaultFlavorNotFound.py
* Add app fails with no developer - test_appAdd_developerEmpty.py
* Add app fails with developer not found - test_appAdd_developerNotFound.py
* Add app fails with flavor not found - test_appAdd_flavorNotFound.py
* Add app fails with image type only - test_appAdd_imageTypeOnly.py
* Add app fails when adding it twice - test_appAdd_keyExists.py
* Add app fails with no parms - test_appAdd_noParms.py
* Add app fails with invalid port - test_appAdd_portInvalidDigits.py
* Add app fails with port out of range - test_appAdd_portOutOfRange.py
* Add app fails with image type QCOW - test_appAdd_qcow.py
* Add app fails with unsupported protocol - test_appAdd_unsupportedPortProtocol.py
* Add app fails with Invalid Image Type - test_appAdd_invalidImageType.py
* Add app fails with invalid app name - test_appAdd_appNameInvalid.py
* Add app fails with invalid ip_access - test_appAdd_invalidIpAccess.py
* Add app fails with no access ports - test_appAdd_noPorts.py
* Add app fails with invalid port format - test_appAdd_invalidPortFormat.py
* Show apps querying by 1 parm - test_appShow_queryParms.py
* Delete app fails with Key Not Found - test_appDelete_keyNotFound.py


















