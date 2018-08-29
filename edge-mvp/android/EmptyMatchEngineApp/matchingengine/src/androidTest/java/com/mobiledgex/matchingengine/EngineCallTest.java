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

import java.io.IOException;
import java.util.UUID;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import distributed_match_engine.AppClient;
import distributed_match_engine.DynamicLocationGroup;
import distributed_match_engine.LocOuterClass;
import io.grpc.StatusRuntimeException;

import static distributed_match_engine.AppClient.IDTypes.IPADDR;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

import android.location.Location;
import android.telephony.TelephonyManager;
import android.util.Log;

import javax.net.ssl.SSLException;


@RunWith(AndroidJUnit4.class)
public class EngineCallTest {
    public static final String TAG = "EngineCallTest";
    public static final long GRPC_TIMEOUT_MS = 15000;

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

            // FIXME: Read application cert and keys.
            InstrumentationRegistry.getInstrumentation().getUiAutomation().executeShellCommand(
                    "pm grant " + InstrumentationRegistry.getTargetContext().getPackageName()
                            + " android.permission.READ_EXTERNAL_STORAGE");
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

    // Every call needs registration to be called first at some point.
    public void registerClient(String carrierName, MatchingEngine me, Location location) {
            AppClient.Match_Engine_Status registerResponse;
            MatchingEngineRequest regRequest = createMockMatchingEngineRequest(carrierName, me, location);
            try {
                registerResponse = me.registerClient(regRequest, GRPC_TIMEOUT_MS);
                assertEquals("Response SessionCookie should equal MatchingEngine SessionCookie",
                        registerResponse.getSessionCookie(), me.getSessionCookie());
            /*} catch (SSLException se) {
                Log.e(TAG, Log.getStackTraceString(se));
                assertTrue("SSLException registering client", false);*/
            } catch (ExecutionException ee) {
                Log.e(TAG, Log.getStackTraceString(ee));
                assertTrue("ExecutionException registering client", false);
            } catch (InterruptedException ioe) {
                Log.e(TAG, Log.getStackTraceString(ioe));
                assertTrue("InterruptedException registering client", false);
            }

    }

    public String getCarrierName(Context context) {
        TelephonyManager telManager = (TelephonyManager)context.getSystemService(Context.TELEPHONY_SERVICE);
        String networkOperatorName = telManager.getNetworkOperatorName();
        return networkOperatorName;
    }

    public MatchingEngineRequest createMockMatchingEngineRequest(String networkOperatorName,
                                                                 String developerName,
                                                                 String appName,
                                                                 MatchingEngine me,
                                                                 String host, int port, Location location) {
        AppClient.Match_Engine_Request request;

        // Directly create request for testing:
        LocOuterClass.Loc aLoc = LocOuterClass.Loc.newBuilder()
                .setLat(location.getLatitude())
                .setLong(location.getLongitude())
                .build();

        request = AppClient.Match_Engine_Request.newBuilder()
                .setVer(5)
                .setIdType(AppClient.IDTypes.IPADDR)
                .setId("")
                .setCarrierID(3l) // uint64 --> String? mnc, mcc?
                .setCarrierName(networkOperatorName) // Mobile Network Carrier
                .setTower(0) // cid and lac (int)
                .setGpsLocation(aLoc)
                .setAppId(5011l) // uint64 --> String again. TODO: Clarify use.
                .setProtocol(ByteString.copyFromUtf8("http")) // This one is appId context sensitive.
                .setServerPort(ByteString.copyFromUtf8("1234")) // App dependent.
                .setDevName(developerName) // From signing certificate?
                .setAppName(appName)
                .setAppVers("1") // Or versionName, which is visual name?
                .setSessionCookie(me.getSessionCookie() == null ? "" : me.getSessionCookie())
                .setVerifyLocToken(me.getTokenServerToken() == null ? "" : me.getTokenServerToken()) // Present only for VerifyLocation.
                .build();

        return new MatchingEngineRequest(request, host, port);
    }

