package com.mobiledgex.matchingengine;

import android.util.Log;

import java.io.IOException;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;

import distributed_match_engine.AppClient;
import distributed_match_engine.Match_Engine_ApiGrpc;
import io.grpc.ManagedChannel;
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
            Log.e(TAG, "Mex MatchEngine is disabled.");
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
        ManagedChannel channel = null;
        NetworkManager nm = null;
        try {
            channel = mMatchingEngine.channelPicker(mRequest.getHost(), mMatchingEngine.getPort());
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            nm = mMatchingEngine.getNetworkManager();
            nm.switchToCellularInternetNetworkBlocking();

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .addUserToGroup(mRequest.dynamicLocGroupAdd);

            // Nothing a sdk user can do below but read the exception cause:
        } catch (MexKeyStoreException mkse) {
            throw new ExecutionException("Exception calling AddUserToGroup: ", mkse);
        } catch (MexTrustStoreException mtse) {
            throw new ExecutionException("Exception calling AddUserToGroup: ", mtse);
        } catch (KeyManagementException kme) {
            throw new ExecutionException("Exception calling AddUserToGroup: ", kme);
        } catch (NoSuchAlgorithmException nsa) {
            throw new ExecutionException("Exception calling AddUserToGroup: ", nsa);
        } catch (IOException ioe) {
            throw new ExecutionException("Exception calling AddUserToGroup: ", ioe);
        } finally {
            if (channel != null) {
                channel.shutdown();
                channel.awaitTermination(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS);
            }
            if (nm != null) {
                nm.resetNetworkToDefault();
            }
        }

        mMatchingEngine.setMatchEngineStatus(reply);
        return reply;
    }
}
