package testutil

type KVPair struct {
	Key string
	Val string
}

var PreUpgradeData = map[string][]KVPair{
	"UpgradeMexSaltExample": {
		KVPair{
			Key: "_testdatakey/key1",
			Val: "val1",
		},
		KVPair{
			Key: "_testdatakey/key2",
			Val: "val2",
		},
	},
	"UpgradeFuncExample": {
		KVPair{
			Key: "_testdatakey/key3",
			Val: "val3",
		},
		KVPair{
			Key: "_testdatakey/key4",
			Val: "val4",
		},
	},
	"UpgradeFuncReplaceEverything": {
		KVPair{
			Key: "_testdatakey/key5",
			Val: "val5",
		},
		KVPair{
			Key: "_testdatakey/key6",
			Val: "val6",
		},
	},
}

var PostUpgradeData = map[string][]KVPair{
	"UpgradeMexSaltExample": {
		KVPair{
			Key: "_testdatakey/key1",
			Val: "val1",
		},
		KVPair{
			Key: "_testdatakey/key2",
			Val: "val2",
		},
	},
	"UpgradeFuncExample": {
		KVPair{
			Key: "_testdatakey/key3",
			Val: "val33",
		},
		KVPair{
			Key: "_testdatakey/key4",
			Val: "val44",
		},
	},
	"UpgradeFuncReplaceEverything": {
		KVPair{
			Key: "val5",
			Val: "_testdatakey/key5",
		},
		KVPair{
			Key: "val6",
			Val: "_testdatakey/key6",
		},
	},
}