    public MatchingEngineRequest createMockMatchingEngineRequest(String networkOperatorName, MatchingEngine me, Location location) {
        AppClient.Match_Engine_Request request;

        // Directly create request for testing:
        LocOuterClass.Loc aLoc = LocOuterClass.Loc.newBuilder()
                .setLat(location.getLatitude())
                .setLong(location.getLongitude())
                .build();

        request = AppClient.Match_Engine_Request.newBuilder()
                .setVer(5)
                .setIdType(AppClient.IDTypes.IPADDR)
                .setId("")
                .setCarrierID(3l) // uint64 --> String? mnc, mcc?
                .setCarrierName(networkOperatorName) // Mobile Network Carrier
                .setTower(0) // cid and lac (int)
                .setGpsLocation(aLoc)
                .setAppId(5011l) // uint64 --> String again. TODO: Clarify use.
                .setProtocol(ByteString.copyFromUtf8("http")) // This one is appId context sensitive.
                .setServerPort(ByteString.copyFromUtf8("1234")) // App dependent.
                .setDevName("EmptyMatchEngineApp") // From signing certificate?
                .setAppName("EmptyMatchEngineApp")
                .setAppVers("1") // Or versionName, which is visual name?
                .setSessionCookie(me.getSessionCookie() == null ? "" : me.getSessionCookie())
                .setVerifyLocToken(me.getTokenServerToken() == null ? "" : me.getTokenServerToken()) // Present only for VerifyLocation.
                .build();

        return new MatchingEngineRequest(request, me.getHost(), me.getPort());
    }

    public DynamicLocationGroupAdd createDynamicLocationGroupAdd(String networkOperatorName, MatchingEngine me, long groupLocationId, Location location) {
        // Directly create request for testing Dynamic Location Groups:
        LocOuterClass.Loc aLoc = LocOuterClass.Loc.newBuilder()
                .setLat(location.getLatitude())
                .setLong(location.getLongitude())
                .build();

        UUID uuid = UUID.randomUUID();
        AppClient.DynamicLocGroupAdd groupAdd = AppClient.DynamicLocGroupAdd.newBuilder()
                .setVer(0)
                .setIdType(IPADDR)
                .setId("")
                .setUuid(uuid.toString())
                .setCarrierID(3l)
                .setCarrierName(networkOperatorName)
                .setTower(0)
                .setGpsLocation(aLoc)
                .setLgId(groupLocationId)
                .setSessionCookie("12345")
                .setCommType(AppClient.DynamicLocGroupAdd.DlgCommType.DlgSecure)
                .setUserData("UserData").build();

        return new DynamicLocationGroupAdd(groupAdd, me.getHost(), me.getPort());
    }

    @Test
    public void registerClientTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
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

