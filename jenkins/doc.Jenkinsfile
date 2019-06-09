pipeline {
    options {
        timeout(time: 10, unit: 'MINUTES')
    }
    agent any
    environment {
        BUILD_TAG = sh(returnStdout: true, script: 'date +"%Y-%m-%d" | tr -d "\n"')
    }
    stages {
        stage('Checkout') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    checkout scm
                }
            }
        }
        stage('Generate protoc-gen-swagger') {
            steps {
                dir(path: 'go/src/github.com/grpc-ecosystem') {
                    git url: 'git@github.com:grpc-ecosystem/grpc-gateway'
                }
                dir(path: 'go/src/github.com/grpc-ecosystem/grpc-gateway') {
                    sh label: 'switch to mobiledgex fork',
                       script: 'git remote add mobiledgex git@github.com:mobiledgex/grpc-gateway'
                }
                dir(path: 'go/src/github.com/grpc-ecosystem/grpc-gateway') {
                    sh label: 'pull latest changes from fork',
                       script: 'git pull mobiledgex master'
                }
                dir(path: 'go/src') {
                    sh label: 'build protoc-gen-swagger',
                       script: 'go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger'
                }
            }
        }
        stage('Generate docs') {
            steps {
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'git clean edge-cloud', script: 'git clean -f -d -x'
                }
                dir(path: 'go/src/github.com/mobiledgex/edge-cloud') {
                    sh label: 'make doc', script: '''#!/bin/bash
export PATH=$PATH:$HOME/go/bin:$WORKSPACE/go/bin
export GOPATH=$WORKSPACE/go
export GO111MODULE=on
go mod download
make doc
                    '''
                }
                rtUpload (
                    serverId: "artifactory",
                    spec:
                        """{
                            "files": [
                                {
                                    "pattern": "go/src/github.com/mobiledgex/edge-cloud/edgeproto/doc/*.json",
                                    "target": "build-artifacts/swagger-spec/${BUILD_TAG}/"
                                }
                            ]
                        }"""
                )
            }
        }
    }
}
