#
# Copyright 2018 Mobiledgex, Inc.
#

CXX = g++
CPPFLAGS += -I.
CPPFLAGS += -I../common
CXXFLAGS += -std=c++11
LDFLAGS += -L/usr/local/lib \
           -ldl\
           -lcurl

all: sample_client

sample_client: ../common/test_credentials.o sample_client.o
	$(CXX) $^ $(LDFLAGS) $(CPPFLAGS) -L. -o $@

clean:
	rm -rf *.o sample_client
