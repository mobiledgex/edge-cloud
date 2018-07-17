package com.mobiledgex.matchingengine;

import android.util.Log;

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

public class RegisterClient implements Callable {
    public static final String TAG = "RegisterClient";
    public static final String SESSION_COOKIE_KEY = "session_cookie";
    public static final String TOKEN_SERVER_URI_KEY = "token_server_u_r_i";

    private MatchingEngine mMatchingEngine;
    private AppClient.Match_Engine_Request mRequest;
    private long mTimeoutInMilliseconds = -1;

    RegisterClient(MatchingEngine matchingEngine) {
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
            throw new IllegalArgumentException("RegisterClient() timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
        return true;
    }

    private String createDmeUri() {
        return "http://"
                + mMatchingEngine.getHost()
                + ":"
                + mMatchingEngine.getPort();
    }

    private String getRedirectUri(String uri) {
        HttpUrl url = HttpUrl.parse(uri);
        return url.queryParameter("followURL");
    }

    /**
     *
     * @return
     * @throws MissingRequestException
     * @throws StatusRuntimeException
     */
    @Override
    public AppClient.Match_Engine_Status call() throws MissingRequestException,
            StatusRuntimeException, IOException {
        if (mRequest == null) {
            throw new MissingRequestException("Usage error: RegisterClient() does not have a request object to make call!");
        }

        // Contact DME (that's GRPC server?):
        /*
        Response response;
        OkHttpClient client = new OkHttpClient(); // From GPRC http client.
        Request httpRequest = new Request.Builder()
                //.url(createDmeUri())
                .url("")
                .build();

        // Not autoclosable:
        response = client.newCall(httpRequest).execute();
        if (!response.isRedirect()) {
            throw new IOException("Expected a Redirect Response from DME: " + response);
        }
        Headers responseHeaders = response.headers();
        String sessionCookie = responseHeaders.get(SESSION_COOKIE_KEY);
        String followURI = responseHeaders.get(TOKEN_SERVER_URI_KEY);
        String redirectTo = getRedirectUri(followURI);

        if (sessionCookie == null || redirectTo == null) {
            throw new IllegalStateException("Unexpected server behavior.");
        }
*/
        // Follow URL is verify, which the client is supposed to do, not here.


        AppClient.Match_Engine_Status reply;
        // FIXME: UsePlaintxt means no encryption is enabled to the MatchEngine server!
        ManagedChannel channel = null;
        try {
            channel = ManagedChannelBuilder.forAddress(mMatchingEngine.getHost(), mMatchingEngine.getPort()).usePlaintext().build();
            Match_Engine_ApiGrpc.Match_Engine_ApiBlockingStub stub = Match_Engine_ApiGrpc.newBlockingStub(channel);

            reply = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .registerClient(mRequest);
        } finally {
            if (channel != null) {
                channel.shutdown();
            }
        }
        mRequest = null;

        int ver;
        if (reply != null) {
            ver = reply.getVer();
            Log.d(TAG, "Version of Match_Engine_Status: " + ver);
        }

        // Future requests must use a valid session cookie.
        //mMatchingEngine.setSessionCookie(sessionCookie);
        mMatchingEngine.setSessionCookie(reply.getSessionCookie());
        mMatchingEngine.setMatchEngineStatus(reply);
        return reply;
    }
}
