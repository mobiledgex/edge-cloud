#
# Copyright 2018 Mobiledgex, Inc.
#

CXX = g++
CPPFLAGS += -I.
CPPFLAGS += -I../common
CXXFLAGS += -std=c++17
LDFLAGS += -L/usr/local/lib \
           -ldl\
           -lcurl

all: sample_client

sample_client: sample_client.o
	$(CXX) $^ $(LDFLAGS) $(CPPFLAGS) -L. -o $@

clean:
	rm -rf *.o sample_client
