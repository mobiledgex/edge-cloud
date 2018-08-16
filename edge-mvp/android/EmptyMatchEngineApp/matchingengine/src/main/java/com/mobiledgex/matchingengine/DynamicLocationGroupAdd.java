package com.mobiledgex.matchingengine;

import distributed_match_engine.AppClient;

public class DynamicLocationGroupAdd {
    public AppClient.DynamicLocGroupAdd dynamicLocGroupAdd;
    public String host;
    public int port;

    public DynamicLocationGroupAdd(AppClient.DynamicLocGroupAdd request, String host, int port) {
        dynamicLocGroupAdd = request;
        this.host = host;
        this.port = port;
    }

    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }
}
