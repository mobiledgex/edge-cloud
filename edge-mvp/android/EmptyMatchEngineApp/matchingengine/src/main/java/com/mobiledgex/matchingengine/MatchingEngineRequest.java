package com.mobiledgex.matchingengine;

import distributed_match_engine.AppClient;

public class MatchingEngineRequest {
    public AppClient.Match_Engine_Request matchEngineRequest;
    public String host;
    public int port;

    public MatchingEngineRequest(AppClient.Match_Engine_Request request, String host, int port) {
        matchEngineRequest = request;
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
