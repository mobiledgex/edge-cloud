package com.mobiledgex.emptymatchengineapp;


// Matching Engine API:
import distributed_match_engine.AppClient;
import com.mobiledgex.matchingengine.MatchingEngine;

import android.content.Context;

import java.util.concurrent.Future;


import android.support.test.InstrumentationRegistry;
import android.support.test.runner.AndroidJUnit4;
import org.junit.Test;
import org.junit.runner.RunWith;

@RunWith(AndroidJUnit4.class)
public class TestMatchEngine {

    @Test
    public void findTheCloudlet() {
        Context appContext = InstrumentationRegistry.getTargetContext();
        // Find closest cloudlet:
        MatchingEngine task = new MatchingEngine(appContext, 40000);

        // Get Location:



        AppClient.Match_Engine_Request req = task.createRequest(loc);
        task.setRequest(req);

        Future<FindCloudletResponse> future = threadpool.submit(task);
    }
}
