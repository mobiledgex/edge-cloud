# Upgrade of the protobuf data in etcd

Every now and again your build is going to fail with an error message that looks like this:

```
======WARNING=======
Current data model hash(HASH_b0497a835ca283bda2cf00778783a508) doesn't match the latest supported one(HASH_52e6980599cd59bbbd0de8d5f4d53d4b).
This is due to an unsupported change in the key of some objects in a .proto file.
In order to ensure a smooth upgrade for the production environment please make sure to add the following to version.proto file:

enum VersionHash {
	...
	HASH_52e6980599cd59bbbd0de8d5f4d53d4b = 2;
	HASH_b0497a835ca283bda2cf00778783a508 = 3 [(protogen.upgrade_func) = "sample_upgrade_function"]; <<<===== Add this line
	...
}

Implementation of "sample_upgrade_function" should be added tp edge-cloud/upgrade/upgrade-types.go

NOTE: If no upgrade function is needed don't need to add "[(protogen.upgrade_func) = "sample_upgrade_function];" to
the VersionHash enum.

A unit test data for the automatic unit test of the upgrade function should be added to testutil/upgrade_test_data.go
   - PreUpgradeData - what key/value objects are trying to be upgraded
   - PostUpgradeData - what the resulting object store should look like
====================
```

The error itself if pretty explanatory on what to do, but "why?" do you ask...well, read on.

## Why???

Controller stores the data in etcd database as key/value pairs. The object key as defined in protobuf message with a `[(gogoproto.nullable) = false];` option. This field in a message is going to get serialized into a key for the key/value pair. If the key changes the old entry in the etcd database won't be accessible anymore by the controller which expects the key to be different. To address this we need to allow the controller to upgrade the etcd database key/value pairs to match the new data model.

## How

To detect the change MEX protobuf plugin hashes all the key messages to generate a unique identifier for the data model. If something changed in the key, which might create a problem retrieving data from existing etcd, the hash won't match anymore and the error as shown above is displayed. 

Upgrade functions work on the raw data from etcd, so you can unmarshal it into any structure you need to perform the upgrade.
Also adding test data will allow you to automatically unit-test your upgrade function.

## Ok...what do I do

Follow the directions provided above to implement the conversion of the objects.
Also, there is a special way to invoke the upgrade on the running etcd:

By default if the controller starts up and detects a mismatch between the data-model hash in the etcd and current running hash it will fail to start. In order to trigger an upgrade you need to start the controller with `-autoUpgrade` command line argument. This will call the upgrade functions to convert the etcd to the match the new data-model before running the controller code. 

NOTE: for the development if you don't care about the changing structure of the objects you can run controller with `-skipVersionCheck` command line argument to bypass the upgrade and version check altogether, but be careful about possible implications of the controller unable to operate on parts of etcd.