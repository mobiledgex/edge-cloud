package com.mobiledgex.matchingengine;

import android.content.Context;
import android.location.Location;
import android.os.AsyncTask;
import android.os.Build;
import android.support.test.InstrumentationRegistry;
import android.support.test.runner.AndroidJUnit4;
import android.telephony.TelephonyManager;
import android.util.Log;

import com.google.android.gms.location.FusedLocationProviderClient;
import com.google.protobuf.ByteString;
import com.mobiledgex.matchingengine.util.MexLocation;

import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;

import java.io.IOException;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import distributed_match_engine.AppClient;
import distributed_match_engine.LocOuterClass;
import io.grpc.StatusRuntimeException;

import static junit.framework.Assert.assertTrue;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;

@RunWith(AndroidJUnit4.class)
public class LimitsTest {
    public static final String TAG = "LimitsTest";
    public static final long GRPC_TIMEOUT_MS = 10000;

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

    public String getCarrierName(Context context) {
        TelephonyManager telManager = (TelephonyManager)context.getSystemService(Context.TELEPHONY_SERVICE);
        String networkOperatorName = telManager.getNetworkOperatorName();
        return networkOperatorName;
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

    // Every call needs registration to be called first.
    public void registerClient(String carrierName, MatchingEngine me, Location location) {
        AppClient.Match_Engine_Status registerResponse;
        MatchingEngineRequest regRequest = createMockMatchingEngineRequest(carrierName, me, location);
        try {
            registerResponse = me.registerClient(regRequest, GRPC_TIMEOUT_MS);
            assertEquals("Response SessionCookie should equal MatchingEngine SessionCookie",
                    registerResponse.getSessionCookie(), me.getSessionCookie());
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertTrue("ExecutionException registering client", false);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertTrue("InterruptedException registering client", false);
        }

    }

    /**
     * This is an extremely simple and basic end to end test of blocking versus Future using
     * VerifyLocation. FIXME: Manual inspection only for now.
     */
    @Test
    public void basicLatencyTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context, Executors.newFixedThreadPool(20));
        me.setMexLocationAllowed(true);

        MexLocation mexLoc = new MexLocation(me);
        Location location;
        AppClient.Match_Engine_Loc_Verify response1 = null;

        enableMockLocation(context,true);
        Location loc = createLocation("basicLatencyTest", -122.149349, 37.459609);

