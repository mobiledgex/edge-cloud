package com.mobiledgex.matchingengine;

import android.util.Log;

import java.util.concurrent.Callable;
import java.util.concurrent.TimeUnit;

import distributed_match_engine.AppClient;
import distributed_match_engine.Match_Engine_ApiGrpc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

public class GetLocation implements Callable {
    public static final String TAG = "GetLocation";

    private MatchingEngine mMatchingEngine;
    private AppClient.Match_Engine_Request mRequest;
    private long mTimeoutInMilliseconds = -1;

    GetLocation(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    public boolean setRequest(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
        if (request == null) {
            throw new IllegalArgumentException("Request object must not be null.");
        }
        mRequest = request;

        if (timeoutInMilliseconds <= 0) {
            throw new IllegalArgumentException("GetLocation() timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    @Override
    public AppClient.Match_Engine_Loc call() throws MissingRequestException {
        if (mRequest == null) {
            throw new MissingRequestException("Usage error: GetLocation does not have a request object to make location verification call!");
        }

        AppClient.Match_Engine_Loc reply;
        // FIXME: UsePlaintxt means no encryption is enabled to the MatchEngine server!
        ManagedChannel channel = null;
        try {
            channel = ManagedChannelBuilder.forAddress(mMatchingEngine.getHost(), mMatchingEngine.getPort()).usePlaintext().build();
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .getLocation(mRequest);
        } catch (Exception e) {
            e.printStackTrace();
            throw e;
        } finally {
            if (channel != null) {
                channel.shutdown();
            }
        }
        mRequest = null;

        int ver;
        if (reply != null) {
            ver = reply.getVer();
            Log.d(TAG, "Version of Match_Engine_Loc: " + ver);
        }

        return reply;
    }
}
