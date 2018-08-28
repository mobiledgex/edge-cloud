package com.mobiledgex.matchingengine;

import android.net.ConnectivityManager;
import android.net.LinkProperties;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.net.NetworkRequest;
import android.net.wifi.WifiManager;
import android.os.PersistableBundle;
import android.support.annotation.RequiresApi;
import android.telephony.CarrierConfigManager;
import android.util.Log;

import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import io.grpc.ManagedChannel;

import static android.telephony.CarrierConfigManager.KEY_CARRIER_WFC_IMS_AVAILABLE_BOOL;

public class NetworkManager {
    public static final String TAG = "NetworkManager";
    public static NetworkManager mNetworkManager;

    private ConnectivityManager mConnectivityManager;
    private boolean mWaitingForLink = false;
    private final Object mWaitForActiveNetwork = new Object();
    private long mNetworkActiveTimeoutMilliseconds = 5000;
    private final Object mSyncObject = new Object();
    private long mTimeoutInMilliseconds = 10000;

    // GRPC ManagedChannel, subject to NetworkManager.
    private ManagedChannel mGrpcChannel;

    private Network mNetwork;
    private NetworkRequest mDefaultRequest;

    private ExecutorService mThreadPool;
    private boolean mNetworkSwitchingEnabled = true;

    public boolean isNetworkSwitchingEnabled() {
        return mNetworkSwitchingEnabled;
    }

    public void setNetworkSwitchingEnabled(boolean networkSwitchingEnabled) {
        this.mNetworkSwitchingEnabled = networkSwitchingEnabled;
    }

    /**
     * Some network operations can only work on a single network type, and must wait for a suitable
     * network. This serializes those calls. The app should create separate queues to manage
     * usage otherwise if a particular parallel use pattern is needed.
     * @param callable
     * @param networkRequest
     */
    synchronized void runOnNetwork(Callable callable, NetworkRequest networkRequest)
            throws InterruptedException, ExecutionException {
        if (mNetworkSwitchingEnabled == false) {
            Log.e(TAG, "NetworkManager is disabled.");
            return;
        }
        switchToNetworkBlocking(networkRequest);
        Future<Callable> future = mThreadPool.submit(callable);
        future.get();
        resetNetworkToDefault();
    }

    public static NetworkManager getSingleton(ConnectivityManager connectivityManager) {
        if (mNetworkManager == null) {
            mNetworkManager = new NetworkManager(connectivityManager);
        }
        return mNetworkManager;
    }

    public static NetworkManager getSingleton(ConnectivityManager connectivityManager, ExecutorService executorService) {
        if (mNetworkManager == null) {
            mNetworkManager = new NetworkManager(connectivityManager, executorService);
        }
        return mNetworkManager;
    }

    private NetworkManager(ConnectivityManager connectivityManager) {
        this.mConnectivityManager = connectivityManager;
        mThreadPool = Executors.newSingleThreadExecutor();
    }

    private NetworkManager(ConnectivityManager connectivityManager, ExecutorService executorService) {
        this.mConnectivityManager = connectivityManager;
        mThreadPool = executorService;
    }

    public void setTimeout(long timeoutInMilliseconds) {
        if (timeoutInMilliseconds < 1) {
            throw new IllegalArgumentException("Network Switching Timeout should be greater than 0ms.");
        }
        mTimeoutInMilliseconds = timeoutInMilliseconds;
    }

    public long getTimeout() {
        return mTimeoutInMilliseconds;
    }

    public long getNetworkActiveTimeoutMilliseconds() {
        return mNetworkActiveTimeoutMilliseconds;
    }

    public void setNetworkActiveTimeoutMilliseconds(long networkActiveTimeoutMilliseconds) {
        if (networkActiveTimeoutMilliseconds < 0) {
            throw new IllegalArgumentException("networkActiveTimeoutMilliseconds should be greater than 0ms.");
        }
        this.mNetworkActiveTimeoutMilliseconds = networkActiveTimeoutMilliseconds;
    }

    public NetworkRequest getCellularNetworkRequest() {
        NetworkRequest networkRequest = new NetworkRequest.Builder()
                .addTransportType(NetworkCapabilities.TRANSPORT_CELLULAR)
                .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
                .build();
        return networkRequest;
    }

    public NetworkRequest getWifiNetworkRequest() {
        NetworkRequest networkRequest = new NetworkRequest.Builder()
                .addTransportType(NetworkCapabilities.TRANSPORT_WIFI)
                .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
                .build();
        return networkRequest;
    }

