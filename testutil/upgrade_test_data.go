package testutil

type KVPair struct {
	Key string
	Val string
}

var PreUpgradeData = map[string][]KVPair{
	"TestUpgradeExample": {
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
	"TestUpgradeExample": {
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
