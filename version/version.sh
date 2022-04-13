#!/bin/sh
# Copyright 2022 MobiledgeX, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


REPO=$1
OUT=version.go
# if no local/branch commits beyond master, master and head will be the same
BUILD_MASTER=`git describe --tags origin/master`
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
