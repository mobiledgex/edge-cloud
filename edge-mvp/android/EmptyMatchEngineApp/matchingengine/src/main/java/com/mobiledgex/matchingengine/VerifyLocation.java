package com.mobiledgex.matchingengine;

import android.util.Log;

import com.squareup.okhttp.Headers;
import com.squareup.okhttp.HttpUrl;
import com.squareup.okhttp.OkHttpClient;
import com.squareup.okhttp.Request;
import com.squareup.okhttp.Response;

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

public class VerifyLocation implements Callable {
    public static final String TAG = "VerifyLocationTask";

    private MatchingEngine mMatchingEngine;
    private MatchingEngineRequest mRequest; // Singleton.
    private long mTimeoutInMilliseconds = -1;

    VerifyLocation(MatchingEngine matchingEngine) {
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
            throw new IllegalArgumentException("VerifyLocation timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    private String getToken() throws IOException {
        String token;

        OkHttpClient httpClient = new OkHttpClient();
        httpClient.setFollowSslRedirects(false);
        httpClient.setFollowRedirects(false);

        Request request = new Request.Builder()
                .url(mMatchingEngine.getTokenServerURI())
                .build();

        Response response = httpClient.newCall(request).execute();
        if (!response.isRedirect()) {
            throw new IllegalStateException("Expected a redirect!");
        } else {
            Headers headers = response.headers();
            String locationHeaderUrl = headers.get("Location");
            if (locationHeaderUrl == null) {
                throw new IllegalStateException("Required Location Header Missing.");
            }
            HttpUrl url = HttpUrl.parse(locationHeaderUrl);
            token = url.queryParameter("dt-id");
            if (token == null) {
                throw new IllegalStateException("Required Token ID Missinng");
            }
        }

        return token;
    }

    private AppClient.Match_Engine_Request addTokenToRequest(String token) {
        AppClient.Match_Engine_Request grpcRequest = mRequest.matchEngineRequest;
        AppClient.Match_Engine_Request tokenizedRequest = AppClient.Match_Engine_Request.newBuilder()
                .setVer(grpcRequest.getVer())
                .setIdType(grpcRequest.getIdType())
                .setUuid(grpcRequest.getUuid())
                .setId(grpcRequest.getId())
                .setCarrierID(grpcRequest.getCarrierID())
                .setCarrierName(grpcRequest.getCarrierName())
                .setTower(grpcRequest.getTower())
                .setGpsLocation(grpcRequest.getGpsLocation())
                .setAppId(grpcRequest.getAppId())
                .setProtocol(grpcRequest.getProtocol())
                .setServerPort(grpcRequest.getServerPort())
                .setDevName(grpcRequest.getDevName())
                .setAppName(grpcRequest.getAppName())
                .setAppVers(grpcRequest.getAppVers())
                .setSessionCookie(grpcRequest.getSessionCookie())
                .setVerifyLocToken(token)
                .build();
        return tokenizedRequest;
    }

    @Override
    public AppClient.Match_Engine_Loc_Verify call()
            throws MissingRequestException, StatusRuntimeException,
                   IOException, InterruptedException, ExecutionException {
        if (mRequest == null || mRequest.matchEngineRequest == null) {
            throw new MissingRequestException("Usage error: VerifyLocation does not have a request object to make location verification call!");
        }
        AppClient.Match_Engine_Request grpcRequest = mRequest.matchEngineRequest;

        // Make One time use of HTTP Request to Token Server:
        NetworkManager nm = mMatchingEngine.getNetworkManager();
        nm.switchToCellularInternetNetworkBlocking();

        String token = getToken(); // This token is short lived.
        grpcRequest = addTokenToRequest(token);

        AppClient.Match_Engine_Loc_Verify reply;
        ManagedChannel channel = null;
        try {
            channel = mMatchingEngine.channelPicker(mRequest.getHost(), mRequest.getPort());
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .verifyLocation(grpcRequest);

            // Nothing a sdk user can do below but read the exception cause:
        } catch (MexKeyStoreException mkse) {
            throw new ExecutionException("Exception calling VerifyLocation: ", mkse);
        } catch (MexTrustStoreException mtse) {
            throw new ExecutionException("Exception calling VerifyLocation: ", mtse);
        } catch (KeyManagementException kme) {
            throw new ExecutionException("Exception calling VerifyLocation: ", kme);
        } catch (NoSuchAlgorithmException nsa) {
            throw new ExecutionException("Exception calling VerifyLocation: ", nsa);
        } finally {
            if (channel != null) {
                channel.shutdown();
                channel.awaitTermination(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS);
            }
            nm.resetNetworkToDefault();
        }
        mRequest = null;

        // FIXME: Reply TBD.
        int ver = -1;
        if (reply != null) {
            ver = reply.getVer();
            Log.d(TAG, "Version of Match_Engine_Loc_Verify: " + ver);
        }

        mMatchingEngine.setTokenServerToken(token);
        mMatchingEngine.setMatchEngineLocationVerify(reply);
        return reply;
    }
}
