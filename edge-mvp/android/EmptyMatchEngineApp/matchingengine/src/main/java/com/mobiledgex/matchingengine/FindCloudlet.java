package com.mobiledgex.matchingengine;

import android.util.Log;

import java.util.concurrent.TimeUnit;
import java.util.concurrent.Callable;

import distributed_match_engine.AppClient;
import distributed_match_engine.Match_Engine_ApiGrpc;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;

public class FindCloudlet implements Callable {
    public static final String TAG = "FindCloudlet";

    private MatchingEngine mMatchingEngine;
    private AppClient.Match_Engine_Request mRequest; // Singleton.
    private long mTimeoutInMilliseconds = -1;

    public FindCloudlet(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    public boolean setRequest(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
        if (request == null) {
            throw new IllegalArgumentException("Request object must not be null.");
        } else if (!mMatchingEngine.isMexLocationAllowed()) {
            Log.d(TAG, "Mex Location is disabled.");
            mRequest = null;
            return false;
        }
        mRequest = request;

        if (timeoutInMilliseconds <= 0) {
            throw new IllegalArgumentException("VerifyLocation timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    @Override
    public FindCloudletResponse call() throws MissingRequestException, StatusRuntimeException {
        if (mRequest == null) {
            throw new MissingRequestException("Usage error: FindCloudlet does not have a request object to use MatchEngine!");
        }

        FindCloudletResponse cloudletResponse = null;

        AppClient.Match_Engine_Reply reply;
        // FIXME: UsePlaintxt means no encryption is enabled to the MatchEngine server!
        ManagedChannel channel = null;
        try {
            channel = ManagedChannelBuilder.forAddress(mMatchingEngine.getHost(), mMatchingEngine.getPort()).usePlaintext().build();
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .findCloudlet(mRequest);
        } finally {
            if (channel != null) {
                channel.shutdown();
            }
        }

        if (reply != null) {
            int ver = reply.getVer();
            Log.d(TAG, "Version of Match_Engine_Reply: " + ver);
            byte []serviceIp = (reply.getServiceIp() == null) ? null : reply.getServiceIp().toByteArray();
            if (reply.getCloudletLocation() != null) {
                GPSLocation loc = new GPSLocation(reply.getCloudletLocation().getLong(),
                        reply.getCloudletLocation().getLat(),
                        reply.getCloudletLocation().getTimestamp().getSeconds(),
                        reply.getCloudletLocation().getTimestamp().getNanos());
                cloudletResponse = new FindCloudletResponse(reply.getVer(),
                        reply.getToken(),
                        serviceIp,
                        reply.getServicePort(),
                        loc);
            } else {
                cloudletResponse = new FindCloudletResponse(reply.getVer(),
                        reply.getToken(),
                        serviceIp,
                        reply.getServicePort(),
                        null);
            }

        }
        return cloudletResponse;
    }
}
