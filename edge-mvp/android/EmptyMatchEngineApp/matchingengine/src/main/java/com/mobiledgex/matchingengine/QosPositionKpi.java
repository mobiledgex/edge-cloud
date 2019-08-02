package com.mobiledgex.matchingengine;

import android.util.Log;

import java.io.IOException;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;

import java.util.Iterator;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;

import io.grpc.ManagedChannel;
import io.grpc.StatusRuntimeException;

import distributed_match_engine.MatchEngineApiGrpc;
import distributed_match_engine.AppClient;
import distributed_match_engine.AppClient.QosPositionKpiRequest;
import distributed_match_engine.AppClient.QosPositionKpiReply;

public class QosPositionKpi implements Callable {
    public static final String TAG = "QueryQosKpi";
    private MatchingEngine mMatchingEngine;
    private AppClient.QosPositionKpiRequest mQosPositionKpiRequest;
    private String mHost;
    private int mPort;
    private long mTimeoutInMilliseconds;

    QosPositionKpi(MatchingEngine matchingEngine) {
        mMatchingEngine = matchingEngine;
    }

    boolean setRequest(QosPositionKpiRequest qosPositionKpiRequest, String host, int port, long timeoutInMilliseconds) throws IllegalArgumentException {
        if (!mMatchingEngine.isMatchingEngineLocationAllowed()) {
            Log.e(TAG, "MobiledgeX location is disabled.");
            mQosPositionKpiRequest = null;
            return false;
        }

        if (qosPositionKpiRequest == null) {
            throw new IllegalArgumentException("Missing " + TAG + " Argument!");
        }

        if (qosPositionKpiRequest.getPositionsCount() == 0) {
            throw new IllegalArgumentException("PredictiveQos Request missing entries!");
        }

        if (host == null || host.equals("")) {
            throw new IllegalArgumentException("Host destination is required.");
        }
        if (port < 0) {
            throw new IllegalArgumentException("Port number must be positive.");
        }
        if (timeoutInMilliseconds <= 0) {
            throw new IllegalArgumentException("PredictiveQos Request timeout must be positive.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;

        mHost = host;
        mPort = port == 0 ? mMatchingEngine.getPort() : port; // Using engine default port.

        mQosPositionKpiRequest = qosPositionKpiRequest;

        return true;
    }

    @Override
    public ChannelIterator<QosPositionKpiReply> call() throws MissingRequestException, StatusRuntimeException, InterruptedException, ExecutionException {
        if (mQosPositionKpiRequest == null) {
            throw new MissingRequestException("Usage error: QueryQosKpi does not have a request object to use MatchingEngine!");
        }

        Iterator<QosPositionKpiReply> response;
        ManagedChannel channel;
        NetworkManager nm = null;
        try {
            channel = mMatchingEngine.channelPicker(mHost, mPort);
            MatchEngineApiGrpc.MatchEngineApiBlockingStub stub = MatchEngineApiGrpc.newBlockingStub(channel);

            nm = mMatchingEngine.getNetworkManager();
            nm.switchToCellularInternetNetworkBlocking();

            response = stub.withDeadlineAfter(mTimeoutInMilliseconds, TimeUnit.MILLISECONDS)
                    .getQosPositionKpi(mQosPositionKpiRequest);
        } finally {
            if (nm != null) {
                nm.resetNetworkToDefault();
            }
        }

        return new ChannelIterator<>(channel, response);
    }
}