            MatchingEngineRequest request = createMockMatchingEngineRequest(getCarrierName(context), me, location);
            response = me.registerClient(request, GRPC_TIMEOUT_MS);
            assert (response != null);
        /*} catch (SSLException se) {
            Log.e(TAG, Log.getStackTraceString(se));
            assertFalse("registerClientTest: SSLException!", true);*/
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("registerClientTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertFalse("registerClientTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("registerClientTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }


        assertEquals("Sessions must be equal.", response.getSessionCookie(), me.getSessionCookie());
        // Temporary.
        Log.i(TAG, "registerClientTest response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getStatus(), AppClient.Match_Engine_Status.ME_Status.ME_SUCCESS);
    }

    @Test
    public void registerClientFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
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

            MatchingEngineRequest request = createMockMatchingEngineRequest(getCarrierName(context), me, location);
            responseFuture = me.registerClientFuture(request, GRPC_TIMEOUT_MS);
            response = responseFuture.get();
            assert(response != null);
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("registerClientFutureTest: ExecutionException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("registerClientFutureTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        assertEquals("Sessions must be equal.", response.getSessionCookie(), me.getSessionCookie());
        // Temporary.
        Log.i(TAG, "registerClientFutureTest() response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getStatus(), AppClient.Match_Engine_Status.ME_Status.ME_SUCCESS);
    }

    @Test
    public void mexDisabledTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(false);
        MexLocation mexLoc = new MexLocation(me);

        Location loc = createLocation("mexDisabledTest", -122.149349, 37.459609);
        boolean allRun = false;

        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            MatchingEngineRequest request = createMockMatchingEngineRequest(getCarrierName(context), me, location);

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
            } catch (IOException ioe) {
                Log.i(TAG, "Expected exception for verifyLocation. " + Log.getStackTraceString(ioe));
            }
            try {
                AppClient.Match_Engine_Status registerStatusResponse = me.registerClient(request, GRPC_TIMEOUT_MS);
            } catch (MissingRequestException mre) {
                // This is expected, request is missing.
                Log.i(TAG, "Expected exception for registerClient. Mex Disabled.");
            } catch (InterruptedException ioe) {
                Log.i(TAG, "Expected exception for registerClient. " + Log.getStackTraceString(ioe));
            }
            allRun = true;
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("mexDisabledTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertFalse("mexDisabledTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("mexDisabledTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        assertTrue("All requests must run with failures.", allRun);
    }

    /**
     * This test disabled networking. This test will only ever pass if the DME server accepts
     * non-cellular communications.
     */
    @Test
    public void mexNetworkingDisabledTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setNetworkSwitchingEnabled(false);
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);

        Location loc = createLocation("mexNetworkingDisabledTest", -122.149349, 37.459609);
        boolean allRun = false;

        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            MatchingEngineRequest request = createMockMatchingEngineRequest(
                    getCarrierName(context),
                    "EmptyMatchEngineRequest",
                    "EmptyMatchEngineRequest",
                    me,
                    "nosim.dme.mobiledgex.net", // This should point to a mapped hostname or actual test server
                    50051,
                    location);

            AppClient.Match_Engine_Status registerStatusResponse = me.registerClient(request, GRPC_TIMEOUT_MS);
            if (registerStatusResponse.getStatus() != AppClient.Match_Engine_Status.ME_Status.ME_SUCCESS) {
                assertFalse("mexNetworkDisabledTest: registerClient failed!", true);
            }
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("mexNetworkingDisabledTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertFalse("mexNetworkingDisabledTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("mexNetworkingDisabledTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
            me.setNetworkSwitchingEnabled(true);
        }
    }

    @Test
    public void findCloudletTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        AppClient.Match_Engine_Status registerResponse;
        FindCloudletResponse cloudletResponse = null;
        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);

        Location loc = createLocation("findCloudletTest", -122.149349, 37.459609);


        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            String carrierName = getCarrierName(context);
            registerClient(carrierName, me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(carrierName, me, location);

            cloudletResponse = me.findCloudlet(request, GRPC_TIMEOUT_MS);

        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("FindCloudlet: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertFalse("FindCloudlet: StatusRunTimeException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("FindCloudlet: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        if (cloudletResponse != null) {
            // Temporary.
            assertEquals(cloudletResponse.service_ip, cloudletResponse.service_ip);
            assertEquals("Sessions must match.", cloudletResponse.sessionCookie, "");
        } else {
            assertFalse("No findCloudlet response!", false);
        }
    }

    @Test
    public void findCloudletFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        Future<FindCloudletResponse> response;
        FindCloudletResponse result = null;
        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);

        Location loc = createLocation("findCloudletTest", -122.149349, 37.459609);

        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = mexLoc.getBlocking(context, 10000);

            String carrierName = getCarrierName(context);
            registerClient(carrierName, me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(carrierName, me, location);

            response = me.findCloudletFuture(request, 10000);
            result = response.get();
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("FindCloudletFuture: ExecutionExecution!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("FindCloudletFuture: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        assertEquals(result.service_ip, result.service_ip);
        assertEquals("SessionCookies must match.", result.sessionCookie, "");

    }

    @Test
    public void verifyLocationTest() {
        Context context = InstrumentationRegistry.getTargetContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);
        AppClient.Match_Engine_Loc_Verify response = null;

        try {
            enableMockLocation(context, true);
            Location mockLoc = createLocation("verifyLocationTest", -122.149349, 37.459609);
            setMockLocation(context, mockLoc);
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            String carrierName = getCarrierName(context);
            registerClient(carrierName, me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(carrierName, me, location);

            response = me.verifyLocation(request, GRPC_TIMEOUT_MS);
            assert (response != null);
        } catch (IOException ioe) {
            Log.e(TAG, Log.getStackTraceString(ioe));
            assertFalse("VerifyLocation: IOException!", true);
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("VerifyLocation: ExecutionExecution!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertFalse("VerifyLocation: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("VerifyLocation: InterruptedException!", true);
        } finally {
            enableMockLocation(context, false);
        }


        // Temporary.
        assertEquals(response.getVer(), 0);
        assertEquals(response.getTowerStatus(), AppClient.Match_Engine_Loc_Verify.Tower_Status.TOWER_UNKNOWN);
        assertEquals(response.getGpsLocationStatus(), AppClient.Match_Engine_Loc_Verify.GPS_Location_Status.LOC_ERROR_OTHER);
    }

    @Test
    public void verifyLocationFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);
        AppClient.Match_Engine_Loc_Verify response = null;

        Future<AppClient.Match_Engine_Loc_Verify> locFuture;

        try {
            enableMockLocation(context, true);
            Location mockLoc = createLocation("verifyLocationFutureTest", -122.149349, 37.459609);
            setMockLocation(context, mockLoc);
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            String carrierName = getCarrierName(context);
            registerClient(carrierName, me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(carrierName, me, location);

            locFuture = me.verifyLocationFuture(request, GRPC_TIMEOUT_MS);
            response = locFuture.get();
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("verifyLocationFutureTest: ExecutionException Failed!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("verifyLocationFutureTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }


        // Temporary.
        assertEquals(response.getVer(), 0);
        assertEquals(response.getTowerStatus(), AppClient.Match_Engine_Loc_Verify.Tower_Status.TOWER_UNKNOWN);
        assertEquals(response.getGpsLocationStatus(), AppClient.Match_Engine_Loc_Verify.GPS_Location_Status.LOC_ERROR_OTHER);
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


        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);

        AppClient.Match_Engine_Loc_Verify verifyLocationResult = null;
        try {
            setMockLocation(context, mockLoc); // North Pole.
            Location location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            String carrierName = getCarrierName(context);
            registerClient(carrierName, me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(carrierName, me, location);

            verifyLocationResult = me.verifyLocation(request, GRPC_TIMEOUT_MS);
            assert(verifyLocationResult != null);
        } catch (IOException ioe) {
            Log.e(TAG, Log.getStackTraceString(ioe));
            assertFalse("verifyMockedLocationTest_NorthPole: IOException!", true);
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("verifyMockedLocationTest_NorthPole: ExecutionException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("verifyMockedLocationTest_NorthPole: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        assertEquals(verifyLocationResult.getVer(), 0);
        assertEquals(verifyLocationResult.getTowerStatus(), AppClient.Match_Engine_Loc_Verify.Tower_Status.TOWER_UNKNOWN);
        assertEquals(verifyLocationResult.getGpsLocationStatus(), AppClient.Match_Engine_Loc_Verify.GPS_Location_Status.LOC_ERROR_OTHER); // Based on test data.

    }

    @Test
    public void getLocationTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);
        MexLocation mexLoc = new MexLocation(me);
        Location location;
        AppClient.Match_Engine_Status registerResponse;
        AppClient.Match_Engine_Loc response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("getLocationTest", -122.149349, 37.459609);

        String carrierName = getCarrierName(context);
        try {
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);


            registerClient(carrierName, me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(carrierName, me, location);

            response = me.getLocation(request, GRPC_TIMEOUT_MS);
            assert(response != null);
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("getLocationTest: ExecutionExecution!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertFalse("getLocationTest: StatusRuntimeException Failed!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("getLocationTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        Log.i(TAG, "getLocation() response: " + response.toString());
        assertEquals(response.getVer(), 0);

        assertEquals(response.getCarrierName(), carrierName);
        assertEquals(response.getStatus(), AppClient.Match_Engine_Loc.Loc_Status.LOC_FOUND);

        assertEquals(response.getTower(), 0);
        // FIXME: Server is currently a pure echo of client location.
        assertEquals((int) response.getNetworkLocation().getLat(), (int) loc.getLatitude());
        assertEquals((int) response.getNetworkLocation().getLong(), (int) loc.getLongitude());

        // Expected Invalid:
        assertEquals("SessionCookies must match", response.getSessionCookie(), "");

    }

    @Test
    public void getLocationFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);

        MexLocation mexLoc = new MexLocation(me);
        Location location;
        Future<AppClient.Match_Engine_Loc> responseFuture;
        AppClient.Match_Engine_Loc response = null;

        enableMockLocation(context,true);
        Location loc = createLocation("getLocationTest", -122.149349, 37.459609);

        String carrierName = getCarrierName(context);
        try {
            // Directly create request for testing:
            // Passed in Location (which is a callback interface)
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);


            registerClient(carrierName, me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(carrierName, me, location);

            responseFuture = me.getLocationFuture(request, GRPC_TIMEOUT_MS);
            response = responseFuture.get();
            assert(response != null);
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("getLocationFutureTest: ExecutionException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("getLocationFutureTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        Log.i(TAG, "getLocationFutureTest() response: " + response.toString());
        assertEquals(response.getVer(), 0);
        assertEquals(response.getCarrierName(), carrierName);
        assertEquals(response.getStatus(), AppClient.Match_Engine_Loc.Loc_Status.LOC_FOUND);

        assertEquals(response.getTower(), 0);
        // FIXME: Server is currently a pure echo of client location.
        assertEquals((int) response.getNetworkLocation().getLat(), (int) loc.getLatitude());
        assertEquals((int) response.getNetworkLocation().getLong(), (int) loc.getLongitude());

        assertEquals("SessionCookies must match", response.getSessionCookie(), "");
    }

    @Test
    public void dynamicLocationGroupAddTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);

        AppClient.Match_Engine_Status response = null;

        enableMockLocation(context,true);
        Location location = createLocation("createDynamicLocationGroupAddTest", -122.149349, 37.459609);
        MexLocation mexLoc = new MexLocation(me);

        String carrierName = getCarrierName(context);
        try {
            setMockLocation(context, location);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            registerClient(carrierName, me, location);

            // FIXME: Need groupId source.
            long groupId = 1001L;
            DynamicLocationGroupAdd dynamicLocGroupAdd = createDynamicLocationGroupAdd(carrierName, me, groupId, location);

            response = me.addUserToGroup(dynamicLocGroupAdd, GRPC_TIMEOUT_MS);
            assertTrue("DynamicLocation Group Add should return: ME_SUCCESS", response.getStatus() == AppClient.Match_Engine_Status.ME_Status.ME_SUCCESS);
            assertTrue("Group cookie result.", response.getGroupCookie().equals("")); // FIXME: This GroupCookie should have a value.

        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("dynamicLocationGroupAddTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertFalse("dynamicLocationGroupAddTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("dynamicLocationGroupAddTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        assertEquals("SessionCookies must match", response.getSessionCookie(), "");
    }

    @Test
    public void dynamicLocationGroupAddFutureTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);

        AppClient.Match_Engine_Status response = null;

        enableMockLocation(context,true);
        Location location = createLocation("createDynamicLocationGroupAddTest", -122.149349, 37.459609);
        MexLocation mexLoc = new MexLocation(me);

        String carrierName = getCarrierName(context);
        try {
            setMockLocation(context, location);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            registerClient(carrierName, me, location);

            // FIXME: Need groupId source.
            long groupId = 1001L;
            DynamicLocationGroupAdd dynamicLocGroupAdd = createDynamicLocationGroupAdd(carrierName, me, groupId, location);

            Future<AppClient.Match_Engine_Status> responseFuture = me.addUserToGroupFuture(dynamicLocGroupAdd, GRPC_TIMEOUT_MS);
            response = responseFuture.get();
            assertTrue("DynamicLocation Group Add should return: ME_SUCCESS", response.getStatus() == AppClient.Match_Engine_Status.ME_Status.ME_SUCCESS);
            assertTrue("Group cookie result.", response.getGroupCookie().equals("")); // FIXME: This GroupCookie should have a value.
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("dynamicLocationGroupAddFutureTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertFalse("dynamicLocationGroupAddFutureTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("dynamicLocationGroupAddFutureTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        // Temporary.
        assertEquals("SessionCookies must match", response.getSessionCookie(), "");
    }

    @Test
    public void getCloudletListTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);

        AppClient.Match_Engine_Status response = null;

        enableMockLocation(context,true);
        Location location = createLocation("getCloudletListTest", -122.149349, 37.459609);
        MexLocation mexLoc = new MexLocation(me);

        try {
            setMockLocation(context, location);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse("Mock'ed Location is missing!", location == null);

            registerClient(me.retrieveNetworkCarrierName(context), me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(me.retrieveNetworkCarrierName(context), me, location);

            AppClient.Match_Engine_Cloudlet_List list = me.getCloudletList(request, GRPC_TIMEOUT_MS);

            assertEquals(list.getVer(), 0);
            assertEquals(list.getStatus(), AppClient.Match_Engine_Cloudlet_List.CL_Status.CL_UNDEFINED);
            assertEquals(list.getCloudletsCount(), 0); // NOTE: This is entirely test server dependent.
            for (int i = 0; i < list.getCloudletsCount(); i++) {
                Log.v(TAG, "Cloudlet: " + list.getCloudlets(i).toString());
            }

        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("getCloudletListTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("getCloudletListTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("getCloudletListTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }

    @Test
    public void getCloudletListFutureTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);

        AppClient.Match_Engine_Status response = null;

        enableMockLocation(context,true);
        Location location = createLocation("getCloudletListFutureTest", -122.149349, 37.459609);
        MexLocation mexLoc = new MexLocation(me);

        try {
            setMockLocation(context, location);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse("Mock'ed Location is missing!", location == null);

            registerClient(me.retrieveNetworkCarrierName(context), me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(me.retrieveNetworkCarrierName(context), me, location);

            Future<AppClient.Match_Engine_Cloudlet_List> listFuture = me.getCloudletListFuture(request, GRPC_TIMEOUT_MS);
            AppClient.Match_Engine_Cloudlet_List list = listFuture.get();

            assertEquals(list.getVer(), 0);
            assertEquals(list.getStatus(), AppClient.Match_Engine_Cloudlet_List.CL_Status.CL_UNDEFINED);
            assertEquals(list.getCloudletsCount(), 0); // NOTE: This is entirely test server dependent.
            for (int i = 0; i < list.getCloudletsCount(); i++) {
                Log.v(TAG, "Cloudlet: " + list.getCloudlets(i).toString());
            }

        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("getCloudletListFutureTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("getCloudletListFutureTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("getCloudletListFutureTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }
}
