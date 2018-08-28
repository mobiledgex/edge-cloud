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
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;

public class GetLocation implements Callable {
    public static final String TAG = "GetLocation";

    private MatchingEngine mMatchingEngine;
    private MatchingEngineRequest mRequest;
    private long mTimeoutInMilliseconds = -1;

    GetLocation(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    public boolean setRequest(MatchingEngineRequest request, long timeoutInMilliseconds) {
        if (request == null) {
            throw new IllegalArgumentException("Request object must not be null.");
        } else if (!mMatchingEngine.isMexLocationAllowed()) {
            Log.d(TAG, "Mex Location is disabled.");
            mRequest = null;
            return false;
        }
        mRequest = request;

        if (timeoutInMilliseconds <= 0) {
            throw new IllegalArgumentException("GetLocation() timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    @Override
    public AppClient.Match_Engine_Loc call()
            throws MissingRequestException, StatusRuntimeException, InterruptedException, ExecutionException {
        if (mRequest == null || mRequest.matchEngineRequest == null) {
            throw new MissingRequestException("Usage error: GetLocation does not have a request object to make location verification call!");
        }

        AppClient.Match_Engine_Loc reply;
        ManagedChannel channel = null;
        NetworkManager nm = null;
        try {
            channel = mMatchingEngine.channelPicker(mRequest.getHost(), mRequest.getPort());
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            nm = mMatchingEngine.getNetworkManager();
            nm.switchToCellularInternetNetworkBlocking();

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .getLocation(mRequest.matchEngineRequest);

            // Nothing a sdk user can do below but read the exception cause:
        } catch (MexKeyStoreException mkse) {
            throw new ExecutionException("Exception calling GetLocation: ", mkse);
        } catch (MexTrustStoreException mtse) {
            throw new ExecutionException("Exception calling GetLocation: ", mtse);
        } catch (KeyManagementException kme) {
            throw new ExecutionException("Exception calling GetLocation: ", kme);
        } catch (NoSuchAlgorithmException nsa) {
            throw new ExecutionException("Exception calling GetLocation: ", nsa);
        } catch (IOException ioe) {
            throw new ExecutionException("Exception calling GetLocation: ", ioe);
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
            Log.d(TAG, "Version of Match_Engine_Loc: " + ver);
        }

        mMatchingEngine.setMatchEngineLocation(reply);
        return reply;
    }
}
