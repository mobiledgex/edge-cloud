#!/bin/sh

REPO=$1
OUT=version.go
# if no local/branch commits beyond master, master and head will be the same
BUILD_MASTER=`git describe --tags master`
BUILD_HEAD=`git describe --tags --dirty=+`
BUILD_AUTHOR=`git config user.name`
DATE=`date`

cat <<EOF > $OUT.tmp
package version

var BuildMaster = "$BUILD_MASTER"
var BuildHead = "$BUILD_HEAD"
var BuildAuthor = "$BUILD_AUTHOR"
var BuildDate = "$DATE"

func ${REPO}BuildProps(prefix string) map[string]string {
	m := map[string]string{
		prefix + "BuildMaster": BuildMaster,
		prefix + "BuildHead":   BuildHead,
		prefix + "BuildDate":   BuildDate,
	}
	if BuildAuthor != "" {
		m[prefix + "BuildAuthor"] = BuildAuthor
	}
	return m
}
EOF

gofmt $OUT.tmp > $OUT
rm $OUT.tmp