    public NetworkRequest getBluetoothNetworkRequest() {
        NetworkRequest networkRequest = new NetworkRequest.Builder()
                .addTransportType(NetworkCapabilities.TRANSPORT_BLUETOOTH)
                .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
                .build();
        return networkRequest;
    }

    public NetworkRequest getEthernetNetworkRequest() {
        NetworkRequest networkRequest = new NetworkRequest.Builder()
                .addTransportType(NetworkCapabilities.TRANSPORT_ETHERNET)
                .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
                .build();
        return networkRequest;
    }

    public NetworkRequest getWiFiAwareNetworkRequest() {
        NetworkRequest networkRequest = new NetworkRequest.Builder()
                .addTransportType(NetworkCapabilities.TRANSPORT_WIFI_AWARE)
                .addTransportType(NetworkCapabilities.TRANSPORT_WIFI)
                .addTransportType(NetworkCapabilities.TRANSPORT_CELLULAR)
                .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
                .build();
        return networkRequest;
    }

    public Network getNetwork() {
        return mNetwork;
    }

    // This Roaming Data value is un-reliable except under a new NetworkCapabilities Key in API 28.
    @RequiresApi(api = android.os.Build.VERSION_CODES.P)
    boolean isRoamingData() {
        boolean isroaming = mConnectivityManager.getNetworkCapabilities(mConnectivityManager.getActiveNetwork())
            .hasCapability(NetworkCapabilities.NET_CAPABILITY_NOT_ROAMING);
        return isroaming;
    }

    /**
     * Checks if the Carrier + Phone combination supports WiFiCalling. This is a supports value, it
     * does not return whether or not it is enabled.
     * @return
     */
    boolean isWiFiCallingSupported(CarrierConfigManager carrierConfigManager) {
        PersistableBundle configBundle = carrierConfigManager.getConfig();

        boolean isWifiCalling = configBundle.getBoolean(KEY_CARRIER_WFC_IMS_AVAILABLE_BOOL);
        return isWifiCalling;
    }

    public void resetNetworkToDefault() throws InterruptedException, ExecutionException {
        if (mNetworkSwitchingEnabled == false) {
            Log.e(TAG, "NetworkManager is disabled.");
            return;
        }

        if (mDefaultRequest == null) {
            mConnectivityManager.bindProcessToNetwork(null);
        } else {
            switchToNetworkBlocking(mDefaultRequest);
        }

        NetworkCapabilities networkCapabilities = mConnectivityManager.getNetworkCapabilities(mConnectivityManager.getActiveNetwork());
        logTransportCapabilities(networkCapabilities);

    }

    class NetworkSwitcherCallable implements Callable {
        NetworkRequest mNetworkRequest;
        boolean activeListenerAdded = false;
        final long start = System.currentTimeMillis();

