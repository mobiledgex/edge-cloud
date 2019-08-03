package com.mobiledgex.matchingengine;

import android.util.Log;

import java.io.IOException;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;

import distributed_match_engine.AppClient.GetLocationRequest;
import distributed_match_engine.AppClient.GetLocationReply;
import distributed_match_engine.MatchEngineApiGrpc;

import io.grpc.ManagedChannel;
import io.grpc.StatusRuntimeException;

public class GetLocation implements Callable {
    public static final String TAG = "GetLocation";

    private MatchingEngine mMatchingEngine;
    private GetLocationRequest mRequest;
    private String mHost;
    private int mPort;
    private long mTimeoutInMilliseconds = -1;

    GetLocation(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    public boolean setRequest(GetLocationRequest request, String host, int port,
                              long timeoutInMilliseconds) {
        if (request == null) {
            throw new IllegalArgumentException("Request object must not be null.");
        } else if (!mMatchingEngine.isMatchingEngineLocationAllowed()) {
            Log.e(TAG, "MatchingEngine location is disabled.");
            mRequest = null;
            return false;
        }

        if (host == null || host.equals("")) {
            return false;
        }

        mRequest = request;
        mHost = host;
        mPort = port;

        if (timeoutInMilliseconds <= 0) {
            throw new IllegalArgumentException("GetLocation() timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    @Override
    public GetLocationReply call()
            throws MissingRequestException, StatusRuntimeException, InterruptedException, ExecutionException {
        if (mRequest == null) {
            throw new MissingRequestException("Usage error: GetLocation does not have a request object to make location verification call!");
        }

        GetLocationReply reply = null;
        ManagedChannel channel = null;
        NetworkManager nm = null;
        try {
            channel = mMatchingEngine.channelPicker(mHost, mPort);
            MatchEngineApiGrpc.MatchEngineApiBlockingStub stub = MatchEngineApiGrpc.newBlockingStub(channel);

            nm = mMatchingEngine.getNetworkManager();
            nm.switchToCellularInternetNetworkBlocking();

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .getLocation(mRequest);
        } finally {
            if (channel != null) {
                channel.shutdown();
                channel.awaitTermination(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS);
            }
            if (nm != null) {
                nm.resetNetworkToDefault();
            }
        }
        mRequest = null;

        int ver;
        if (reply != null) {
            ver = reply.getVer();
            Log.d(TAG, "Version of GetLocationReply: " + ver);
        }

        mMatchingEngine.setGetLocationReply(reply);
        return reply;
    }
}
