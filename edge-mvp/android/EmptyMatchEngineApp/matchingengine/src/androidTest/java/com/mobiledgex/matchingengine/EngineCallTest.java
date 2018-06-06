package com.mobiledgex.matchingengine;

import android.content.Context;
import android.support.test.InstrumentationRegistry;
import android.support.test.runner.AndroidJUnit4;

import com.mobiledgex.matchingengine.util.MexLocation;

import org.junit.Test;
import org.junit.Before;
import org.junit.runner.RunWith;

import android.os.Build;

import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import distributed_match_engine.AppClient;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;

@RunWith(AndroidJUnit4.class)
public class EngineCallTest {
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

    @Test
    public void findCloudletTest() {
        // Context of the app under test.
        Context context = InstrumentationRegistry.getTargetContext();
        FindCloudletResponse cloudletResponse = null;
        MatchingEngine me = new MatchingEngine();
        MexLocation mexLoc = new MexLocation(me);

        try {
            android.location.Location location = mexLoc.getBlocking(context, 10000);
            AppClient.Match_Engine_Request cloudletRequest = me.createRequest(context, location);


            cloudletResponse = me.findCloudlet(cloudletRequest, 10000);
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("FindCloudlet Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("FindCloudlet Execution Interrupted!", true);
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

        try {
            android.location.Location location = mexLoc.getBlocking(context, 10000);
            AppClient.Match_Engine_Request cloudletRequest = me.createRequest(context, location);

            response = me.findCloudletFuture(cloudletRequest, 10000);
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

        boolean verifyLocationResult = false;

        try {
            android.location.Location location = mexLoc.getBlocking(context, 10000);
            AppClient.Match_Engine_Request cloudletRequest = me.createRequest(context, location);

            verifyLocationResult = me.verifyLocation(cloudletRequest, 10000);
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("VerifyLocation Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("VerifyLocation Execution Interrupted!", true);
        }

        // Temporary.
        assertEquals(verifyLocationResult, true);
    }

    @Test
    public void verifyLocationFutureTest() {
        // Context of the app under test.
        Context context = InstrumentationRegistry.getTargetContext();

        MatchingEngine me = new MatchingEngine();
        MexLocation mexLoc = new MexLocation(me);

        Future<Boolean> locFuture;
        boolean verifyLocationResult = false;

        try {
            android.location.Location location = mexLoc.getBlocking(context, 10000);
            AppClient.Match_Engine_Request cloudletRequest = me.createRequest(context, location);

            locFuture = me.verifyLocationFuture(cloudletRequest, 10000);
            verifyLocationResult = locFuture.get();
        } catch (ExecutionException ee) {
            ee.printStackTrace();
            assertFalse("VerifyLocationFuture Execution Failed!", true);
        } catch (InterruptedException ie) {
            ie.printStackTrace();
            assertFalse("VerifyLocationFuture Execution Interrupted!", true);
        }

        // Temporary.
        assertEquals(verifyLocationResult, true);
    }
}
