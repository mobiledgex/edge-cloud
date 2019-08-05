package com.mobiledgex.matchingengine;

import android.util.Log;

import java.io.IOException;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;

import distributed_match_engine.AppClient.DynamicLocGroupRequest;
import distributed_match_engine.AppClient.DynamicLocGroupReply;
import distributed_match_engine.MatchEngineApiGrpc;
import io.grpc.ManagedChannel;
import io.grpc.StatusRuntimeException;

public class AddUserToGroup implements Callable {
    public static final String TAG = "AddUserToGroup";

    private MatchingEngine mMatchingEngine;

    private DynamicLocGroupRequest mRequest;
    private String mHost;
    private int mPort;
    private long mTimeoutInMilliseconds = -1;

    public AddUserToGroup(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    public boolean setRequest(DynamicLocGroupRequest request,
                              String host, int port, long timeoutInMilliseconds) {
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
            throw new IllegalArgumentException(TAG + "timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    @Override
    public DynamicLocGroupReply call()
            throws MissingRequestException, StatusRuntimeException, InterruptedException, ExecutionException {
        if (mRequest == null) {
            throw new MissingRequestException("Usage error: AddUserToGroup does not have a request object to use MatchEngine!");
        }

        DynamicLocGroupReply reply;
        ManagedChannel channel = null;
        NetworkManager nm = null;
        try {
            channel = mMatchingEngine.channelPicker(mHost, mPort);
            MatchEngineApiGrpc.MatchEngineApiBlockingStub stub = MatchEngineApiGrpc.newBlockingStub(channel);

            nm = mMatchingEngine.getNetworkManager();
            nm.switchToCellularInternetNetworkBlocking();

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .addUserToGroup(mRequest);
        } finally {
            if (channel != null) {
                channel.shutdown();
                channel.awaitTermination(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS);
            }
            if (nm != null) {
                nm.resetNetworkToDefault();
            }
        }

        mMatchingEngine.setDynamicLocGroupReply(reply);
        return reply;
    }
}
