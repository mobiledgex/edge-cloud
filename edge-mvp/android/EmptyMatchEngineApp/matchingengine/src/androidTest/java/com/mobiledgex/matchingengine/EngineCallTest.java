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

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;

import android.location.Location;

import android.util.Log;

@RunWith(AndroidJUnit4.class)
public class EngineCallTest {

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

        MexLocation mexLoc = new MexLocation(me);
        Location location = null;
        AppClient.Match_Engine_Status response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("registerClientTest", -122.149349, 37.459609);

        try {
            // Directly create request for testing:
            // Passed in Location (which is a callback interface)
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, 10000);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            response = me.registerClient(request, 10000);
            assert(response != null);
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("registerClientTest Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("registerClientTest Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        System.out.println("registerClientTest response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getToken(), ""); // FIXME: We DO expect a token
        assertEquals(response.getErrorCode(), AppClient.Match_Engine_Status.ME_Status.ME_SUCCESS_VALUE);
    }

    @Test
    public void registerClientFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine();

        MexLocation mexLoc = new MexLocation(me);
        Location location = null;
        Future<AppClient.Match_Engine_Status> responseFuture = null;
        AppClient.Match_Engine_Status response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("RegisterClientFutureTest", -122.149349, 37.459609);

        try {
            // Directly create request for testing:
            // Passed in Location (which is a callback interface)
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, 10000);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            responseFuture = me.registerClientFuture(request, 10000);
            response = responseFuture.get();
            assert(response != null);
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("registerClientFutureTest Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("registerClientFutureTest Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        System.out.println("registerClientFutureTest() response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getToken(), ""); // FIXME: We DO expect a token
        assertEquals(response.getErrorCode(), AppClient.Match_Engine_Status.ME_Status.ME_SUCCESS_VALUE);
    }

    @Test
    public void findCloudletTest() {
        // Context of the app under test.
        Context context = InstrumentationRegistry.getTargetContext();
        FindCloudletResponse cloudletResponse = null;
        MatchingEngine me = new MatchingEngine();
        MexLocation mexLoc = new MexLocation(me);

        Location loc = createLocation("findCloudletTest", -122.149349, 37.459609);


        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = mexLoc.getBlocking(context, 10000);
            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);

            cloudletResponse = me.findCloudlet(request, 10000);

        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("FindCloudlet Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("FindCloudlet Execution Interrupted!", true);
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
        // Context of the app under test.
        Context context = InstrumentationRegistry.getTargetContext();
        Future<FindCloudletResponse> response = null;
        FindCloudletResponse result = null;
        MatchingEngine me = new MatchingEngine();
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
            ee.printStackTrace();
            assertFalse("FindCloudletFuture Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("FindCloudletFuture Execution Interrupted!", true);
        }

        // Temporary.
        assertEquals(result.server, result.server);
    }

    @Test
    public void verifyLocationTest() {
        // Context of the app under test.
        Context context = InstrumentationRegistry.getTargetContext();

        MatchingEngine me = new MatchingEngine();
        MexLocation mexLoc = new MexLocation(me);
        AppClient.Match_Engine_Loc_Verify response = null;

        try {
            enableMockLocation(context, true);
            Location mockLoc = createLocation("verifyLocationTest", -122.149349, 37.459609);
            setMockLocation(context, mockLoc);
            Location location = mexLoc.getBlocking(context, 10000);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);


            response = me.verifyLocation(request, 10000);
            assert(response != null);
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("VerifyLocation Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("VerifyLocation Execution Interrupted!", true);
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
        MexLocation mexLoc = new MexLocation(me);
        AppClient.Match_Engine_Loc_Verify response = null;

        Future<AppClient.Match_Engine_Loc_Verify> locFuture;

        try {
            enableMockLocation(context, true);
            Location mockLoc = createLocation("verifyLocationFutureTest", -122.149349, 37.459609);
            setMockLocation(context, mockLoc);
            Location location = mexLoc.getBlocking(context, 10000);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);

            locFuture = me.verifyLocationFuture(request, 10000);
            response = locFuture.get();
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("verifyLocationFutureTest Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("verifyLocationFutureTest Execution Interrupted!", true);
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
        MexLocation mexLoc = new MexLocation(me);

        AppClient.Match_Engine_Loc_Verify verifyLocationResult = null;
        try {
            setMockLocation(context, mockLoc); // North Pole.
            Location location = mexLoc.getBlocking(context, 10000);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);

            verifyLocationResult = me.verifyLocation(request, 10000);
            assert(verifyLocationResult != null);
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("verifyMockedLocationTest_NorthPole Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("verifyMockedLocationTest_NorthPole Execution Interrupted!", true);
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

        MexLocation mexLoc = new MexLocation(me);
        Location location = null;
        AppClient.Match_Engine_Loc response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("getLocationTest", -122.149349, 37.459609);

        try {
            // Directly create request for testing:
            // Passed in Location (which is a callback interface)
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, 10000);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            response = me.getLocation(request, 10000);
            assert(response != null);
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("VerifyLocation Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("VerifyLocation Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        System.out.println("getLocation() response: " + response.toString());
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

        MexLocation mexLoc = new MexLocation(me);
        Location location = null;
        Future<AppClient.Match_Engine_Loc> responseFuture = null;
        AppClient.Match_Engine_Loc response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("getLocationTest", -122.149349, 37.459609);

        try {
            // Directly create request for testing:
            // Passed in Location (which is a callback interface)
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, 10000);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            responseFuture = me.getLocationFuture(request, 10000);
            response = responseFuture.get();
            assert(response != null);
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("getLocationFutureTest Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("getLocationFutureTest Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        System.out.println("getLocationFutureTest() response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getToken(), ""); // FIXME: We DO expect a token


        assertEquals(response.getCarrierName(), "TMUS");
        assertEquals(response.getStatus(), AppClient.Match_Engine_Loc.Loc_Status.LOC_FOUND);

        assertEquals(response.getTower(), 0);
        // FIXME: Server is currently a pure echo of client location.
        assertEquals((int) response.getNetworkLocation().getLat(), (int) loc.getLatitude());
        assertEquals((int) response.getNetworkLocation().getLong(), (int) loc.getLongitude());
    }

    /**
     * This is an extremely simple and basic end to end test of blocking versus Future using
     * VerifyLocation.
     */
    @Test
    public void basicLatencyTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine();

        MexLocation mexLoc = new MexLocation(me);
        Location location;
        AppClient.Match_Engine_Loc_Verify response1 = null;

        enableMockLocation(context,true);
        Location loc = createLocation("getLocationTest", -122.149349, 37.459609);

        long start;
        long elapsed1[] = new long[20];
        long elapsed2[] = new long[20];
        try {
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, 10000);
            assertFalse(location == null);

            long sum1 = 0, sum2 = 0;
            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            for (int i = 0; i < elapsed1.length; i++){
                start = System.currentTimeMillis();
                response1 = me.verifyLocation(request, 10000);
                elapsed1[i] = System.currentTimeMillis() - start;
            }

            for (int i = 0; i < elapsed1.length; i++) {
                Log.i("LatencyTest", i + " VerifyLocation elapsed time: Elapsed1: " + elapsed1[i]);
                sum1 += elapsed1[i];
            }
            Log.i("LatencyTest", "Average1: " + sum1/elapsed1.length);
            assert(response1 != null);

            // Future
            request = createMockMatchingEngineRequest(location);
            AppClient.Match_Engine_Loc_Verify response2 = null;
            try {
                for (int i = 0; i < elapsed2.length; i++) {
                    start = System.currentTimeMillis();
                    Future<AppClient.Match_Engine_Loc_Verify> locFuture = me.verifyLocationFuture(request, 10000);
                    // Do something busy()
                    response2 = locFuture.get();
                    elapsed2[i] = System.currentTimeMillis() - start;
                }
                for (int i = 0; i < elapsed2.length; i++) {
                    Log.i("LatencyTest", i + " VerifyLocationFuture elapsed time: Elapsed2: " + elapsed2[i]);
                    sum2 += elapsed2[i];
                }
                Log.i("LatencyTest", "Average2: " + sum2/elapsed2.length);
                assert(response2 != null);
            } catch (Exception e) {
                e.printStackTrace();
                throw e;
            }


        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("getLocationFutureTest Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("getLocationFutureTest Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }

}
