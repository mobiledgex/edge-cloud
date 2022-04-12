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


GRPC_VERSION=v1.10.x

git clone -b $GRPC_VERSION --recursive -j8 https://github.com/grpc/grpc
cd /tmp/grpc
make 
make install
# php support
git submodule update --init
make grpc_php_plugin

cp /tmp/grpc/bins/opt/protobuf/protoc /usr/local/bin/

cd /tmp/grpc/third_party/protobuf
make
make install

cd /tmp
git clone -b $GRPC_VERSION --recursive https://github.com/grpc/grpc-java.git
cd /tmp/grpc-java/compiler
../gradlew java_pluginExecutable
