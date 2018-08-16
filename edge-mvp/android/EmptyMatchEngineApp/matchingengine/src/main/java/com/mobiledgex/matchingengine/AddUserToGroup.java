package com.mobiledgex.matchingengine;

import android.util.Log;

import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;

import distributed_match_engine.AppClient;
import distributed_match_engine.Match_Engine_ApiGrpc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;

public class AddUserToGroup implements Callable {
    public static final String TAG = "AddUserToGroup";

    private MatchingEngine mMatchingEngine;
    private DynamicLocationGroupAdd mRequest; // Singleton.
    private long mTimeoutInMilliseconds = -1;

    public AddUserToGroup(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    public boolean setRequest(DynamicLocationGroupAdd request, long timeoutInMilliseconds) {
        if (request == null) {
            throw new IllegalArgumentException("Request object must not be null.");
        } else if (!mMatchingEngine.isMexLocationAllowed()) {
            Log.d(TAG, "Mex MatchEngine is disabled.");
            mRequest = null;
            return false;
        }
        mRequest = request;

        if (timeoutInMilliseconds <= 0) {
            throw new IllegalArgumentException(TAG + "timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    @Override
    public AppClient.Match_Engine_Status call()
            throws MissingRequestException, StatusRuntimeException, InterruptedException, ExecutionException {
        if (mRequest == null || mRequest.dynamicLocGroupAdd == null) {
            throw new MissingRequestException("Usage error: AddUserToGroup does not have a request object to use MatchEngine!");
        }

        AppClient.Match_Engine_Status reply;
        // FIXME: UsePlaintxt means no encryption is enabled to the MatchEngine server!
        ManagedChannel channel = null;
        try {
            channel = ManagedChannelBuilder.forAddress(mRequest.host, mRequest.port).usePlaintext().build();
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            NetworkManager nm = mMatchingEngine.getNetworkManager();
            nm.switchToCellularInternetNetworkBlocking();

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .addUserToGroup(mRequest.dynamicLocGroupAdd);

            nm.resetNetworkToDefault();
        } finally {
            if (channel != null) {
                channel.shutdown();
                channel.awaitTermination(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS);
            }
        }

        mMatchingEngine.setMatchEngineStatus(reply);
        return reply;
    }
}
