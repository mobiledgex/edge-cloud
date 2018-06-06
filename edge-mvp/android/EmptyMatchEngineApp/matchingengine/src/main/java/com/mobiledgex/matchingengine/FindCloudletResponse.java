package com.mobiledgex.matchingengine;

public class FindCloudletResponse {
    public byte[] server;
    public int port;
    public String service;

    public GPSLocation loc;

    public long version = 0;

    FindCloudletResponse(long version, byte[] server, int port, GPSLocation loc) {
        this.version = version;
        this.server = server;
        this.port = port;
        this.loc = loc;
    }
}
