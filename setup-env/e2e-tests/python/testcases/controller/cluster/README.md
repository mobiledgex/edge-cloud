### Cluster Testcases
* add cluster with name and no default_flavor - test_clusterAdd_noDefaultFlavor.py
* add cluster with name and default_flavor    - test_clusterAdd_defaultFlavor.py
* add cluster and check every controller - test_clusterAddMultiControllers.py
* add 100 clusters - test_clusterAdd_100.py
* add cluster fails with no name  - test_clusterAdd_noName.py
* add cluster fails starting with AutoCluster - test_clusterAdd_AutoCluster.py
* add cluster fails with underscore or not matching "^[0-9a-zA-Z][-0-9a-zA-Z.]*$" - test_clusterAdd_InvalidClusterName.py
* delete cluster fails withn unknown name - test_clusterDelete_nameNotFound.py
* delete cluster fails with no name - test_clusterDelete_noName.py
* delete cluster fails with invalid name - test_clusterDelete_invalidName.py
* delete cluster fails before deleting clusterinst - test_clusterDelete_beforeClusterInst.py
* delete cluster fails before app - test_clusterDelete_beforeApp.py
* update cluster fails not supported - test_clusterUpdate_notSupported.py

### Cluster Instance Testcases
* add clusterinst with flavor name - test_clusterInstAdd.py
* add clusterinst with default flavor and no flavor name - test_clusterInstAdd_noFlavor.py
* add clusterinst and check every controller - test_clusterInstAddMultiControllers.py
* add 100 cluster and cluster instances - test_clusterInstAdd_100.py
* add clusterinst with liveness=1 - test_clusterInstAdd_liveness1.py
* add clusterinst with liveness=2 - test_clusterInstAdd_liveness2.py
* add clusterinst fails with cluster not found - test_clusterInstAdd_clusterNotFound.py
* add clusterinst twice fails with key exists - test_clusterInstAdd_keyExists.py
* add clusterinst fails with no default flavor and no flavor name - test_clusterInstAdd_noDefaultFlavor_noFlavor.py
* add clusterinst fails with missing parms - test_clusterInstAdd_missingParms.py
* add clusterinst fails with cloudletInfo not found - test_clusterInstAdd_cloudletInfoNotFound.py
* add clusterinst fails with cluster not found - test_clusterInstAdd_flavorNotExist.py
* add clusterinst fails when operator does not match cloudlet - test_clusterInstAdd_operatorNotMatchCloudlet.py
* update clusterinst fails with not supported - test_clusterInstUpdate_notSupported.py
* delete with and without flavor-name - test_clusterInstDelete.py
* delete fails with key not found - test_clusterInstDelete_notFound.py
* delete clusterinst before app inst - test_clusterInstDelete_beforeAppInst.py

