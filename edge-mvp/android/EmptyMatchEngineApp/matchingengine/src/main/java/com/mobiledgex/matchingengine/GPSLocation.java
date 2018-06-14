package com.mobiledgex.matchingengine;

public class GPSLocation {
    private double lng;
    private double lat;
    private long timestamp_seconds;
    private int timestamp_nano;

    GPSLocation(double lng, double lat, long timestamp, int nano) {
        if (lng < -180d || lng > 180) {
            throw new IllegalArgumentException("Longitude out of range: " + lng);
        }
        if (lat < -90d || lat > 90d) {
            throw new IllegalArgumentException("Latitude out of range: " + lat);
        }
        this.lng = lng;
        this.lat = lat;
    }

    public double getLong() {
        return lng;
    }

    public void setLong(double lng) {
        this.lng = lng;
    }

    public double getLat() {
        return lat;
    }

    public void setLat(double lat) {
        this.lat = lat;
    }

    public long getTimestampSeconds() {
        return timestamp_seconds;
    }

    public void setTimestampSeconds(long timestamp) {
        this.timestamp_seconds = timestamp;
    }

    public int getTimestamp_nano() {
        return timestamp_nano;
    }

    public void setTimestamp_nano(int timestamp_nano) {
        this.timestamp_nano = timestamp_nano;
    }
}
