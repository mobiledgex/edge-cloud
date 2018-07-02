package com.mobiledgex.matchingengine;

import android.content.Context;
import android.support.test.InstrumentationRegistry;
import android.support.test.runner.AndroidJUnit4;

import com.google.android.gms.location.FusedLocationProviderClient;
import com.google.protobuf.ByteString;
import com.mobiledgex.matchingengine.util.MexLocation;

import org.junit.Test;
import org.junit.Before;
import org.junit.runner.RunWith;

import android.os.Build;

import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import distributed_match_engine.AppClient;
import distributed_match_engine.LocOuterClass;
import io.grpc.StatusRuntimeException;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertTrue;

import android.location.Location;
import android.util.Log;


@RunWith(AndroidJUnit4.class)
public class EngineCallTest {
    public static final String TAG = "EngineCallTest";
    public static final long GRPC_TIMEOUT_MS = 100;

    FusedLocationProviderClient fusedLocationClient;


    @Before
    public void grantPermissions() {

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            InstrumentationRegistry.getInstrumentation().getUiAutomation().executeShellCommand(
                    "pm grant " + InstrumentationRegistry.getTargetContext().getPackageName()
                            + " android.permission.READ_PHONE_STATE");
            InstrumentationRegistry.getInstrumentation().getUiAutomation().executeShellCommand(
                    "pm grant " + InstrumentationRegistry.getTargetContext().getPackageName()
                            + " android.permission.ACCESS_COARSE_LOCATION");
        }
    }

    /**
     * Enable or Disable MockLocation.
     * @param context
     * @param enableMock
     * @return
     */
    public boolean enableMockLocation(Context context, boolean enableMock) {
        if (fusedLocationClient == null) {
            fusedLocationClient = new FusedLocationProviderClient(context);
        }
        if (enableMock == false) {
            fusedLocationClient.setMockMode(false);
            return false;
        } else {
            fusedLocationClient.setMockMode(true);
            return true;
        }
    }

    public Location createLocation(String provider, double longitude, double latitude) {
        Location loc = new Location(provider);
        loc.setLongitude(longitude);
        loc.setLatitude(latitude);
        return loc;
    }

    /**
     * Utility Func. Single point mock location, fills in some extra fields. Does not calculate speed, nor update interval.
     * @param context
     * @param location
     */
    public void setMockLocation(Context context, Location location) throws InterruptedException {
        if (fusedLocationClient == null) {
            fusedLocationClient = new FusedLocationProviderClient(context);
        }

        location.setTime(System.currentTimeMillis());
        location.setElapsedRealtimeNanos(1000);
        location.setAccuracy(3f);
        fusedLocationClient.setMockLocation(location);
        try {
            Thread.sleep(100); // Give Mock a bit of time to take effect.
        } catch (InterruptedException ie) {
            throw ie;
        }
        fusedLocationClient.flushLocations();
    }

    public AppClient.Match_Engine_Request createMockMatchingEngineRequest(Location location) {
        AppClient.Match_Engine_Request request;

        // Directly create request for testing:
        LocOuterClass.Loc aLoc = LocOuterClass.Loc.newBuilder()
                .setLat(location.getLatitude())
                .setLong(location.getLongitude())
                .build();

        request = AppClient.Match_Engine_Request.newBuilder()
                .setVer(5)
                .setIdType(AppClient.Match_Engine_Request.IDType.MSISDN)
                .setId("")
                .setCarrierID(3l) // uint64 --> String? mnc, mcc?
                .setCarrierName("TMUS") // Mobile Network Carrier
                .setTower(0) // cid and lac (int)
                .setGpsLocation(aLoc)
                .setAppId(5011l) // uint64 --> String again. TODO: Clarify use.
                .setProtocol(ByteString.copyFromUtf8("http")) // This one is appId context sensitive.
                .setServerPort(ByteString.copyFromUtf8("1234")) // App dependent.
                .setDevName("EmptyMatchEngineApp") // From signing certificate?
                .setAppName("EmptyMatchEngineApp")
                .setAppVers("1") // Or versionName, which is visual name?
                .setToken("") // None.
                .build();

        return request;
    }

    @Test
    public void registerClientTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(true);

        MexLocation mexLoc = new MexLocation(me);
        Location location;
        AppClient.Match_Engine_Status response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("registerClientTest", -122.149349, 37.459609);

        try {
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            response = me.registerClient(request, GRPC_TIMEOUT_MS);
            assert(response != null);
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("registerClientTest: Execution Failed!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("registerClientTest: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("registerClientTest: Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        Log.i(TAG, "registerClientTest response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getToken(), ""); // FIXME: We DO expect a token
        assertEquals(response.getErrorCode(), AppClient.Match_Engine_Status.ME_Status.ME_SUCCESS_VALUE);
    }

    @Test
    public void registerClientFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(true);

        MexLocation mexLoc = new MexLocation(me);
        Location location;
        Future<AppClient.Match_Engine_Status> responseFuture;
        AppClient.Match_Engine_Status response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("RegisterClientFutureTest", -122.149349, 37.459609);

        try {
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            responseFuture = me.registerClientFuture(request, GRPC_TIMEOUT_MS);
            response = responseFuture.get();
            assert(response != null);
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("registerClientFutureTest: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("registerClientFutureTest: Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        Log.i(TAG, "registerClientFutureTest() response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getToken(), ""); // FIXME: We DO expect a token
        assertEquals(response.getErrorCode(), AppClient.Match_Engine_Status.ME_Status.ME_SUCCESS_VALUE);
    }

    @Test
    public void mexDisabledTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(false);
        MexLocation mexLoc = new MexLocation(me);

        Location loc = createLocation("findCloudletTest", -122.149349, 37.459609);
        boolean allRun = false;

        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);

            try {
                FindCloudletResponse cloudletResponse = me.findCloudlet(request, GRPC_TIMEOUT_MS);
            } catch (MissingRequestException mre) {
                // This is expected, request is missing.
                Log.i(TAG, "Expected exception for findCloudlet. Mex Disabled.");
            }
            try {
                AppClient.Match_Engine_Loc locResponse = me.getLocation(request, GRPC_TIMEOUT_MS);
            } catch (MissingRequestException mre) {
                // This is expected, request is missing.
                Log.i(TAG, "Expected exception for getLocation. Mex Disabled.");
            }
            try {
                AppClient.Match_Engine_Loc_Verify  locVerifyResponse = me.verifyLocation(request, GRPC_TIMEOUT_MS);
            } catch (MissingRequestException mre) {
                // This is expected, request is missing.
                Log.i(TAG, "Expected exception for verifyLocation. Mex Disabled.");
            }
            try {
                AppClient.Match_Engine_Status registerStatusResponse = me.registerClient(request, GRPC_TIMEOUT_MS);
            } catch (MissingRequestException mre) {
                // This is expected, request is missing.
                Log.i(TAG, "Expected exception for registerClient. Mex Disabled.");
            }
            allRun = true;
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("FindCloudlet: Execution Failed!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("FindCloudlet: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("FindCloudlet: Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        assertTrue("All requests must run with failures.", allRun);
    }

    @Test
    public void findCloudletTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        FindCloudletResponse cloudletResponse = null;
        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);

        Location loc = createLocation("findCloudletTest", -122.149349, 37.459609);


        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);

            cloudletResponse = me.findCloudlet(request, GRPC_TIMEOUT_MS);

        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("FindCloudlet: Execution Failed!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("FindCloudlet: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("FindCloudlet: Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        if (cloudletResponse != null) {
            // Temporary.
            assertEquals(cloudletResponse.server, cloudletResponse.server);
        } else {
            assertFalse("No findCloudlet response!", false);
        }
    }

    @Test
    public void findCloudletFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        Future<FindCloudletResponse> response;
        FindCloudletResponse result = null;
        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);

        Location loc = createLocation("findCloudletTest", -122.149349, 37.459609);

        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = mexLoc.getBlocking(context, 10000);
            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);

            response = me.findCloudletFuture(request, 10000);
            result = response.get();
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("FindCloudletFuture: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("FindCloudletFuture: Execution Interrupted!", true);
        }

        // Temporary.
        assertEquals(result.server, result.server);
    }

    @Test
    public void verifyLocationTest() {
        Context context = InstrumentationRegistry.getTargetContext();

        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);
        AppClient.Match_Engine_Loc_Verify response = null;

        try {
            enableMockLocation(context, true);
            Location mockLoc = createLocation("verifyLocationTest", -122.149349, 37.459609);
            setMockLocation(context, mockLoc);
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);


            response = me.verifyLocation(request, GRPC_TIMEOUT_MS);
            assert(response != null);
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("VerifyLocation: Execution Failed!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("VerifyLocation: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("VerifyLocation: Execution Interrupted!", true);
        } finally {
            enableMockLocation(context, false);
        }

        // Temporary.
        assertEquals(response.getVer(), 0);
        assertEquals(response.getToken(), "");
        assertEquals(response.getTowerStatus(), AppClient.Match_Engine_Loc_Verify.Tower_Status.UNKNOWN);
        assertEquals(response.getGpsLocationStatus(), AppClient.Match_Engine_Loc_Verify.GPS_Location_Status.LOC_MISMATCH);
    }

    @Test
    public void verifyLocationFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();

        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);
        AppClient.Match_Engine_Loc_Verify response = null;

        Future<AppClient.Match_Engine_Loc_Verify> locFuture;

        try {
            enableMockLocation(context, true);
            Location mockLoc = createLocation("verifyLocationFutureTest", -122.149349, 37.459609);
            setMockLocation(context, mockLoc);
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);

            locFuture = me.verifyLocationFuture(request, GRPC_TIMEOUT_MS);
            response = locFuture.get();
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("verifyLocationFutureTest: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("verifyLocationFutureTest: Execution Interrupted!", true);
        }

        // Temporary.
        assertEquals(response.getVer(), 0);
        assertEquals(response.getToken(), "");
        assertEquals(response.getTowerStatus(), AppClient.Match_Engine_Loc_Verify.Tower_Status.UNKNOWN);
        assertEquals(response.getGpsLocationStatus(), AppClient.Match_Engine_Loc_Verify.GPS_Location_Status.LOC_MISMATCH);
    }


    /**
     * Mocked Location test. Note that responsibility of verifying location is in the MatchingEngine
     * on the server side, not client side.
     */
    @Test
    public void verifyMockedLocationTest_NorthPole() {
        Context context = InstrumentationRegistry.getTargetContext();
        enableMockLocation(context,true);

        Location mockLoc = createLocation("verifyMockedLocationTest_NorthPole", 90d, 0d);


        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);

        AppClient.Match_Engine_Loc_Verify verifyLocationResult = null;
        try {
            setMockLocation(context, mockLoc); // North Pole.
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);

            verifyLocationResult = me.verifyLocation(request, GRPC_TIMEOUT_MS);
            assert(verifyLocationResult != null);
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("verifyMockedLocationTest_NorthPole: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("verifyMockedLocationTest_NorthPole: Execution Interrupted!", true);
        }

        // Temporary.
        assertEquals(verifyLocationResult.getVer(), 0);
        assertEquals(verifyLocationResult.getTowerStatusValue(), AppClient.Match_Engine_Loc_Verify.Tower_Status.UNKNOWN_VALUE);
        assertEquals(verifyLocationResult.getGpsLocationStatusValue(), AppClient.Match_Engine_Loc_Verify.GPS_Location_Status.LOC_MISMATCH_VALUE); // Based on test data.

        enableMockLocation(context,false);
    }

    @Test
    public void getLocationTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);
        Location location;
        AppClient.Match_Engine_Loc response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("getLocationTest", -122.149349, 37.459609);

        try {
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            response = me.getLocation(request, GRPC_TIMEOUT_MS);
            assert(response != null);
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("VerifyLocation: Execution Failed!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("VerifyLocation: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("VerifyLocation: Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        Log.i(TAG, "getLocation() response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getToken(), ""); // FIXME: We DO expect a token


        assertEquals(response.getCarrierName(), "TMUS");
        assertEquals(response.getStatus(), AppClient.Match_Engine_Loc.Loc_Status.LOC_FOUND);

        assertEquals(response.getTower(), 0);
        // FIXME: Server is currently a pure echo of client location.
        assertEquals((int) response.getNetworkLocation().getLat(), (int) loc.getLatitude());
        assertEquals((int) response.getNetworkLocation().getLong(), (int) loc.getLongitude());
    }

    @Test
    public void getLocationFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine();
        me.setMexLocationAllowed(true);

        MexLocation mexLoc = new MexLocation(me);
        Location location;
        Future<AppClient.Match_Engine_Loc> responseFuture;
        AppClient.Match_Engine_Loc response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("getLocationTest", -122.149349, 37.459609);

        try {
            // Directly create request for testing:
            // Passed in Location (which is a callback interface)
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            responseFuture = me.getLocationFuture(request, GRPC_TIMEOUT_MS);
            response = responseFuture.get();
            assert(response != null);
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("getLocationFutureTest: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("getLocationFutureTest: Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        Log.i(TAG, "getLocationFutureTest() response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getToken(), ""); // FIXME: We DO expect a token


        assertEquals(response.getCarrierName(), "TMUS");
        assertEquals(response.getStatus(), AppClient.Match_Engine_Loc.Loc_Status.LOC_FOUND);

        assertEquals(response.getTower(), 0);
        // FIXME: Server is currently a pure echo of client location.
        assertEquals((int) response.getNetworkLocation().getLat(), (int) loc.getLatitude());
        assertEquals((int) response.getNetworkLocation().getLong(), (int) loc.getLongitude());
    }
}
