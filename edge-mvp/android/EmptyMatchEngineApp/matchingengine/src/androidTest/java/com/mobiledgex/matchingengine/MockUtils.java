package com.mobiledgex.matchingengine;

import android.content.Context;
import android.location.Location;
import android.telephony.TelephonyManager;
import android.util.Log;

import java.util.concurrent.ExecutionException;

import distributed_match_engine.AppClient;
import distributed_match_engine.LocOuterClass;

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

    public static LocOuterClass.Loc androidToMessageLoc(Location location) {
        return LocOuterClass.Loc.newBuilder()
                .setLat(location.getLatitude())
                .setLong(location.getLongitude())
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
                .setCommType(AppClient.DynamicLocGroupRequest.DlgCommType.DlgSecure)
                .setUserData(userData == null ? "" : userData)
                .build();
    }
}
