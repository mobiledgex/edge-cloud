package com.mobiledgex.matchingengine;

import android.util.Log;

import com.google.protobuf.ByteString;
import com.squareup.okhttp.Headers;
import com.squareup.okhttp.HttpUrl;
import com.squareup.okhttp.OkHttpClient;
import com.squareup.okhttp.Request;
import com.squareup.okhttp.Response;

import java.io.IOException;
import java.util.concurrent.Callable;
import java.util.concurrent.TimeUnit;

import distributed_match_engine.AppClient;
import distributed_match_engine.Match_Engine_ApiGrpc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;

public class VerifyLocation implements Callable {
    public static final String TAG = "VerifyLocationTask";

    private MatchingEngine mMatchingEngine;
    private AppClient.Match_Engine_Request mRequest; // Singleton.
    private long mTimeoutInMilliseconds = -1;

    VerifyLocation(MatchingEngine matchingEngine) {
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
        AppClient.Match_Engine_Request tokenizedRequest = AppClient.Match_Engine_Request.newBuilder()
                .setVer(mRequest.getVer())
                .setIdType(mRequest.getIdType())
                .setUuid(mRequest.getUuid())
                .setId(mRequest.getId())
                .setCarrierID(mRequest.getCarrierID())
                .setCarrierName(mRequest.getCarrierName())
                .setTower(mRequest.getTower())
                .setGpsLocation(mRequest.getGpsLocation())
                .setAppId(mRequest.getAppId())
                .setProtocol(mRequest.getProtocol())
                .setServerPort(mRequest.getServerPort())
                .setDevName(mRequest.getDevName())
                .setAppName(mRequest.getAppName())
                .setAppVers(mRequest.getAppVers())
                .setSessionCookie(mRequest.getSessionCookie())
                .setVerifyLocToken(token)
                .build();
        return tokenizedRequest;
    }

    @Override
    public AppClient.Match_Engine_Loc_Verify call()
            throws MissingRequestException, StatusRuntimeException, IOException {
        if (mRequest == null) {
            throw new MissingRequestException("Usage error: VerifyLocation does not have a request object to make location verification call!");
        }

        // Make One time use of HTTP Request to Token Server:
        String token = getToken(); // This is short lived.
        mRequest = addTokenToRequest(token);

        AppClient.Match_Engine_Loc_Verify reply;
        // FIXME: UsePlaintxt means no encryption is enabled to the MatchEngine server!
        ManagedChannel channel = null;
        try {
            channel = ManagedChannelBuilder.forAddress(mMatchingEngine.getHost(), mMatchingEngine.getPort()).usePlaintext().build();
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .verifyLocation(mRequest);
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
            Log.d(TAG, "Version of Match_Engine_Loc_Verify: " + ver);
        }

        mMatchingEngine.setTokenServerToken(token);
        mMatchingEngine.setMatchEngineLocationVerify(reply);
        return reply;
    }
}
