#!/bin/sh

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
EOF

gofmt $OUT.tmp > $OUT
rm $OUT.tmp
