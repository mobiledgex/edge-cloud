### Flavor Testcases
* add flavor with max values - test_flavorAdd_maxValues.py
* add flavor with min values - test_flavorAdd_minValues.py
* add 100 flavors - test_flavorAdd_100.py
* add flavor and check every controller - test_flavorAdd_multiControllers.py
* add flavor failes with greater than max values - test_flavorAdd_largerThanMaxValues.py
* add flavor fails with same name - test_flavorAdd_sameName.py
* add flavor fails with invalid name -  test_flavorAdd_invalidName.py
* add flavor fails with empty/unknown name - test_flavorAdd_noName.py
* add flavor fails with name only - test_flavorAdd_nameOnly.py
* add flavor fails with ram=0 - test_flavorAdd_ram0.py
* add flavor fails with vcpus=0 - test_flavorAdd_vcpus0.py
* add flavor fails with disk=0 - test_flavorAdd_disk0.py
* add flavor fails with invalid ram/vcpus/disk - test_flavorAdd_invalidParms.py
* delete flavor fails with unknown name - test_flavorDelete_unknown.py
* delete flavor with name only - test_flavorDelete_nameOnly.py
* delete flavor with name and wrong ram/vcpus/disk - test_flavorDelete_wrongParms.py
* delete flavor fails before deleting cluster flavor - test_flavorDelete_beforeClusterFlavor.py
* update flavor unsupported - test_flavorUpdate_notSupported.py
* show flavor single - test_flavorShow_single.py
* show flavor by ram/vcpus/disk - test_flavorShow_ramVcupsDisk.py
