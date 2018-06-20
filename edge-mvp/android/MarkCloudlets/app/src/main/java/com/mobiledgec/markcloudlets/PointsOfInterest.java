package com.mobiledgec.markcloudlets;

public class PointsOfInterest {
    private double latitude;
    private double longitude;
    private String cloudlets;
    private String operators;

    public PointsOfInterest(double latitude, double longitude, String cloudlets, String operators){
        this.setLatitude(latitude);
        this.setLongitude(longitude);
        this.setCloudlets(cloudlets);
        this.setOperators(operators);
    }

    public double getLatitude() {
        return latitude;
    }

    public void setLatitude(double latitude) {
        this.latitude = latitude;
    }

    public double getLongitude() {
        return longitude;
    }

    public void setLongitude(double longitude) {
        this.longitude = longitude;
    }

    public String getCloudlets() {
        return cloudlets;
    }

    public void setCloudlets(String cloudlets) {
        this.cloudlets = cloudlets;
    }

    public String getOperators() {
        return operators;
    }

    public void setOperators(String operators) {
        this.operators = operators;
    }
}
