package main

import (
	"testing"
)

func checkKey(t *testing.T, keytype KeyType, id string, want string) {
	str := GetObjStoreKey(keytype, id)
	if str != want {
		t.Errorf("key %d/%s: wanted %s but got %s", keytype, id, want, str)
	}
}

func TestGetDbKey(t *testing.T) {
	InitRegion(1)
	checkKey(t, RegionType, "key1", "1/0/key1")
	checkKey(t, AppType, "key2", "1/1/key2")
	InitRegion(2)
	checkKey(t, CloudletType, "key3", "2/2/key3")
	checkKey(t, DeveloperType, "key4", "2/3/key4")
	InitRegion(5)
	checkKey(t, OperatorType, "key5", "5/4/key5")
	checkKey(t, AppInstType, "", "5/5/")
}