        NetworkSwitcherCallable(NetworkRequest networkRequest) {
            mNetworkRequest = networkRequest;
        }
        @Override
        public Network call() throws InterruptedException, NetworkRequestTimeoutException {
            if (mNetworkSwitchingEnabled == false) {
                Log.e(TAG, "NetworkManager is disabled.");
                return null;
            }

            final ConnectivityManager.OnNetworkActiveListener activeListener = new ConnectivityManager.OnNetworkActiveListener() {
                @Override
                public void onNetworkActive() {
                    synchronized (mWaitForActiveNetwork) {
                        mWaitForActiveNetwork.notify();
                        long elapsed = System.currentTimeMillis() - start;
                        Log.d(TAG, "Network Switch Time Wait total: " + elapsed);
                    }
                }
            };
            try {
                synchronized (mSyncObject) {
                    mWaitingForLink = true;
                }

                ConnectivityManager.NetworkCallback networkCallback = new ConnectivityManager.NetworkCallback() {
                    @Override
                    public void onAvailable(Network network) {
                        Log.i(TAG, "requestNetwork onAvailable(), binding process to network.");
                        mConnectivityManager.bindProcessToNetwork(network);
                        activeListenerAdded = true;
                        mConnectivityManager.addDefaultNetworkActiveListener(activeListener);
                        mNetwork = network;
                    }

                    @Override
                    public void onCapabilitiesChanged(Network network, NetworkCapabilities networkCapabilities) {
                        Log.d(TAG, "requestNetwork onCapabilitiesChanged(): " + network.toString());
                        logTransportCapabilities(networkCapabilities);
                    }

                    @Override
                    public void onLinkPropertiesChanged(Network network, LinkProperties linkProperties) {
                        Log.d(TAG, "requestNetwork onLinkPropertiesChanged(): " + network.toString());
                        Log.i(TAG, " -- linkProperties: " + linkProperties.getRoutes());
                        synchronized (mSyncObject) {
                            mWaitingForLink = false;
                            mSyncObject.notify();
                        }
                    }

                    @Override
                    public void onLosing(Network network, int maxMsToLive) {
                        Log.i(TAG, "requestNetwork onLosing(): " + network.toString());
                    }

                    @Override
                    public void onLost(Network network) {
                        // unbind lost network.
                        mConnectivityManager.bindProcessToNetwork(null);
                        Log.i(TAG, "requestNetwork onLost(): " + network.toString());
                    }

                };
                mConnectivityManager.requestNetwork(mNetworkRequest, networkCallback);

                // Wait for availability.
                synchronized (mSyncObject) {
                    long timeStart = System.currentTimeMillis();
                    long elapsed;
                    while (mWaitingForLink == true &&
                            (elapsed = System.currentTimeMillis() - timeStart) < mTimeoutInMilliseconds) {
                        mSyncObject.wait(mTimeoutInMilliseconds - elapsed);
                    }
                    if (mWaitingForLink) {
                        // Timed out while waiting for available network.
                        NetworkCapabilities networkCapabilities = mConnectivityManager.getNetworkCapabilities(mConnectivityManager.getActiveNetwork());
                        mConnectivityManager.unregisterNetworkCallback(networkCallback);
                        mNetwork = null;
                        mNetworkRequest = null;

                        logTransportCapabilities(networkCapabilities);
                        throw new NetworkRequestTimeoutException("NetworkRequest timed out with no availability.");
                    }
                    elapsed = System.currentTimeMillis() - timeStart;
                    Log.i(TAG, "Elapsed time waiting for link: " + elapsed);
                    mNetworkRequest = null;
                }

                // Network is available, and link is up, but may not be active yet.
                if (!mConnectivityManager.isDefaultNetworkActive()) {
                    synchronized (mWaitForActiveNetwork) {
                        mWaitForActiveNetwork.wait(mNetworkActiveTimeoutMilliseconds);
                    }
                }
            } finally {
                if (activeListenerAdded) {
                    mConnectivityManager.removeDefaultNetworkActiveListener(activeListener);
                }
            }

            if (android.os.Build.VERSION.SDK_INT >= 28) {
                if (isRoamingData()) {
                    Log.i(TAG, "Network Roaming Data Status: " + isRoamingData());
                }
            }
            return mNetwork;
        }
    }

