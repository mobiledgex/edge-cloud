package com.mobiledgex.matchingengine;

import android.content.Context;
import android.location.Location;
import android.telephony.TelephonyManager;
import android.util.Log;

import java.lang.invoke.MethodHandles;
import java.util.ArrayList;
import java.util.concurrent.ExecutionException;

import distributed_match_engine.AppClient;
import distributed_match_engine.LocOuterClass;
import distributed_match_engine.AppClient.QosPosition;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertTrue;

public class MockUtils {
    private final static String TAG = "MockUtils";

    public static String getCarrierName(Context context) {
        TelephonyManager telManager = (TelephonyManager)context.getSystemService(Context.TELEPHONY_SERVICE);
        String networkOperatorName = telManager.getNetworkOperatorName();
        return networkOperatorName;
    }
    public static Location createLocation(String provider, double longitude, double latitude) {
        Location loc = new Location(provider);
        loc.setLongitude(longitude);
        loc.setLatitude(latitude);
        loc.setTime(System.currentTimeMillis());
        return loc;
    }

    public static ArrayList<QosPosition> createQosPositionArray(Location firstLocation, double direction_degrees, double totalDistanceKm, double increment) {
        // Create a bunch of locations to get QOS information. Server is to be proxied by the DME server.
        ArrayList<QosPosition> positions = new ArrayList<>();

        Location lastLocation = firstLocation;
        long id = 1;
        double traverse;

        // First point:
        AppClient.QosPosition firstPositionKpi = AppClient.QosPosition.newBuilder()
                .setPositionid(id)
                .setGpsLocation(androidToMessageLoc(firstLocation))
                .build();
        positions.add(firstPositionKpi);

        // Everything in between:
        for (traverse = increment; traverse + increment < totalDistanceKm - increment; traverse += increment, id++) {
            Location next = MockUtils.createLocation(lastLocation.getLongitude(), lastLocation.getLatitude(), direction_degrees, increment);
            QosPosition np = AppClient.QosPosition.newBuilder()
                    .setPositionid(id)
                    .setGpsLocation(androidToMessageLoc(next))
                    .build();
            positions.add(np);
            lastLocation = next;
        }

        // Last point, if needed.
        if (traverse < totalDistanceKm) {
            Location lastLoc = MockUtils.createLocation(lastLocation.getLongitude(), lastLocation.getLatitude(), direction_degrees, totalDistanceKm);
            QosPosition lastPosition = QosPosition.newBuilder()
                    .setPositionid(id)
                    .setGpsLocation(androidToMessageLoc(lastLoc))
                    .build();
            positions.add(lastPosition);
        }

        return positions;
    }

    /**
     * Returns a destination long/lat as a Location object, along direction (in degrees), some distance in kilometers away.
     *
     * @param longitude_src
     * @param latitude_src
     * @param direction_degrees
     * @param kilometers
     * @return
     */
    public static Location createLocation(double longitude_src, double latitude_src, double direction_degrees, double kilometers) {
        double direction_radians = direction_degrees * (Math.PI / 180);

        // Provider is static class name:
        Location newLoc = new Location(MethodHandles.lookup().lookupClass().getName());

        // Not accurate:
        newLoc.setLongitude(longitude_src + kilometers * Math.cos(direction_radians));
        newLoc.setLatitude(latitude_src + kilometers * Math.sin(direction_radians));

        return newLoc;
    }

    public static LocOuterClass.Loc androidToMessageLoc(Location location) {
        return LocOuterClass.Loc.newBuilder()
                .setLatitude(location.getLatitude())
                .setLongitude(location.getLongitude())
                .setTimestamp(LocOuterClass.Timestamp.newBuilder()
                        .setSeconds(System.currentTimeMillis()/1000)
                        .build())
                .build();
    }
    public static AppClient.RegisterClientRequest createMockRegisterClientRequest(String developerName,
                                                                           String appName,
                                                                           MatchingEngine me) {
        return AppClient.RegisterClientRequest.newBuilder()
                .setVer(0)
                .setDevName(developerName) // From signing certificate?
                .setAppName(appName)
                .setAppVers("1.0")
                .build();
    }

    public static AppClient.FindCloudletRequest createMockFindCloudletRequest(String networkOperatorName, MatchingEngine me, Location location) {
        return AppClient.FindCloudletRequest.newBuilder()
                .setVer(0)
                .setSessionCookie(me.getSessionCookie() == null ? "" : me.getSessionCookie())
                .setCarrierName(networkOperatorName)
                .setGpsLocation(androidToMessageLoc(location))
                .build();
    }

    public static AppClient.VerifyLocationRequest createMockVerifyLocationRequest(String networkOperatorName, MatchingEngine me, Location location) {

        return AppClient.VerifyLocationRequest.newBuilder()
                .setVer(0)
                .setSessionCookie(me.getSessionCookie() == null ? "" : me.getSessionCookie())
                .setCarrierName(networkOperatorName)
                .setGpsLocation(androidToMessageLoc(location))
                .setVerifyLocToken(me.getTokenServerToken() == null ? "" : me.getTokenServerToken())
                .build();

    }

    public static AppClient.GetLocationRequest createMockGetLocationRequest(String networkOperatorName, MatchingEngine me) {
        return AppClient.GetLocationRequest.newBuilder()
                .setVer(0)
                .setSessionCookie(me.getSessionCookie() == null ? "" : me.getSessionCookie())
                .setCarrierName(networkOperatorName)
                .build();
    }

    public static AppClient.AppInstListRequest createMockAppInstListRequest(String networkOperatorName, MatchingEngine me, Location location) {
        return AppClient.AppInstListRequest.newBuilder()
                .setVer(0)
                .setSessionCookie(me.getSessionCookie() == null ? "" : me.getSessionCookie())
                .setCarrierName(networkOperatorName)
                .setGpsLocation(androidToMessageLoc(location))
                .build();
    }

    public static AppClient.DynamicLocGroupRequest createMockDynamicLocGroupRequest(MatchingEngine me, String userData) {

        return AppClient.DynamicLocGroupRequest.newBuilder()
                .setVer(0)
                .setSessionCookie(me.getSessionCookie() == null ? "" : me.getSessionCookie())
                .setLgId(1)
                .setCommType(AppClient.DynamicLocGroupRequest.DlgCommType.DLG_SECURE)
                .setUserData(userData == null ? "" : userData)
                .build();
    }
}
