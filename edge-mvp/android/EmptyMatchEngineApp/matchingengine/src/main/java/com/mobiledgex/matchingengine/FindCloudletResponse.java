package com.mobiledgex.matchingengine;

public class FindCloudletResponse {
    public long version = 0;
    public String uri;
    public byte[] server;
    public int port;
    public GPSLocation loc;

    public enum Find_Status {
        FIND_UNKNOWN(0),
        FIND_FOUND(1),
        FIND_NOTFOUND(2);
        private int find_status;
        Find_Status(final int find_status) {
            this.find_status = find_status;
        }

        public int getStatus() {
            return find_status;
        }

        // Helper to get corresponding enum value from int value.
        public static Find_Status forNumber(int value) {
            switch (value) {
                case 0: return FIND_UNKNOWN;
                case 1: return FIND_FOUND;
                case 2: return FIND_NOTFOUND;
                default: return null;
            }
        }
    }
    public Find_Status status;
    public String token = "";

    FindCloudletResponse(long version, String uri, byte[] server, int port, GPSLocation loc, Find_Status status, String token) {
        this.version = version;
        this.uri = uri;
        this.server = server;
        this.port = port;
        this.loc = loc;
        this.status = status;
        this.token = token;
    }
}
