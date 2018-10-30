#
# Copyright 2018 Mobiledgex, Inc.
#

CXX = g++
CPPFLAGS += -I.
CXXFLAGS += -std=c++11
LDFLAGS += -L/usr/local/lib \
           -ldl\
           -lcurl

# Run "brew install coreutils" to get "grealpath" for protoc's requirement of absolute paths
ABSOLUTEPATH := $(shell grealpath ../../..)

all: sample_client

sample_client: test_credentials.o sample_client.o
	$(CXX) $^ $(LDFLAGS) -I. -L. -o $@

clean:
	rm -rf *.o sample_client
