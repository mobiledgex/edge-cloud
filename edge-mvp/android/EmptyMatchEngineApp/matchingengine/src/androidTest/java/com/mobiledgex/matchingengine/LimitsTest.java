package com.mobiledgex.matchingengine;

import android.content.Context;
import android.location.Location;
import android.os.Build;
import android.support.test.InstrumentationRegistry;
import android.support.test.runner.AndroidJUnit4;
import android.util.Log;

import com.google.android.gms.location.FusedLocationProviderClient;
import com.google.protobuf.ByteString;
import com.mobiledgex.matchingengine.util.MexLocation;

import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;

import java.util.concurrent.ExecutionException;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import distributed_match_engine.AppClient;
import distributed_match_engine.LocOuterClass;
import io.grpc.StatusRuntimeException;

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

    /**
     * This is an extremely simple and basic end to end test of blocking versus Future using
     * VerifyLocation. FIXME: Manual inspection only for now.
     */
    @Test
    public void basicLatencyTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine();
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
            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            for (int i = 0; i < elapsed1.length; i++){
                start = System.currentTimeMillis();
                response1 = me.verifyLocation(request, GRPC_TIMEOUT_MS);
                elapsed1[i] = System.currentTimeMillis() - start;
            }

            for (int i = 0; i < elapsed1.length; i++) {
                Log.i("basicLatencyTest", i + " VerifyLocation elapsed time: Elapsed1: " + elapsed1[i]);
                sum1 += elapsed1[i];
            }
            Log.i("basicLatencyTest", "Average1: " + sum1/elapsed1.length);
            assert(response1 != null);

            // Future
            request = createMockMatchingEngineRequest(location);
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
                Log.i("basicLatencyTest", "Average2: " + sum2/elapsed2.length);
                assert(response2 != null);
            } catch (Exception e) {
                e.printStackTrace();
                throw e;
            }


        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("basicLatencyTest: Execution Failed!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("basicLatencyTest: Execution Failed!", true);
        }  catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("basicLatencyTest: Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }

    /**
     * Basic threading test using a thread pool to talk to dme-server.
     */
    @Test
    public void threadpoolTest() {
        Context context = InstrumentationRegistry.getTargetContext();
        MatchingEngine me = new MatchingEngine(Executors.newFixedThreadPool(100));
        me.setMexLocationAllowed(true);

        MexLocation mexLoc = new MexLocation(me);
        Location location;

        enableMockLocation(context,true);
        Location loc = createLocation("threadpoolTest", -122.149349, 37.459609);

        try {
            setMockLocation(context, loc);
            location = mexLoc.getBlocking(context, GRPC_TIMEOUT_MS);
            assertFalse(location == null);

            AppClient.Match_Engine_Request request = createMockMatchingEngineRequest(location);
            Future<AppClient.Match_Engine_Loc_Verify> responseFutures[] = new Future[10000];

            for (int i = 0; i < responseFutures.length; i++) {
                responseFutures[i] = me.verifyLocationFuture(request, GRPC_TIMEOUT_MS);
                assert(responseFutures[i] != null);
            }

            for (int i = 0; i < responseFutures.length; i++) {
                AppClient.Match_Engine_Loc_Verify locv = responseFutures[i].get();
                assert (locv != null);
                Log.i(TAG, "Locv: " + locv.getVer());
            }
        } catch (ExecutionException ee) {
            Log.i(TAG, Log.getStackTraceString(ee));
            assertFalse("threadpoolTest: Execution Failed!", true);
        } catch (StatusRuntimeException sre) {
            Log.i(TAG, Log.getStackTraceString(sre));
            assertFalse("threadpoolTest: Execution Failed!", true);
        } catch (InterruptedException ie) {
            Log.i(TAG, Log.getStackTraceString(ie));
            assertFalse("threadpoolTest: Execution Interrupted!", true);
        } finally {
            enableMockLocation(context,false);
        }
    }

}