        long start;
        long elapsed1[] = new long[20];
        long elapsed2[] = new long[20];
        try {
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            long sum1 = 0, sum2 = 0;
            String carrierName = getCarrierName(context);
            registerClient(carrierName, me, location);
            MatchingEngineRequest request = createMockMatchingEngineRequest(carrierName, me, location);
            for (int i = 0; i < elapsed1.length; i++) {
                start = System.currentTimeMillis();
                response1 = me.verifyLocation(request, GRPC_TIMEOUT_MS);
                elapsed1[i] = System.currentTimeMillis() - start;
            }

            for (int i = 0; i < elapsed1.length; i++) {
                Log.i("basicLatencyTest", i + " VerifyLocation elapsed time: Elapsed1: " + elapsed1[i]);
                sum1 += elapsed1[i];
            }
            Log.i("basicLatencyTest", "Average1: " + sum1 / elapsed1.length);
            assert (response1 != null);

            // Future
            request = createMockMatchingEngineRequest(carrierName, me, location);
            AppClient.Match_Engine_Loc_Verify response2 = null;
            try {
                for (int i = 0; i < elapsed2.length; i++) {
                    start = System.currentTimeMillis();
                    Future<AppClient.Match_Engine_Loc_Verify> locFuture = me.verifyLocationFuture(request, GRPC_TIMEOUT_MS);
                    // Do something busy()
                    response2 = locFuture.get();
                    elapsed2[i] = System.currentTimeMillis() - start;
                }
                for (int i = 0; i < elapsed2.length; i++) {
                    Log.i("basicLatencyTest", i + " VerifyLocationFuture elapsed time: Elapsed2: " + elapsed2[i]);
                    sum2 += elapsed2[i];
                }
                Log.i("basicLatencyTest", "Average2: " + sum2 / elapsed2.length);
                assert (response2 != null);
            } catch (Exception e) {
                e.printStackTrace();
                throw e;
            }

        } catch (IOException ioe) {
            Log.i(TAG, Log.getStackTraceString(ioe));
            assertFalse("basicLatencyTest: IOException!", true);
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("basicLatencyTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("basicLatencyTest: StatusRuntimeException!", true);
        }  catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("basicLatencyTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }

    @Test
    public void basicLatencyTestConcurrent() {
        final String TAG = "basicLatencyTestConcurrent";
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context);
        me.setMexLocationAllowed(true);

        MexLocation mexLoc = new MexLocation(me);
        Location location;
        AppClient.Match_Engine_Loc_Verify response1 = null;

        enableMockLocation(context,true);
        Location loc = createLocation("basicLatencyTest", -122.149349, 37.459609);

        final long start = System.currentTimeMillis();
        final long elapsed2[] = new long[20];
        final Future<AppClient.Match_Engine_Loc_Verify> responseFutures[] = new Future[elapsed2.length];
        final AppClient.Match_Engine_Loc_Verify responses[] = new AppClient.Match_Engine_Loc_Verify[elapsed2.length];
        try {
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            long sum2 = 0;
            String carrierName = getCarrierName(context);
            registerClient(carrierName, me, location);
            MatchingEngineRequest request;

            // Future
            request = createMockMatchingEngineRequest(carrierName, me, location);
            AppClient.Match_Engine_Loc_Verify response2 = null;
            try {
                // Background launch all:
                for (int i = 0; i < elapsed2.length; i++) {
                    responseFutures[i] = me.verifyLocationFuture(request, GRPC_TIMEOUT_MS);
                    elapsed2[i] = 0;
                }
                // Because everything is ultimately a Future (not callbacks) to be used with other
                // such calls, test needs match that reality and launch threads to get() independently
                AsyncTask parallelTasks[] = new AsyncTask[elapsed2.length];
                final boolean error[] = new boolean[1]; // Hacky way to not create a class for one var.
                for (int i = 0; i < elapsed2.length; i++) {
                    final int ii = i;
                    parallelTasks[i].execute(new Runnable() {
                        @Override
                        public void run() {
                            try {
                                responses[ii] = responseFutures[ii].get();
                                elapsed2[ii] = System.currentTimeMillis() - start;
                            } catch (InterruptedException ie) {
                                ie.printStackTrace();
                                assertTrue("Cancelled!", false);
                                error[0] = true;
                            } catch (ExecutionException ee) {
                                ee.printStackTrace();
                                assertTrue("Execution failed.", false);
                                error[0] = true;
                            }
                            Log.i(TAG, ii + " VerifyLocationFuture elapsed time: Elapsed2: " + elapsed2[ii]);
                        }
                    });
                }
                // Nothing fancy, just check every so often until parallel thread Async Tasks returned a result.
                boolean done = false;
                while (!done && error[0]==false) {
                    long count = 0;
                    for (int i = 0; i < responses.length; i++) {
                        if (responses[i] != null) {
                            count++;
                        }
                        if (count == responses.length) {
                            done = true;
                            break;
                        }
                    }
                    if (done) {
                        break;
                    }
                    Thread.sleep(50);
                }
                if (error[0] == false) {
                    for (int i = 0; i < elapsed2.length; i++) {
                        sum2 += elapsed2[i];
                    }
                    Log.i(TAG, "Average2: " + sum2 / elapsed2.length);
                } else {
                    assertTrue("Something went wrong with test.", false);
                }
                assert (response2 != null);
            } catch (Exception e) {
                e.printStackTrace();
                throw e;
            }
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("basicLatencyTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("basicLatencyTest: StatusRuntimeException!", true);
        }  catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("basicLatencyTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }

    /**
     * This test is set at some high values for GRPC API threads and parallel Application ASync
     * Tasks, and in debug mode, prints latency numbers. This is closer to a stress test, and not
     * realistic development practices. Can generate too many requests exceptions on system ConnectivityManager.
     */
    @Test
    public void parameterizedLatencyTest1() {
        parameterizedLatencyTestConcurrent("parameterizedLatencyTest1", 10, 1100, 180 * 1000);
    }

    /**
     * Concurrency test.
     * @param TAG Test case name.
     * @param threadPoolSize Size of GPRC API thread pool to handle App parallel tasks.
     * @param concurrency This setting can use a lot of memory due to parallel tasks.
     */
    public void parameterizedLatencyTestConcurrent(final String TAG, final int threadPoolSize, final int concurrency, long timeoutMs) {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(context, Executors.newWorkStealingPool(threadPoolSize));

        final long start = System.currentTimeMillis();
        me.setMexLocationAllowed(true);

        MexLocation mexLoc = new MexLocation(me);
        Location location;

        enableMockLocation(context,true);
        Location loc = createLocation("basicLatencyTest", -122.149349, 37.459609);

        final long elapsed[] = new long[concurrency];
        final Future<AppClient.Match_Engine_Loc_Verify> responseFutures[] = new Future[elapsed.length];
        final AppClient.Match_Engine_Loc_Verify responses[] = new AppClient.Match_Engine_Loc_Verify[elapsed.length];
        try {
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            long sum = 0;
            String carrierName = getCarrierName(context);
            registerClient(carrierName, me, location);
            MatchingEngineRequest request;

            // Future
            request = createMockMatchingEngineRequest(carrierName, me, location);

            // Background launch all:
            for (int i = 0; i < elapsed.length; i++) {
                responseFutures[i] = me.verifyLocationFuture(request, GRPC_TIMEOUT_MS);
                elapsed[i] = 0;
            }
            // Because everything is ultimately a Future (not callbacks) to be used with other
            // such calls, test needs match that reality and launch threads to get() independently
            AsyncTask parallelTasks[] = new AsyncTask[elapsed.length];
            final boolean error[] = new boolean[1]; // Hacky way to not create a class for one var.
            for (int i = 0; i < elapsed.length; i++) {
                final int ii = i;
                parallelTasks[i].execute(new Runnable() {
                    @Override
                    public void run() {
                        try {
                            responses[ii] = responseFutures[ii].get();
                            elapsed[ii] = System.currentTimeMillis() - start;
                        } catch (InterruptedException ie) {
                            ie.printStackTrace();
                            assertTrue("Cancelled!", false);
                            error[0] = true;
                        } catch (ExecutionException ee) {
                            ee.printStackTrace();
                            assertTrue("Execution failed.", false);
                            error[0] = true;
                        }
                        Log.i(TAG, ii + " VerifyLocationFuture elapsed time: Elapsed2: " + elapsed[ii]);
                    }
                });
            }
            // Nothing fancy, just check every so often until parallel thread Async Tasks returned a result.
            boolean done = false;
            long count = 0;
            while (!done && error[0]==false && (System.currentTimeMillis() - start < timeoutMs)) {
                for (int i = 0; i < responses.length; i++) {
                    if (responses[i] != null) {
                        count++;
                    }
                    if (count == responses.length) {
                        done = true;
                        break;
                    }
                }
                if (done) {
                    break;
                }
                Thread.sleep(50);
            }
            if (done && (error[0] == false)) {
                for (int i = 0; i < elapsed.length; i++) {
                    sum += elapsed[i];
                }
                Log.v(TAG, "Average: " + sum / elapsed.length);
            } else {
                assertTrue("Something went wrong with test.", false);
            }
            assertTrue("Test did not complete successfully.", done);

        } catch (ExecutionException ee) {
            Log.e(TAG, Log.getStackTraceString(ee));
            assertFalse("basicLatencyTest: ExecutionException!", true);
        } catch (StatusRuntimeException sre) {
            Log.e(TAG, Log.getStackTraceString(sre));
            assertFalse("basicLatencyTest: StatusRuntimeException!", true);
        } catch (InterruptedException ie) {
            Log.e(TAG, Log.getStackTraceString(ie));
            assertFalse("basicLatencyTest: InterruptedException!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }

}
