package com.mobiledgex.matchingengine;

import android.app.UiAutomation;
import android.content.Context;
import android.os.Environment;
import android.os.Looper;
import android.support.test.InstrumentationRegistry;
import android.support.test.runner.AndroidJUnit4;

import com.google.android.gms.location.FusedLocationProviderClient;
import com.mobiledgex.matchingengine.util.MeLocation;

import org.junit.Test;
import org.junit.Before;
import org.junit.runner.RunWith;

import android.os.Build;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.nio.channels.FileChannel;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import distributed_match_engine.AppClient;
import io.grpc.StatusRuntimeException;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

import android.location.Location;
import android.util.Log;

@RunWith(AndroidJUnit4.class)
public class EngineCallTest {
    public static final String TAG = "EngineCallTest";
    public static final long GRPC_TIMEOUT_MS = 15000;

    // There's no clear way to get this programmatically outside the app signing certificate, and may
    // not be required in the future.
    public static final String developerName = "MobiledgeX";
    public static final String applicationName = "MobiledgeX SDK Demo";

    FusedLocationProviderClient fusedLocationClient;

    public static String hostOverride = "mexdemo.dme.mobiledgex.net";
    public static int portOverride = 50051;

    public boolean useHostOverride = true;

    @Before
    public void LooperEnsure() {
        // SubscriberManager needs a thread. Start one:
        if (Looper.myLooper()==null)
            Looper.prepare();
    }

