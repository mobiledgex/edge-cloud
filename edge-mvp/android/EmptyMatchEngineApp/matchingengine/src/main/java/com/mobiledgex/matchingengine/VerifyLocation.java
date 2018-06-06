package com.mobiledgex.matchingengine;

import android.util.Log;

import java.util.concurrent.Callable;
import java.util.concurrent.TimeUnit;

import distributed_match_engine.AppClient;
import distributed_match_engine.Match_Engine_ApiGrpc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

public class VerifyLocation implements Callable {
    public static final String TAG = "VerifyLocationTask";

    MatchingEngine mMatchingEngine;
    AppClient.Match_Engine_Request mRequest; // Singleton.
    long mTimeoutInMilliseconds = -1;

    VerifyLocation(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    public boolean setRequest(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
        if (request == null) {
            throw new IllegalArgumentException("Request object must not be null.");
        }
        mRequest = request;

        if (timeoutInMilliseconds < 0) {
            throw new IllegalArgumentException("VerifyLocation timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    @Override
    public Boolean call() throws MissingRequestException {
        if (mRequest == null) {
            throw new MissingRequestException("Usage error: VerifyLocation does not have a request object to make location verification call!");
        }

        AppClient.Match_Engine_Loc_Verify reply;
        // FIXME: UsePlaintxt means no encryption is enabled to the MatchEngine server!
        ManagedChannel channel = null;
        try {
            channel = ManagedChannelBuilder.forAddress(mMatchingEngine.getHost(), mMatchingEngine.getPort()).usePlaintext().build();
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .verifyLocation(mRequest);
        } catch (Exception e) {
            throw e;
        } finally {
            if (channel != null) {
                channel.shutdown();
            }
        }
        mRequest = null;
        // FIXME: Reply TBD.
        int ver = -1;
        if (reply != null) {
            ver = reply.getVer();
            Log.d(TAG, "Version of Match_Engine_Loc_Reply: " + ver);
        }
        // Stub.
        return reply.getGpsLocationStatusValue() > 0; // TODO: 1,true if < 50.
    }
}