    private void logTransportCapabilities(NetworkCapabilities networkCapabilities) {
        Log.d(TAG, " -- networkCapabilities: TRANSPORT_CELLULAR: " + networkCapabilities.hasCapability(NetworkCapabilities.TRANSPORT_CELLULAR));
        Log.d(TAG, " -- networkCapabilities: TRANSPORT_WIFI: " + networkCapabilities.hasCapability(NetworkCapabilities.TRANSPORT_WIFI));
        Log.d(TAG, " -- networkCapabilities: TRANSPORT_BLUETOOTH: " + networkCapabilities.hasCapability(NetworkCapabilities.TRANSPORT_BLUETOOTH));
        Log.d(TAG, " -- networkCapabilities: TRANSPORT_ETHERNET: " + networkCapabilities.hasCapability(NetworkCapabilities.TRANSPORT_ETHERNET));
        Log.d(TAG, " -- networkCapabilities: TRANSPORT_VPN: " + networkCapabilities.hasCapability(NetworkCapabilities.TRANSPORT_VPN));
        Log.d(TAG, " -- networkCapabilities: TRANSPORT_WIFI_AWARE: " + networkCapabilities.hasCapability(NetworkCapabilities.TRANSPORT_WIFI_AWARE));
        Log.d(TAG, " -- networkCapabilities: TRANSPORT_LOWPAN: " + networkCapabilities.hasCapability(NetworkCapabilities.TRANSPORT_LOWPAN));

        Log.d(TAG, " -- networkCapabilities: NET_CAPABILITY_INTERNET: " + networkCapabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET));
        if (android.os.Build.VERSION.SDK_INT >= 28) {
            Log.i(TAG, " -- is Roaming Data: " + isRoamingData());
        } else {
            Log.i(TAG, " -- is Roaming Data: UNKNOWN");
        }
    }

    public boolean isCurrentNetworkInternetCellularDataCapable() {
        boolean hasDataCellCapabilities = false;

        if (mConnectivityManager != null) {
            Network network = mConnectivityManager.getBoundNetworkForProcess();
            if (network != null) {
                NetworkCapabilities networkCapabilities = mConnectivityManager.getNetworkCapabilities(network);
                if (networkCapabilities.hasCapability(NetworkCapabilities.TRANSPORT_CELLULAR) &&
                        networkCapabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)) {
                    hasDataCellCapabilities = true;
                }
            }
        }
        return hasDataCellCapabilities;
    }

    /**
     * Wrapper function to switch, if possible, to a Cellular Data Network connection. This isn't instant. Callback interface.
     */
    public void switchToCellularInternetNetwork(ConnectivityManager.NetworkCallback networkCallback) {
        boolean isCellularData = isCurrentNetworkInternetCellularDataCapable();
        if (isCellularData) {
            return; // Nothing to do, have cellular data
        }

        NetworkRequest networkRequest = new NetworkRequest.Builder()
                .addTransportType(NetworkCapabilities.TRANSPORT_CELLULAR)
                .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
                .build();

        switchToNetwork(networkRequest, networkCallback);
    }

    /**
     * Switch to a particular network type. This is a synchronous call.
     * @return
     */
    public Network switchToCellularInternetNetworkBlocking() throws InterruptedException, ExecutionException {
        boolean isCellularData = isCurrentNetworkInternetCellularDataCapable();
        if (isCellularData) {
            return null; // Nothing to do, have cellular data
        }

        NetworkRequest request = getCellularNetworkRequest();

        mNetwork = switchToNetworkBlocking(request);
        return mNetwork;
    }

    /**
     * Switch to a particular network type. Returns a Future.
     * @return
     */
    public Future<Network> switchToCellularInternetNetworkFuture() {
        boolean isCellularData = isCurrentNetworkInternetCellularDataCapable();

        NetworkRequest networkRequest = getCellularNetworkRequest();
        Future<Network> cellNetworkFuture;

        if (isCellularData) {
            return null; // Nothing to do, already have cellular data
        }

        cellNetworkFuture = mThreadPool.submit(new NetworkSwitcherCallable(networkRequest));
        return cellNetworkFuture;
    }

    /**
     * Switch to a particular network type in a blocking fashion for synchronous execution blocks.
     * @return
     */
    public Network switchToNetworkBlocking(NetworkRequest networkRequest) throws InterruptedException, ExecutionException {
        Future<Network> networkFuture;

        networkFuture = mThreadPool.submit(new NetworkSwitcherCallable(networkRequest));
        return networkFuture.get();
    }

    /**
     * Switch to a particular network type. Returns a Future.
     * @return
     */
    public Future<Network> switchToDefaultInternetNetworkFuture(NetworkRequest networkRequest) {
        Future<Network> cellNetworkFuture;

        cellNetworkFuture = mThreadPool.submit(new NetworkSwitcherCallable(networkRequest));
        return cellNetworkFuture;
    }

    /**
     * Switch to a network using Callbacks. This only does network binding. It does not wait for the network to become active.
     * @param networkRequest
     * @param networkCallback
     * @return
     */
    public void switchToNetwork(NetworkRequest networkRequest, final ConnectivityManager.NetworkCallback networkCallback) {
        mConnectivityManager.requestNetwork(networkRequest, new ConnectivityManager.NetworkCallback() {
            @Override
            public void onAvailable(Network network) {
                Log.d(TAG, "requestNetwork onAvailable(), binding process to network.");

                mNetwork = network;
                mConnectivityManager.bindProcessToNetwork(network);
                if (networkCallback == null) {
                    networkCallback.onAvailable(network);
                }
            }

            @Override
            public void onCapabilitiesChanged(Network network, NetworkCapabilities networkCapabilities) {
                logTransportCapabilities(networkCapabilities);
                if (networkCallback == null) {
                    networkCallback.onCapabilitiesChanged(network, networkCapabilities);
                }
            }

            @Override
            public void onLinkPropertiesChanged(Network network, LinkProperties linkProperties) {
                Log.d(TAG, "requestNetwork onLinkPropertiesChanged()");
                Log.d(TAG, " -- linkProperties: " + linkProperties.getRoutes());
                if (networkCallback == null) {
                    networkCallback.onLinkPropertiesChanged(network, linkProperties);
                }
            }

            @Override
            public void onLosing(Network network, int maxMsToLive) {
                Log.d(TAG, "requestNetwork onLosing()");
                if (networkCallback == null) {
                    networkCallback.onLosing(network, maxMsToLive);
                }
            }

            @Override
            public void onLost(Network network) {
                // unbind from process, lost network.
                mConnectivityManager.bindProcessToNetwork(null);
                Log.d(TAG, "requestNetwork onLost()");
                if (networkCallback == null) {
                    networkCallback.onLost(network);
                }
            }
        });
    }

}