    @Before
    public void grantPermissions() {

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            UiAutomation uiAutomation = InstrumentationRegistry.getInstrumentation().getUiAutomation();
            uiAutomation.grantRuntimePermission(
                    InstrumentationRegistry.getTargetContext().getPackageName(),
                    "android.permission.READ_PHONE_STATE");
            uiAutomation.grantRuntimePermission(
                    InstrumentationRegistry.getTargetContext().getPackageName(),
                    "android.permission.ACCESS_COARSE_LOCATION");
        }
    }

    private void copyFile(File source, File destination) {
        FileChannel inputChannel = null;
        FileChannel outputChannel = null;
        try {
            inputChannel = new FileInputStream(source).getChannel();
            outputChannel = new FileOutputStream(destination).getChannel();
            outputChannel.transferFrom(inputChannel, 0, inputChannel.size());
        } catch (IOException ioe) {
            assertFalse(ioe.getMessage(), true);
        } finally {
            try {
                if (inputChannel != null) {
                    inputChannel.close();
                }
            } catch (IOException ioe) {
                ioe.printStackTrace();
            }
            try {
                if (outputChannel != null) {
                    outputChannel.close();
                }
            } catch (IOException ioe) {
                ioe.printStackTrace();
            }

        }
    }

    public void installTestCertificates() {
        Context context = InstrumentationRegistry.getTargetContext();

        // Open and write some certs the test "App" needs.
        File filesDir = context.getFilesDir();
        File externalFilesDir = new File(Environment.getExternalStoragePublicDirectory(
                Environment.DIRECTORY_DOWNLOADS).getPath());

        // Read test only certs from Downloads folder, and copy them to filesDir.
        String [] fileNames = {
                "mex-ca.crt",
                "mex-client.crt",
                "mex-client.key"
        };
        for (int i = 0; i < fileNames.length; i++) {
            File srcFile = new File(externalFilesDir.getAbsolutePath() + "/" + fileNames[i]);
            File destFile = new File(filesDir.getAbsolutePath() + "/" + fileNames[i]);
            copyFile(srcFile, destFile);
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
    public void registerClient(Context context, String carrierName, MatchingEngine me) {
        AppClient.RegisterClientReply registerReply;
        AppClient.RegisterClientRequest regRequest;
        regRequest = MockUtils.createMockRegisterClientRequest(developerName, applicationName, me);
        try {
            if (useHostOverride) {
                registerReply = me.registerClient(regRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                registerReply = me.registerClient(context, regRequest, GRPC_TIMEOUT_MS);
            }
            assertEquals("Response SessionCookie should equal MatchingEngine SessionCookie",
                    registerReply.getSessionCookie(), me.getSessionCookie());
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertTrue("ExecutionException registering client", false);
        } catch (InterruptedException ioe) {
            Log.e(TAG, Log.getStackTraceString(ioe));
            assertTrue("InterruptedException registering client", false);
        }

    }

    @Test
    public void registerClientTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);

        MeLocation meLoc = new MeLocation(me);
        Location location;
        AppClient.RegisterClientReply reply = null;
        String appName = applicationName;

        enableMockLocation(context,true);
        Location loc = MockUtils.createLocation("registerClientTest", 122.3321, 47.6062);

        try {
            setMockLocation(context, loc);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            AppClient.RegisterClientRequest request = MockUtils.createMockRegisterClientRequest(developerName, appName, me);
            if (useHostOverride) {
                reply = me.registerClient(request, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                reply = me.registerClient(request, me.getHost(), me.getPort(), GRPC_TIMEOUT_MS);
            }
            assert (reply != null);
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


        assertEquals("Sessions must be equal.", reply.getSessionCookie(), me.getSessionCookie());
        // Temporary.
        Log.i(TAG, "registerClientTest reply: " + reply.toString());
        assertEquals(0, reply.getVer());
        assertEquals(AppClient.ReplyStatus.RS_SUCCESS, reply.getStatus());
    }

    @Test
    public void registerClientFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);

        MeLocation meLoc = new MeLocation(me);
        Location location;
        Future<AppClient.RegisterClientReply> registerReplyFuture;
        AppClient.RegisterClientReply reply = null;

        enableMockLocation(context,true);
        Location loc = MockUtils.createLocation("RegisterClientFutureTest", 122.3321, 47.6062);

        try {
            setMockLocation(context, loc);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            AppClient.RegisterClientRequest request = MockUtils.createMockRegisterClientRequest(developerName, applicationName, me);
            if (useHostOverride) {
                registerReplyFuture = me.registerClientFuture(request, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                registerReplyFuture = me.registerClientFuture(context, request, GRPC_TIMEOUT_MS);
            }
            reply = registerReplyFuture.get();
            assert(reply != null);
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("registerClientFutureTest: ExecutionException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("registerClientFutureTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

        assertEquals("Sessions must be equal.", reply.getSessionCookie(), me.getSessionCookie());
        // Temporary.
        Log.i(TAG, "registerClientFutureTest() response: " + reply.toString());
        assertEquals(0, reply.getVer());
        assertEquals(AppClient.ReplyStatus.RS_SUCCESS, reply.getStatus());
    }

    @Test
    public void mexDisabledTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(false);
        me.setAllowSwitchIfNoSubscriberInfo(true);
        MeLocation meLoc = new MeLocation(me);

        Location loc = MockUtils.createLocation("mexDisabledTest", 122.3321, 47.6062);
        boolean allRun = false;

        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            try {
                // Non-Mock.
                AppClient.RegisterClientRequest registerClientRequest = me.createRegisterClientRequest(
                        context, developerName, null, null, null, null);
                AppClient.RegisterClientReply registerStatusReply = me.registerClient(registerClientRequest, me.getHost(), me.getPort(), GRPC_TIMEOUT_MS);
            } catch (IllegalArgumentException iae) {
                Log.i(TAG, "Expected exception for registerClient. Mex Disabled.");
            } catch (InterruptedException ioe) {
                Log.i(TAG, "Expected exception for registerClient. " + Log.getStackTraceString(ioe));
            }

            try {
                AppClient.FindCloudletRequest findCloudletRequest;
                findCloudletRequest = me.createFindCloudletRequest(context, me.retrieveNetworkCarrierName(context), location);
                AppClient.FindCloudletReply findCloudletReply;
                if (useHostOverride) {
                    findCloudletReply = me.findCloudlet(findCloudletRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
                } else {
                    findCloudletReply = me.findCloudlet(context, findCloudletRequest, GRPC_TIMEOUT_MS);
                }
            } catch (IllegalArgumentException iae) {
                // This is expected, request is missing.
                Log.i(TAG, "Expected exception for findCloudlet. Mex Disabled.");
            }

            try {
                AppClient.GetLocationRequest locationRequest = me.createGetLocationRequest(context, MockUtils.getCarrierName(context));
                AppClient.GetLocationReply getLocationReply;
                if (useHostOverride) {
                    getLocationReply = me.getLocation(locationRequest, me.getHost(), me.getPort(), GRPC_TIMEOUT_MS);
                } else {
                    getLocationReply = me.getLocation(context, locationRequest, GRPC_TIMEOUT_MS);
                }
            } catch (IllegalArgumentException iae) {
                // This is expected, request is missing.
                Log.i(TAG, "Expected exception for getLocation. Mex Disabled.");
            }
            try {
                AppClient.VerifyLocationRequest verifyLocationRequest = me.createVerifyLocationRequest(context, MockUtils.getCarrierName(context), location);
                AppClient.VerifyLocationReply verifyLocationReply;
                if (useHostOverride) {
                    verifyLocationReply = me.verifyLocation(verifyLocationRequest, me.getHost(), me.getPort(), GRPC_TIMEOUT_MS);
                } else {
                    verifyLocationReply = me.verifyLocation(context, verifyLocationRequest, GRPC_TIMEOUT_MS);
                }
            } catch (IllegalArgumentException iae) {
                // This is expected, request is missing.
                Log.i(TAG, "Expected exception for verifyLocation. Mex Disabled.");
            } catch (IOException ioe) {
                Log.i(TAG, "Expected exception for verifyLocation. " + Log.getStackTraceString(ioe));
            }


            try {
                // Non-Mock.
                AppClient.AppInstListRequest appInstListRequest = me.createAppInstListRequest(context, MockUtils.getCarrierName(context), location);
                AppClient.AppInstListReply appInstListReply;
                if (useHostOverride) {
                    appInstListReply = me.getAppInstList(appInstListRequest, me.getHost(), me.getPort(), GRPC_TIMEOUT_MS);
                } else {
                    appInstListReply = me.getAppInstList(appInstListRequest, GRPC_TIMEOUT_MS);
                }
            } catch (IllegalArgumentException iae) {
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
    public void meNetworkingDisabledTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setNetworkSwitchingEnabled(false);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);
        MeLocation meLoc = new MeLocation(me);

        Location loc = MockUtils.createLocation("meNetworkingDisabledTest", 122.3321, 47.6062);

        AppClient.RegisterClientReply registerClientReply = null;
        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            AppClient.RegisterClientRequest registerClientRequest = MockUtils.createMockRegisterClientRequest(
                    developerName,
                    applicationName,
                    me);

            registerClientReply = me.registerClient(context, registerClientRequest, GRPC_TIMEOUT_MS);
            if (registerClientReply.getStatus() != AppClient.ReplyStatus.RS_SUCCESS) {
                assertFalse("mexNetworkDisabledTest: registerClient somehow succeeded!", true);
            }
        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("meNetworkingDisabledTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertTrue("mexNetworkDisabledTest: registerClient non-null, and somehow succeeded!",registerClientReply == null);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("meNetworkingDisabledTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
            me.setNetworkSwitchingEnabled(true);
        }
    }

    @Test
    public void findCloudletTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        AppClient.RegisterClientReply registerClientReply = null;
        AppClient.FindCloudletReply findCloudletReply = null;
        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);
        MeLocation meLoc = new MeLocation(me);

        Location loc = MockUtils.createLocation("findCloudletTest", 122.3321, 47.6062);


        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            String carrierName = MockUtils.getCarrierName(context);
            registerClient(context, me.retrieveNetworkCarrierName(context), me);

            AppClient.FindCloudletRequest findCloudletRequest = MockUtils.createMockFindCloudletRequest(
                    carrierName, me, location);
            if (useHostOverride) {
                findCloudletReply = me.findCloudlet(findCloudletRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                findCloudletReply = me.findCloudlet(context, findCloudletRequest, GRPC_TIMEOUT_MS);
            }

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

        if (findCloudletReply != null) {
            // Temporary.
            assertEquals("App's expected test cloudlet FQDN doesn't match.", "mobiledgexmobiledgexsdkdemo10.mexdemo-app-cluster.koreacentral-mexdemo.azure.mobiledgex.net", findCloudletReply.getFqdn());
        } else {
            assertFalse("No findCloudlet response!", false);
        }
    }

    @Test
    public void findCloudletFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        Future<AppClient.FindCloudletReply> response;
        AppClient.FindCloudletReply result = null;
        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);
        MeLocation meLoc = new MeLocation(me);

        Location loc = MockUtils.createLocation("findCloudletTest", 122.3321, 47.6062);

        try {
            enableMockLocation(context, true);
            setMockLocation(context, loc);
            Location location = meLoc.getBlocking(context, 10000);

            String carrierName = MockUtils.getCarrierName(context);
            registerClient(context, carrierName, me);

            AppClient.FindCloudletRequest findCloudletRequest = MockUtils.createMockFindCloudletRequest(carrierName, me, location);
            if (useHostOverride) {
                response = me.findCloudletFuture(findCloudletRequest, hostOverride, portOverride, 10000);
            } else {
                response = me.findCloudletFuture(context, findCloudletRequest, 10000);
            }
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
        assertEquals("Fully qualified domain name not expected.", "mobiledgexmobiledgexsdkdemo10.mexdemo-app-cluster.koreacentral-mexdemo.azure.mobiledgex.net", result.getFqdn());

    }

    @Test
    public void verifyLocationTest() {
        Context context = InstrumentationRegistry.getTargetContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);
        MeLocation meLoc = new MeLocation(me);
        AppClient.VerifyLocationReply verifyLocationReply = null;

        try {
            enableMockLocation(context, true);
            Location mockLoc = MockUtils.createLocation("verifyLocationTest", 122.3321, 47.6062);
            setMockLocation(context, mockLoc);
            Location location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            String carrierName = MockUtils.getCarrierName(context);
            registerClient(context, carrierName, me);

            AppClient.VerifyLocationRequest verifyLocationRequest = MockUtils.createMockVerifyLocationRequest(carrierName, me, location);

            if (useHostOverride) {
                verifyLocationReply = me.verifyLocation(verifyLocationRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                verifyLocationReply = me.verifyLocation(context, verifyLocationRequest, GRPC_TIMEOUT_MS);
            }
            assert (verifyLocationReply != null);
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
        assertEquals(0, verifyLocationReply.getVer());
        assertEquals(AppClient.VerifyLocationReply.TowerStatus.TOWER_UNKNOWN, verifyLocationReply.getTowerStatus());
        assertEquals(AppClient.VerifyLocationReply.GPSLocationStatus.LOC_ROAMING_COUNTRY_MISMATCH, verifyLocationReply.getGpsLocationStatus());
    }

    @Test
    public void verifyLocationFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);
        MeLocation meLoc = new MeLocation(me);
        AppClient.VerifyLocationReply verifyLocationReply = null;
        Future<AppClient.VerifyLocationReply> verifyLocationReplyFuture = null;
        Future<AppClient.VerifyLocationRequest> verifyLocationRequestFuture = null;

        try {
            enableMockLocation(context, true);
            Location mockLoc = MockUtils.createLocation("verifyLocationFutureTest", 122.3321, 47.6062);
            setMockLocation(context, mockLoc);
            Location location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);

            String carrierName = MockUtils.getCarrierName(context);
            registerClient(context, carrierName, me);
            AppClient.VerifyLocationRequest verifyLocationRequest = MockUtils.createMockVerifyLocationRequest(carrierName, me, location);

            if (useHostOverride) {
                verifyLocationReplyFuture = me.verifyLocationFuture(verifyLocationRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                verifyLocationReplyFuture = me.verifyLocationFuture(context, verifyLocationRequest, GRPC_TIMEOUT_MS);
            }
            verifyLocationReply = verifyLocationReplyFuture.get();
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
        assertEquals(0, verifyLocationReply.getVer());
        assertEquals(AppClient.VerifyLocationReply.TowerStatus.TOWER_UNKNOWN, verifyLocationReply.getTowerStatus());
        assertEquals(AppClient.VerifyLocationReply.GPSLocationStatus.LOC_ROAMING_COUNTRY_MISMATCH, verifyLocationReply.getGpsLocationStatus());
    }


    /**
     * Mocked Location test. Note that responsibility of verifying location is in the MatchingEngine
     * on the server side, not client side.
     */
    @Test
    public void verifyMockedLocationTest_NorthPole() {
        Context context = InstrumentationRegistry.getTargetContext();
        enableMockLocation(context,true);

        Location mockLoc = MockUtils.createLocation("verifyMockedLocationTest_NorthPole", 90d, 0d);


        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);
        MeLocation meLoc = new MeLocation(me);

        AppClient.VerifyLocationReply verifyLocationReply = null;
        try {
            setMockLocation(context, mockLoc); // North Pole.
            Location location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            String carrierName = MockUtils.getCarrierName(context);
            registerClient(context, carrierName, me);
            AppClient.VerifyLocationRequest verifyLocationRequest = MockUtils.createMockVerifyLocationRequest(carrierName, me, location);

            if (useHostOverride) {
                verifyLocationReply = me.verifyLocation(verifyLocationRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                verifyLocationReply = me.verifyLocation(context, verifyLocationRequest, GRPC_TIMEOUT_MS);
            }
            assert(verifyLocationReply != null);
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
        assertEquals(0, verifyLocationReply.getVer());
        assertEquals(AppClient.VerifyLocationReply.TowerStatus.TOWER_UNKNOWN, verifyLocationReply.getTowerStatus());
        assertEquals(AppClient.VerifyLocationReply.GPSLocationStatus.LOC_ROAMING_COUNTRY_MATCH, verifyLocationReply.getGpsLocationStatus()); // Based on test data.

    }

    @Test
    public void getLocationTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);
        MeLocation meLoc = new MeLocation(me);
        Location location;
        AppClient.GetLocationReply getLocationReply = null;

        enableMockLocation(context,true);
        Location loc = MockUtils.createLocation("getLocationTest", 122.3321, 47.6062);

        String carrierName = MockUtils.getCarrierName(context);
        try {
            setMockLocation(context, loc);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);


            registerClient(context, carrierName, me);
            AppClient.GetLocationRequest getLocationRequest = MockUtils.createMockGetLocationRequest(carrierName, me);

            if (useHostOverride) {
                getLocationReply = me.getLocation(getLocationRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                getLocationReply = me.getLocation(context, getLocationRequest, GRPC_TIMEOUT_MS);
            }
            assert(getLocationReply != null);
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
        Log.i(TAG, "getLocation() response: " + getLocationReply.toString());
        assertEquals(0, getLocationReply.getVer());

        assertEquals(carrierName, getLocationReply.getCarrierName());
        assertEquals(AppClient.GetLocationReply.LocStatus.LOC_FOUND, getLocationReply.getStatus());

        assertEquals(0, getLocationReply.getTower());
        // FIXME: Server is currently a pure echo of client location.
        assertEquals((int) loc.getLatitude(), (int) getLocationReply.getNetworkLocation().getLatitude());
        assertEquals((int) loc.getLongitude(), (int) getLocationReply.getNetworkLocation().getLongitude());
    }

    @Test
    public void getLocationFutureTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);

        MeLocation meLoc = new MeLocation(me);
        Location location;
        Future<AppClient.GetLocationReply> getLocationReplyFuture;
        AppClient.GetLocationReply getLocationReply = null;

        enableMockLocation(context,true);
        Location loc = MockUtils.createLocation("getLocationTest", 122.3321, 47.6062);

        String carrierName = MockUtils.getCarrierName(context);
        try {
            // Directly create request for testing:
            // Passed in Location (which is a callback interface)
            setMockLocation(context, loc);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);


            registerClient(context, carrierName, me);
            AppClient.GetLocationRequest getLocationRequest = MockUtils.createMockGetLocationRequest(carrierName, me);

            if (useHostOverride) {
                getLocationReplyFuture = me.getLocationFuture(getLocationRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                getLocationReplyFuture = me.getLocationFuture(context, getLocationRequest, GRPC_TIMEOUT_MS);
            }
            getLocationReply = getLocationReplyFuture.get();
            assert(getLocationReply != null);
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
        Log.i(TAG, "getLocationFutureTest() response: " + getLocationReply.toString());
        assertEquals(0, getLocationReply.getVer());
        assertEquals(carrierName, getLocationReply.getCarrierName());
        assertEquals(AppClient.GetLocationReply.LocStatus.LOC_FOUND, getLocationReply.getStatus());

        assertEquals(getLocationReply.getTower(), 0);
        // FIXME: Server is currently a pure echo of client location.
        assertEquals((int) loc.getLatitude(), (int) getLocationReply.getNetworkLocation().getLatitude());
        assertEquals((int) loc.getLongitude(), (int) getLocationReply.getNetworkLocation().getLongitude());

        assertEquals("Carriers must match", carrierName, getLocationReply.getCarrierName());
    }

    @Test
    public void dynamicLocationGroupAddTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);

        AppClient.DynamicLocGroupReply dynamicLocGroupReply = null;

        enableMockLocation(context,true);
        Location location = MockUtils.createLocation("createDynamicLocationGroupAddTest", 122.3321, 47.6062);
        MeLocation meLoc = new MeLocation(me);

        String carrierName = MockUtils.getCarrierName(context);
        try {
            setMockLocation(context, location);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            registerClient(context, carrierName, me);

            // FIXME: Need groupId source.
            long groupId = 1001L;
            AppClient.DynamicLocGroupRequest dynamicLocGroupRequest = MockUtils.createMockDynamicLocGroupRequest(me,"");

            if (useHostOverride) {
                dynamicLocGroupReply = me.addUserToGroup(dynamicLocGroupRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                dynamicLocGroupReply = me.addUserToGroup(context, dynamicLocGroupRequest, GRPC_TIMEOUT_MS);
            }
            assertTrue("DynamicLocation Group Add should return: ME_SUCCESS", dynamicLocGroupReply.getStatus() == AppClient.ReplyStatus.RS_SUCCESS);
            assertTrue("Group cookie result.", dynamicLocGroupReply.getGroupCookie().equals("")); // FIXME: This GroupCookie should have a value.

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

        assertEquals("Dynamic GroupCookie must match", "", dynamicLocGroupReply.getGroupCookie());
    }

    @Test
    public void dynamicLocationGroupAddFutureTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);

        AppClient.DynamicLocGroupReply dynamicLocGroupReply = null;

        enableMockLocation(context,true);
        Location location = MockUtils.createLocation("createDynamicLocationGroupAddTest", 122.3321, 47.6062);
        MeLocation meLoc = new MeLocation(me);

        String carrierName = MockUtils.getCarrierName(context);
        try {
            setMockLocation(context, location);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            registerClient(context, carrierName, me);

            // FIXME: Need groupId source.
            long groupId = 1001L;
            AppClient.DynamicLocGroupRequest dynamicLocGroupRequest = MockUtils.createMockDynamicLocGroupRequest(me, "");

            Future<AppClient.DynamicLocGroupReply> responseFuture;
            if (useHostOverride) {
                responseFuture = me.addUserToGroupFuture(dynamicLocGroupRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                responseFuture = me.addUserToGroupFuture(context, dynamicLocGroupRequest, GRPC_TIMEOUT_MS);
            }
            dynamicLocGroupReply = responseFuture.get();
            assertTrue("DynamicLocation Group Add should return: ME_SUCCESS", dynamicLocGroupReply.getStatus() == AppClient.ReplyStatus.RS_SUCCESS);
            assertTrue("Group cookie result.", dynamicLocGroupReply.getGroupCookie().equals("")); // FIXME: This GroupCookie should have a value.
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
        assertEquals("SessionCookies must match", "", dynamicLocGroupReply.getGroupCookie());
    }

    @Test
    public void getAppInstListTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);

        AppClient.AppInstListReply appInstListReply = null;

        enableMockLocation(context,true);
        Location location = MockUtils.createLocation("getCloudletListTest", 122.3321, 47.6062);
        MeLocation meLoc = new MeLocation(me);

        try {
            setMockLocation(context, location);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse("Mock'ed Location is missing!", location == null);

            registerClient(context, MockUtils.getCarrierName(context), me);
            AppClient.AppInstListRequest appInstListRequest;
            AppClient.AppInstListReply list;
            appInstListRequest  = MockUtils.createMockAppInstListRequest(me.retrieveNetworkCarrierName(context), me, location);
            if (useHostOverride) {
                list = me.getAppInstList(appInstListRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                list = me.getAppInstList(appInstListRequest, GRPC_TIMEOUT_MS);
            }

            assertEquals(0, list.getVer());
            assertEquals(AppClient.AppInstListReply.AIStatus.AI_SUCCESS, list.getStatus());
            assertEquals(2, list.getCloudletsCount()); // NOTE: This is entirely test server dependent.
            for (int i = 0; i < list.getCloudletsCount(); i++) {
                Log.v(TAG, "Cloudlet: " + list.getCloudlets(i).toString());
            }

        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("getAppInstListTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            Log.i(TAG, sre.getMessage());
            assertFalse("getAppInstListTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("getAppInstListTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }

    @Test
    public void getAppInstListFutureTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);

        enableMockLocation(context,true);
        Location location = MockUtils.createLocation("getAppInstListFutureTest", 122.3321, 47.6062);
        MeLocation meLoc = new MeLocation(me);

        try {
            setMockLocation(context, location);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse("Mock'ed Location is missing!", location == null);

            registerClient(context, me.retrieveNetworkCarrierName(context), me);
            AppClient.AppInstListRequest appInstListRequest = MockUtils.createMockAppInstListRequest(me.retrieveNetworkCarrierName(context), me, location);

            Future<AppClient.AppInstListReply> listFuture;
            if (useHostOverride) {
                listFuture = me.getAppInstListFuture(appInstListRequest, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            } else {
                listFuture = me.getAppInstListFuture(appInstListRequest, GRPC_TIMEOUT_MS);
            }
            AppClient.AppInstListReply list = listFuture.get();

            assertEquals(0, list.getVer());
            assertEquals(AppClient.AppInstListReply.AIStatus.AI_SUCCESS, list.getStatus());
            assertEquals(2, list.getCloudletsCount()); // NOTE: This is entirely test server dependent.
            for (int i = 0; i < list.getCloudletsCount(); i++) {
                Log.v(TAG, "Cloudlet: " + list.getCloudlets(i).toString());
            }

        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("getAppInstListFutureTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("getAppInstListFutureTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("getAppInstListFutureTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }

    @Test
    public void getQosPositionKpiTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);

        enableMockLocation(context,true);
        Location location = MockUtils.createLocation("getQosPositionKpiTest", 122.3321, 47.6062);
        MeLocation meLoc = new MeLocation(me);

        try {
            setMockLocation(context, location);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse("Mock'ed Location is missing!", location == null);

            registerClient(context, me.retrieveNetworkCarrierName(context), me);

            double totalDistanceKm = 200;
            double increment = 0.1;
            double direction = 45d;

            ArrayList<AppClient.QosPosition> kpiRequests = MockUtils.createQosPositionArray(location, direction, totalDistanceKm, increment);

            AppClient.QosPositionKpiRequest request = me.createQoSKPIRequest(kpiRequests);
            assertFalse("SessionCookie must not be empty.", request.getSessionCookie().isEmpty());


            ChannelIterator<AppClient.QosPositionKpiReply> responseIterator = me.getQosPositionKpi(request, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            // A stream of QosPositionKpiReply(s), with a non-stream block of responses.
            long total = 0;
            while (responseIterator.hasNext()) {
                AppClient.QosPositionKpiReply aR = responseIterator.next();
                for (int i = 0; i < aR.getPositionResultsCount(); i++) {
                    System.out.println(aR.getPositionResults(i));
                }
                total += aR.getPositionResultsCount();
            }
            assertEquals((long)(kpiRequests.size()), total);
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("queryQosKpiTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("queryQosKpiTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("queryQosKpiTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

    }

    @Test
    public void getQosPositionKpiFutureTest() {
        Context context = InstrumentationRegistry.getContext();

        MatchingEngine me = new MatchingEngine(context);
        me.setMatchingEngineLocationAllowed(true);
        me.setAllowSwitchIfNoSubscriberInfo(true);

        enableMockLocation(context,true);
        Location location = MockUtils.createLocation("getQosPositionKpiTest", 122.3321, 47.6062);
        MeLocation meLoc = new MeLocation(me);

        try {
            setMockLocation(context, location);
            location = meLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse("Mock'ed Location is missing!", location == null);

            registerClient(context, me.retrieveNetworkCarrierName(context), me);

            double totalDistanceKm = 200;
            double increment = 0.1;
            double direction = 45d;

            ArrayList<AppClient.QosPosition> kpiRequests = MockUtils.createQosPositionArray(location, direction, totalDistanceKm, increment);

            AppClient.QosPositionKpiRequest request = me.createQoSKPIRequest(kpiRequests);
            assertFalse("SessionCookie must not be empty.", request.getSessionCookie().isEmpty());


            Future<ChannelIterator<AppClient.QosPositionKpiReply>> replyFuture = me.getQosPositionKpiFuture(request, hostOverride, portOverride, GRPC_TIMEOUT_MS);
            // A stream of QosPositionKpiReply(s), with a non-stream block of responses.
            // Wait for value with get().
            ChannelIterator<AppClient.QosPositionKpiReply> responseIterator = replyFuture.get();
            long total = 0;
            while (responseIterator.hasNext()) {
                AppClient.QosPositionKpiReply aR = responseIterator.next();
                for (int i = 0; i < aR.getPositionResultsCount(); i++) {
                    System.out.println(aR.getPositionResults(i));
                }
                total += aR.getPositionResultsCount();
            }
            assertEquals((long)(kpiRequests.size()), total);
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("getQosPositionKpiFutureTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("getQosPositionKpiFutureTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("getQosPositionKpiFutureTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }

    }
}
