package com.mobiledgex.matchingengine;

public class FindCloudletResponse {
    public long version = 0;
    public String uri;
    public byte[] server;
    public int port;
    public GPSLocation loc;
    public boolean status;
    public String token = "";

    FindCloudletResponse(long version, String uri, byte[] server, int port, GPSLocation loc, boolean status, String token) {
        this.version = version;
        this.uri = uri;
        this.server = server;
        this.port = port;
        this.loc = loc;
        this.status = status;
        this.token = token;
    }
}
