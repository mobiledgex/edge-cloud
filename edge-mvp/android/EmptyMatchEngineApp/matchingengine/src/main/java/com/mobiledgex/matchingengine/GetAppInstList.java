package com.mobiledgex.matchingengine;

import android.util.Log;

import java.io.IOException;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;

import distributed_match_engine.AppClient.AppInstListRequest;
import distributed_match_engine.AppClient.AppInstListReply;
import distributed_match_engine.Match_Engine_ApiGrpc;
import io.grpc.ManagedChannel;
import io.grpc.StatusRuntimeException;

public class GetAppInstList implements Callable {
    public static final String TAG = "GetLocation";

    private MatchingEngine mMatchingEngine;
    private AppInstListRequest mRequest;
    private String mHost;
    private int mPort;
    private long mTimeoutInMilliseconds = -1;

    GetAppInstList(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    public boolean setRequest(AppInstListRequest request,
                              String host,
                              int port, long timeoutInMilliseconds) {
        if (request == null) {
            throw new IllegalArgumentException("Request object must not be null.");
        } else if (!mMatchingEngine.isMexLocationAllowed()) {
            Log.e(TAG, "Mex Location is disabled.");
            mRequest = null;
            return false;
        }

        if (host == null || host.equals("")) {
            return false;
        }
        mHost = host;
        mPort = port;
        mRequest = request;

        if (timeoutInMilliseconds <= 0) {
            throw new IllegalArgumentException("GetCloudletList() timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    @Override
    public AppInstListReply call()
            throws MissingRequestException, StatusRuntimeException, InterruptedException, ExecutionException {
        if (mRequest == null) {
            throw new MissingRequestException("Usage error: GetCloudletList does not have a request object!");
        }

        AppInstListReply reply;
        ManagedChannel channel = null;
        NetworkManager nm = null;
        try {
            channel = mMatchingEngine.channelPicker(mHost, mPort);
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            nm = mMatchingEngine.getNetworkManager();
            nm.switchToCellularInternetNetworkBlocking();

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .getAppInstList(mRequest);



            // Nothing a sdk user can do below but read the exception cause:
        } catch (MexKeyStoreException mkse) {
            throw new ExecutionException("Exception calling GetCloudletList: ", mkse);
        } catch (MexTrustStoreException mtse) {
            throw new ExecutionException("Exception calling GetCloudletList: ", mtse);
        } catch (KeyManagementException kme) {
            throw new ExecutionException("Exception calling GetCloudletList: ", kme);
        } catch (NoSuchAlgorithmException nsa) {
            throw new ExecutionException("Exception calling GetCloudletList: ", nsa);
        } catch (IOException ioe) {
            throw new ExecutionException("Exception calling GetCloudletList: ", ioe);
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
            Log.d(TAG, "Version of AppInstListReply: " + ver);
        }

        return reply;
    }
}
