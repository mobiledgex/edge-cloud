package com.mobiledgex.matchingengine;

import android.util.Log;

import com.mobiledgex.matchingengine.util.OkHttpSSLChannelHelper;

import java.io.IOException;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;

import javax.net.ssl.SSLException;
import javax.net.ssl.SSLSocketFactory;

import distributed_match_engine.AppClient;
import distributed_match_engine.Match_Engine_ApiGrpc;
import io.grpc.ManagedChannel;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.okhttp.OkHttpChannelBuilder;

public class RegisterClient implements Callable {
    public static final String TAG = "RegisterClient";
    public static final String SESSION_COOKIE_KEY = "session_cookie";
    public static final String TOKEN_SERVER_URI_KEY = "token_server_u_r_i";

    private MatchingEngine mMatchingEngine;
    private MatchingEngineRequest mRequest;
    private long mTimeoutInMilliseconds = -1;

    RegisterClient(MatchingEngine matchingEngine) {
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
            throw new IllegalArgumentException("RegisterClient() timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    private void isBoundToCellNetwork() {

    }

    @Override
    public AppClient.Match_Engine_Status call() throws MissingRequestException, StatusRuntimeException, InterruptedException, ExecutionException {
        if (mRequest == null || mRequest.matchEngineRequest == null) {
            throw new MissingRequestException("Usage error: RegisterClient() does not have a request object to make call!");
        }


        AppClient.Match_Engine_Status reply = null;
        ManagedChannel channel = null;
        NetworkManager nm = null;
        try {
            channel = mMatchingEngine.channelPicker(mRequest.getHost(), mMatchingEngine.getPort());
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            nm = mMatchingEngine.getNetworkManager();
            nm.switchToCellularInternetNetworkBlocking();

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .registerClient(mRequest.matchEngineRequest);

            // Nothing a sdk user can do below but read the exception cause:
        } catch (MexKeyStoreException mkse) {
            throw new ExecutionException("Exception calling RegisterClient: ", mkse);
        } catch (MexTrustStoreException mtse) {
            throw new ExecutionException("Exception calling RegisterClient: ", mtse);
        } catch (KeyManagementException kme) {
            throw new ExecutionException("Exception calling RegisterClient: ", kme);
        } catch (NoSuchAlgorithmException nsa) {
            throw new ExecutionException("Exception calling RegisterClient: ", nsa);
        } catch (IOException ioe) {
            throw new ExecutionException("Exception calling RegisterClient: ", ioe);
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
            Log.d(TAG, "Version of Match_Engine_Status: " + ver);
        }

        mMatchingEngine.setSessionCookie(reply.getSessionCookie());
        mMatchingEngine.setMatchEngineStatus(reply);

        mMatchingEngine.setTokenServerURI(reply.getTokenServerURI());

        return reply;
    }
}
