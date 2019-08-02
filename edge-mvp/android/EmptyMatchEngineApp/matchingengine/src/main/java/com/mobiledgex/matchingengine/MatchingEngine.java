package com.mobiledgex.matchingengine;

import android.content.Context;
import android.content.pm.ApplicationInfo;
import android.content.pm.PackageManager;
import android.net.ConnectivityManager;
import android.net.wifi.WifiManager;
import android.provider.Settings;
import android.support.annotation.RequiresApi;
import android.telephony.CarrierConfigManager;
import android.telephony.SubscriptionInfo;
import android.telephony.SubscriptionManager;
import android.telephony.TelephonyManager;

import java.io.IOException;

import java.util.List;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import distributed_match_engine.AppClient;
import distributed_match_engine.AppClient.RegisterClientRequest;
import distributed_match_engine.AppClient.RegisterClientReply;
import distributed_match_engine.AppClient.VerifyLocationRequest;
import distributed_match_engine.AppClient.VerifyLocationReply;
import distributed_match_engine.AppClient.FindCloudletRequest;
import distributed_match_engine.AppClient.FindCloudletReply;
import distributed_match_engine.AppClient.GetLocationRequest;
import distributed_match_engine.AppClient.GetLocationReply;
import distributed_match_engine.AppClient.AppInstListRequest;
import distributed_match_engine.AppClient.AppInstListReply;
import distributed_match_engine.AppClient.QosPositionKpiRequest;
import distributed_match_engine.AppClient.QosPositionKpiReply;
import distributed_match_engine.AppClient.QosPosition;


import distributed_match_engine.AppClient.DynamicLocGroupRequest;
import distributed_match_engine.AppClient.DynamicLocGroupReply;

import distributed_match_engine.LocOuterClass;
import distributed_match_engine.LocOuterClass.Loc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import io.grpc.okhttp.OkHttpChannelBuilder;

import android.content.pm.PackageInfo;
import android.util.Log;

import javax.net.ssl.KeyManagerFactory;
import javax.net.ssl.SSLContext;
import javax.net.ssl.SSLSocketFactory;
import javax.net.ssl.TrustManagerFactory;

import static android.content.Context.TELEPHONY_SUBSCRIPTION_SERVICE;

public class MatchingEngine {
    public static final String TAG = "MatchingEngine";
    private final String mInitialDMEContactHost = "tdg.dme.mobiledgex.net";
    private String host = mInitialDMEContactHost;
    private NetworkManager mNetworkManager;
    private int port = 50051;

    // A threadpool for all the MatchEngine API callable interfaces:
    final ExecutorService threadpool;

    // State info for engine
    private String mSessionCookie;
    private String mTokenServerURI;
    private String mTokenServerToken;

    private RegisterClientReply mRegisterClientReply;
    private FindCloudletReply mFindCloudletReply;
    private VerifyLocationReply mVerifyLocationReply;
    private GetLocationReply mGetLocationReply;
    private DynamicLocGroupReply mDynamicLocGroupReply;

    private LocOuterClass.Loc mMatchEngineLocation;

    private boolean isSSLEnabled = true;

    private Context mContext;

    public MatchingEngine(Context context) {
        threadpool = Executors.newSingleThreadExecutor();
        ConnectivityManager connectivityManager = context.getSystemService(ConnectivityManager.class);
        mNetworkManager = NetworkManager.getInstance(connectivityManager, getSubscriptionManager(context));
        mContext = context;
    }
    public MatchingEngine(Context context, ExecutorService executorService) {
        threadpool = executorService;
        ConnectivityManager connectivityManager = context.getSystemService(ConnectivityManager.class);
        mNetworkManager = NetworkManager.getInstance(connectivityManager, getSubscriptionManager(context), threadpool);
        mContext = context;
    }

    // Application state Bundle Key.
    public static final String MATCHING_ENGINE_LOCATION_PERMISSION = "MATCHING_ENGINE_LOCATION_PERMISSION";
    private static boolean mMatchingEngineLocationAllowed = false;

    public static boolean isMatchingEngineLocationAllowed() {
        return mMatchingEngineLocationAllowed;
    }

    public static void setMatchingEngineLocationAllowed(boolean allowMatchingEngineLocation) {
        mMatchingEngineLocationAllowed = allowMatchingEngineLocation;
    }

    private SubscriptionManager getSubscriptionManager(Context context) {
        return (SubscriptionManager)context.getSystemService(TELEPHONY_SUBSCRIPTION_SERVICE);
    }

    public boolean isNetworkSwitchingEnabled() {
        return getNetworkManager().isNetworkSwitchingEnabled();
    }

    public void setNetworkSwitchingEnabled(boolean networkSwitchingEnabled) {
        getNetworkManager().setNetworkSwitchingEnabled(networkSwitchingEnabled);
    }

    public boolean isAllowSwitchIfNoSubscriberInfo() {
        return getNetworkManager().isAllowSwitchIfNoSubscriberInfo();
    }

    public void setAllowSwitchIfNoSubscriberInfo(boolean allowSwitchIfNoSubscriberInfo) {
        getNetworkManager().setAllowSwitchIfNoSubscriberInfo(allowSwitchIfNoSubscriberInfo);
    }

    /**
     * Utility function to get the active subscription network provider list for this device as
     * known to the MatchingEngine. If it is empty, the application should use the public cloud
     * instead, as the Distributed Matching Engine may be unavailable (firewalled) from the current
     * network. Calling MatchingEngine APIs in that state will result in a
     * NetworkRequestNoSubscriptionInfoException.
     *
     * @return
     */
    public List<SubscriptionInfo> getActiveSubscriptionInfoList() {
        List<SubscriptionInfo> subs = getNetworkManager().getActiveSubscriptionInfoList(true);
        return subs;
    }

    /**
     * Check if Roaming Data is enabled on the System.
     * @return
     */
    public boolean isRoamingDataEanbled(Context context) {
        boolean enabled;
        try {
            enabled = android.provider.Settings.Global.getInt(context.getContentResolver(), Settings.Global.DATA_ROAMING) == 1;
        } catch (Settings.SettingNotFoundException snfe) {
            Log.i(TAG, "android.provider.Settings.Global.DATA_ROAMING key is not present!");
            return false; // Unavailable setting.
        }

        return enabled;
    }

    public Future submit(Callable task) {
        return threadpool.submit(task);
    }

    void setSessionCookie(String sessionCookie) {
        this.mSessionCookie = sessionCookie;
    }
    String getSessionCookie() {
        return this.mSessionCookie;
    }

    void setMatchEngineStatus(AppClient.RegisterClientReply status) {
        mRegisterClientReply = status;
    }

    void setGetLocationReply(GetLocationReply locationReply) {
        mGetLocationReply = locationReply;
        mMatchEngineLocation = locationReply.getNetworkLocation();
    }

    void setVerifyLocationReply(AppClient.VerifyLocationReply locationVerify) {
        mVerifyLocationReply = locationVerify;
    }

    void setFindCloudletResponse(AppClient.FindCloudletReply reply) {
        mFindCloudletReply = reply;
    }

    void setDynamicLocGroupReply(DynamicLocGroupReply reply) {
        mDynamicLocGroupReply = reply;
    }

    /**
     * Utility method retrieves current network CarrierName from system service.
     * @param context
     * @return
     */
    public String retrieveNetworkCarrierName(Context context) {
        TelephonyManager telManager = (TelephonyManager)context.getSystemService(Context.TELEPHONY_SERVICE);
        String networkOperatorName = telManager.getNetworkOperatorName();
        if (networkOperatorName == null) {
            Log.w(TAG, "Network Carrier name is not found on device.");
        }
        return networkOperatorName;
    }

    public String generateDmeHostAddress(String networkOperatorName) {
        String host;

        if (networkOperatorName == null || networkOperatorName.isEmpty()) {
            host = mInitialDMEContactHost;
            return host;
        }

        host = networkOperatorName + ".dme.mobiledgex.net";
        return host;
    }

    NetworkManager getNetworkManager() {
        return mNetworkManager;
    }

    void setNetworkManager(NetworkManager networkManager) {
        mNetworkManager = networkManager;
    }

    public String getAppName(Context context) {
        String appName = "";
        String packageLabel = "";

        ApplicationInfo appInfo = context.getApplicationInfo();
        if (context.getPackageManager() != null) {
            CharSequence seq = appInfo.loadLabel(context.getPackageManager());
            if (seq != null) {
                packageLabel = seq.toString();
            }
        }
        if (packageLabel != null && !packageLabel.isEmpty()) {
            appName = packageLabel;
        }
        return appName;
    }

    public RegisterClientRequest createRegisterClientRequest(Context context, String developerName,
                                                             String applicationName, String appVersion,
                                                             String carrierName, String authToken)
    {
        if (!mMatchingEngineLocationAllowed) {
            Log.d(TAG, "Create Request disabled. Matching engine is not configured to allow use.");
            return null;
        }
        if (context == null) {
            throw new IllegalArgumentException("MatchingEngine requires a working application context.");
        }
        // App
        ApplicationInfo appInfo = context.getApplicationInfo();
        String packageLabel = "";
        if (context.getPackageManager() != null) {
            CharSequence seq = appInfo.loadLabel(context.getPackageManager());
            if (seq != null) {
                packageLabel = seq.toString();
            }
        }
        PackageInfo pInfo;
        String versionName = "";
        String appName;
        if (applicationName == null || applicationName.equals("")) {
            appName = packageLabel;
        } else {
            appName = applicationName;
        }

        try {
            pInfo = context.getPackageManager().getPackageInfo(context.getPackageName(), 0);
            versionName = (appVersion == null || appVersion.isEmpty()) ? pInfo.versionName : appVersion;
        } catch (PackageManager.NameNotFoundException nfe) {
            nfe.printStackTrace();
            // Hard stop, or continue?
        }
        if(developerName == null || developerName.equals("")) {
            developerName = packageLabel; // From signing certificate?
        }
        return AppClient.RegisterClientRequest.newBuilder()
                .setDevName((developerName == null) ? "" : developerName)
                .setAppName(appName)
                .setAppVers(versionName)
                .setCarrierName((carrierName == null || carrierName.equals(""))
                        ? retrieveNetworkCarrierName(context) : carrierName)
                .setAuthToken((authToken == null) ? "" : authToken)
                .build();
    }

    public VerifyLocationRequest createVerifyLocationRequest(Context context, String carrierName,
                                                             android.location.Location location) {
        if (context == null) {
            throw new IllegalArgumentException("MatchingEngine requires a working application context.");
        }

        if (!mMatchingEngineLocationAllowed) {
            Log.d(TAG, "Create Request disabled. Matching engine is not configured to allow use.");
            return null;
        }

        if (location == null) {
            throw new IllegalArgumentException("Location parameter is required.");
        }

        String retrievedNetworkOperatorName = retrieveNetworkCarrierName(context);
        if(carrierName == null || carrierName.equals("")) {
            carrierName = retrievedNetworkOperatorName;
        }
        Loc aLoc = androidLocToMeLoc(location);

        return AppClient.VerifyLocationRequest.newBuilder()
                .setSessionCookie(mSessionCookie)
                .setCarrierName(
                        (carrierName == null || carrierName.equals(""))
                            ? retrieveNetworkCarrierName(context) : carrierName)
                .setGpsLocation(aLoc) // Latest token is unknown until retrieved.
                .build();
    }

    public FindCloudletRequest createFindCloudletRequest(Context context, String carrierName,
                                                         android.location.Location location) {
        if (!mMatchingEngineLocationAllowed) {
            Log.d(TAG, "Create Request disabled. Matching engine is not configured to allow use.");
            return null;
        }
        if (context == null) {
            throw new IllegalArgumentException("MatchingEngine requires a working application context.");
        }

        Loc aLoc = androidLocToMeLoc(location);

        return FindCloudletRequest.newBuilder()
                .setSessionCookie(mSessionCookie)
                .setCarrierName(
                        (carrierName == null || carrierName.equals(""))
                            ? retrieveNetworkCarrierName(context) : carrierName
                )
                .setGpsLocation(aLoc)
                .build();
    }

    public GetLocationRequest createGetLocationRequest(Context context, String carrierName) {
        if (!mMatchingEngineLocationAllowed) {
            Log.d(TAG, "Create Request disabled. Matching engine is not configured to allow use.");
            return null;
        }
        if (context == null) {
            throw new IllegalArgumentException("MatchingEngine requires a working application context.");
        }

        return GetLocationRequest.newBuilder()
                .setSessionCookie(mSessionCookie)
                .setCarrierName(
                        (carrierName == null || carrierName.equals(""))
                            ? retrieveNetworkCarrierName(context) : carrierName

                )
                .build();
    }

    public AppInstListRequest createAppInstListRequest(Context context, String carrierName,
                                                       android.location.Location location) {
        if (!mMatchingEngineLocationAllowed) {
            Log.d(TAG, "Create Request disabled. Matching engine is not configured to allow use.");
            return null;
        }
        if (context == null) {
            throw new IllegalArgumentException("MatchingEngine requires a working application context.");
        }


        if (location == null) {
            throw new IllegalArgumentException("Location parameter is required.");
        }

        String retrievedNetworkOperatorName = retrieveNetworkCarrierName(context);
        if(carrierName == null || carrierName.equals("")) {
            carrierName = retrievedNetworkOperatorName;
        }
        Loc aLoc = androidLocToMeLoc(location);

        return AppClient.AppInstListRequest.newBuilder()
                .setSessionCookie(mSessionCookie)
                .setCarrierName((carrierName == null || carrierName.equals(""))
                        ? retrieveNetworkCarrierName(context) : carrierName)
                .setGpsLocation(aLoc)
                .build();
    }

    public DynamicLocGroupRequest createDynamicLocGroupRequest(Context context,
                                                               DynamicLocGroupRequest.DlgCommType commType,
                                                               String userData) {
        if (!mMatchingEngineLocationAllowed) {
            Log.d(TAG, "Create Request disabled. Matching engine is not configured to allow use.");
            return null;
        }
        if (context == null) {
            throw new IllegalArgumentException("MatchingEngine requires a working application context.");
        }

        if (commType == null || commType == DynamicLocGroupRequest.DlgCommType.DLG_UNDEFINED) {
            commType = DynamicLocGroupRequest.DlgCommType.DLG_SECURE;
        }

        return DynamicLocGroupRequest.newBuilder()
                .setSessionCookie(mSessionCookie)
                .setLgId(1001L) // FIXME: NOT IMPLEMENTED
                .setCommType(commType)
                .setUserData(userData == null ? "" : userData)
                .build();
    }

    public QosPositionKpiRequest createQoSKPIRequest(List<QosPosition> requests) {
        return QosPositionKpiRequest.newBuilder()
                .setSessionCookie(mSessionCookie)
                .addAllPositions(requests)
                .build();
    }

    private Loc androidLocToMeLoc(android.location.Location loc) {
        return Loc.newBuilder()
                .setLatitude((loc == null) ? 0.0d : loc. getLatitude())
                .setLongitude((loc == null) ? 0.0d : loc.getLongitude())
                .setHorizontalAccuracy((loc == null) ? 0.0d :loc.getAccuracy())
                //.setVerticalAccuracy(loc.getVerticalAccuracyMeters()) // API Level 26 required.
                .setVerticalAccuracy(0d)
                .setAltitude((loc == null) ? 0.0d : loc.getAltitude())
                .setCourse((loc == null) ? 0.0d : loc.getBearing())
                .setSpeed((loc == null) ? 0.0d : loc.getSpeed())
                .build();
    }

    /**
     * Registers Client using blocking API call.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws StatusRuntimeException
     * @throws InterruptedException
     * @throws ExecutionException
     */
    public RegisterClientReply registerClient(Context context,
                                              RegisterClientRequest request,
                                              long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, ExecutionException {
        String carrierName = retrieveNetworkCarrierName(context);
        return registerClient(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }
    /**
     * Registers Client using blocking API call. Allows specifying a DME host and port.
     * @param request
     * @param host Distributed Matching Engine hostname
     * @param port Distributed Matching Engine port
     * @param timeoutInMilliseconds
     * @return
     * @throws StatusRuntimeException
     */
    public RegisterClientReply registerClient(RegisterClientRequest request,
                                              String host, int port,
                                              long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, ExecutionException {
        RegisterClient registerClient = new RegisterClient(this); // Instanced, so just add host, port as field.
        registerClient.setRequest(request, host, port, timeoutInMilliseconds);
        return registerClient.call();
    }

    public Future<RegisterClientReply> registerClientFuture(Context context,
                                                            RegisterClientRequest request,
                                                            long timeoutInMilliseconds) {
        String carrierName = retrieveNetworkCarrierName(context);
        return registerClientFuture(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }
    /**
     * Registers device on the MatchingEngine server. Returns a Future.
     * @param request
     * @param host Distributed Matching Engine hostname
     * @param port Distributed Matching Engine port
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<RegisterClientReply> registerClientFuture(RegisterClientRequest request,
                                                            String host, int port,
                                                            long timeoutInMilliseconds) {
        RegisterClient registerClient = new RegisterClient(this);
        registerClient.setRequest(request, host, port, timeoutInMilliseconds);
        return submit(registerClient);
    }

    /**
     * findCloudlet finds the closest cloudlet instance as per request.
     * @param context
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws StatusRuntimeException
     * @throws InterruptedException
     * @throws ExecutionException
     */
    public FindCloudletReply findCloudlet(Context context,
                                          FindCloudletRequest request,
                                          long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, ExecutionException {
        String carrierName = retrieveNetworkCarrierName(context);
        return findCloudlet(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);

    }
    /**
     * findCloudlet finds the closest cloudlet instance as per request.
     * @param request
     * @param host Distributed Matching Engine hostname
     * @param port Distributed Matching Engine port
     * @param timeoutInMilliseconds
     * @return cloudlet URI.
     * @throws StatusRuntimeException
     */
    public FindCloudletReply findCloudlet(FindCloudletRequest request,
                                          String host, int port,
                                          long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, ExecutionException {
        FindCloudlet findCloudlet = new FindCloudlet(this);
        findCloudlet.setRequest(request, host, port, timeoutInMilliseconds);
        return findCloudlet.call();
    }


    /**
     * findCloudlet finds the closest cloudlet instance as per request. Returns a Future.
     * @param context
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<FindCloudletReply> findCloudletFuture(Context context,
                                          FindCloudletRequest request,
                                          long timeoutInMilliseconds) {
        String carrierName = retrieveNetworkCarrierName(context);
        return findCloudletFuture(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }

    /**
     * findCloudletFuture finds the closest cloudlet instance as per request. Returns a Future.
     * @param request
     * @param host Distributed Matching Engine hostname
     * @param port Distributed Matching Engine port
     * @param timeoutInMilliseconds
     * @return cloudlet URI Future.
     */
    public Future<FindCloudletReply> findCloudletFuture(FindCloudletRequest request,
                                                        String host, int port,
                                                        long timeoutInMilliseconds) {
        FindCloudlet findCloudlet = new FindCloudlet(this);
        findCloudlet.setRequest(request, host, port, timeoutInMilliseconds);
        return submit(findCloudlet);
    }


    /**
     * verifyLocationFuture validates the client submitted information against known network
     * parameters on the subscriber network side.
     * @param context
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws StatusRuntimeException
     * @throws InterruptedException
     * @throws IOException
     * @throws ExecutionException
     */
    public VerifyLocationReply verifyLocation(Context context, VerifyLocationRequest request,
                                             long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, IOException, ExecutionException {
        String carrierName = retrieveNetworkCarrierName(context);
        return verifyLocation(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }
    /**
     * verifyLocationFuture validates the client submitted information against known network
     * parameters on the subscriber network side.
     * @param request
     * @param host Distributed Matching Engine hostname
     * @param port Distributed Matching Engine port
     * @param timeoutInMilliseconds
     * @return boolean validated or not.
     * @throws StatusRuntimeException
     * @throws InterruptedException
     * @throws IOException
     */
    public VerifyLocationReply verifyLocation(VerifyLocationRequest request,
                                              String host, int port,
                                              long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, IOException, ExecutionException {
        VerifyLocation verifyLocation = new VerifyLocation(this);
        verifyLocation.setRequest(request, host, port, timeoutInMilliseconds);
        return verifyLocation.call();
    }

    /**
     * verifyLocationFuture validates the client submitted information against known network
     * parameters on the subscriber network side. Returns a future.
     * @param context
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<VerifyLocationReply> verifyLocationFuture(Context context,
                                                            VerifyLocationRequest request,
                                                            long timeoutInMilliseconds) {
        String carrierName = retrieveNetworkCarrierName(context);
        return verifyLocationFuture(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }
    /**
     * verifyLocationFuture validates the client submitted information against known network
     * parameters on the subscriber network side. Returns a future.
     * @param request
     * @return Future<Boolean> validated or not.
     */
    public Future<VerifyLocationReply> verifyLocationFuture(VerifyLocationRequest request,
                                                            String host, int port,
                                                            long timeoutInMilliseconds) {
        VerifyLocation verifyLocation = new VerifyLocation(this);
        verifyLocation.setRequest(request, host, port, timeoutInMilliseconds);
        return submit(verifyLocation);
    }

    /**
     * getLocation returns the network verified location of this device.
     * @param context
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws StatusRuntimeException
     * @throws InterruptedException
     * @throws ExecutionException
     */
    public GetLocationReply getLocation(Context context,
                                        GetLocationRequest request,
                                        long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, ExecutionException {
        String carrierName = retrieveNetworkCarrierName(context);
        return getLocation(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }
    /**
     * getLocation returns the network verified location of this device.
     * @param request
     * @param host Distributed Matching Engine hostname
     * @param port Distributed Matching Engine port
     * @param timeoutInMilliseconds
     * @return
     * @throws StatusRuntimeException
     */
    public GetLocationReply getLocation(GetLocationRequest request,
                                        String host, int port,
                                        long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, ExecutionException {
        GetLocation getLocation = new GetLocation(this);
        getLocation.setRequest(request, host, port, timeoutInMilliseconds);
        return getLocation.call();
    }

    /**
     * getLocation returns the network verified location of this device. Returns a Future.
     * @param context
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<GetLocationReply> getLocationFuture(Context context,
                                                      GetLocationRequest request,
                                                      long timeoutInMilliseconds) {
        String carrierName = retrieveNetworkCarrierName(context);
        return getLocationFuture(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }
    /**
     * getLocation returns the network verified location of this device. Returns a Future.
     * @param request
     * @param host Distributed Matching Engine hostname
     * @param port Distributed Matching Engine port
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<GetLocationReply> getLocationFuture(GetLocationRequest request,
                                                      String host, int port,
                                                      long timeoutInMilliseconds) {
        GetLocation getLocation = new GetLocation(this);
        getLocation.setRequest(request, host, port, timeoutInMilliseconds);
        return submit(getLocation);
    }


    /**
     * addUserToGroup is a blocking call.
     * @param context
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws InterruptedException
     * @throws ExecutionException
     */
    public DynamicLocGroupReply addUserToGroup(Context context, DynamicLocGroupRequest request,
                                               long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {
        String carrierName = retrieveNetworkCarrierName(context);
        return addUserToGroup(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }
    /**
     * addUserToGroup is a blocking call.
     * @param request
     * @param host Distributed Matching Engine hostname
     * @param port Distributed Matching Engine port
     * @param timeoutInMilliseconds
     * @return
     */
    public DynamicLocGroupReply addUserToGroup(DynamicLocGroupRequest request,
                                               String host, int port,
                                               long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {
        AddUserToGroup addUserToGroup = new AddUserToGroup(this);
        addUserToGroup.setRequest(request, host, port, timeoutInMilliseconds);
        return addUserToGroup.call();
    }

    /**
     * addUserToGroupFuture
     * @param context
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<DynamicLocGroupReply> addUserToGroupFuture(Context context,
                                                             DynamicLocGroupRequest request,
                                                             long timeoutInMilliseconds) {
        String carrierName = retrieveNetworkCarrierName(context);
        return addUserToGroupFuture(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }
    /**
     * addUserToGroupFuture
     * @param request
     * @param host Distributed Matching Engine hostname
     * @param port Distributed Matching Engine port
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<DynamicLocGroupReply> addUserToGroupFuture(DynamicLocGroupRequest request,
                                                             String host, int port,
                                                             long timeoutInMilliseconds) {
        AddUserToGroup addUserToGroup = new AddUserToGroup(this);
        addUserToGroup.setRequest(request, host, port, timeoutInMilliseconds);
        return submit(addUserToGroup);
    }

    /**
     * Retrieve nearby AppInsts for registered application. This is a blocking call.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws InterruptedException
     * @throws ExecutionException
     */
    public AppInstListReply getAppInstList(AppInstListRequest request,
                                           long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {
        String carrierName = request.getCarrierName();
        return getAppInstList(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }

    /**
     * Retrieve nearby AppInsts for registered application. This is a blocking call.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public AppInstListReply getAppInstList(AppInstListRequest request,
                                           String host, int port,
                                           long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {
        GetAppInstList getAppInstList = new GetAppInstList(this);
        getAppInstList.setRequest(request, host, port, timeoutInMilliseconds);
        return getAppInstList.call();
    }


    /**
     * Retrieve nearby AppInsts for registered application. Returns a Future.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<AppInstListReply> getAppInstListFuture(AppInstListRequest request,
                                                         long timeoutInMilliseconds) {

        String carrierName = request.getCarrierName();
        return getAppInstListFuture(request, generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
    }
    /**
     * Retrieve nearby AppInsts for registered application. Returns a Future.
     * @param request
     * @param host
     * @param port
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<AppInstListReply> getAppInstListFuture(AppInstListRequest request,
                                                         String host, int port,
                                                         long timeoutInMilliseconds) {
        GetAppInstList getAppInstList = new GetAppInstList(this);
        getAppInstList.setRequest(request, host, port, timeoutInMilliseconds);
        return submit(getAppInstList);
    }


    /**
     * Request QOS values from a list of PositionKPIRequests, and returns a stream Iterator of
     * predicted QOS values.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws InterruptedException
     * @throws ExecutionException
     */
    public ChannelIterator<QosPositionKpiReply> getQosPositionKpi(QosPositionKpiRequest request,
                                                                  long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {

        String carrierName = retrieveNetworkCarrierName(mContext);
        QosPositionKpi qosPositionKpi = new QosPositionKpi(this);
        qosPositionKpi.setRequest(request,generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
        return qosPositionKpi.call();
    }

    /**
     * Request QOS values from a list of PositionKPIRequests, and returns an asynchronous Future
     * for a stream Iterator of predicted QOS values.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws InterruptedException
     * @throws ExecutionException
     */
    public Future<ChannelIterator<QosPositionKpiReply>> getQosPositionKpiFuture(QosPositionKpiRequest request,
                                                                  long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {

        String carrierName = retrieveNetworkCarrierName(mContext);
        QosPositionKpi qosPositionKpi = new QosPositionKpi(this);
        qosPositionKpi.setRequest(request,generateDmeHostAddress(carrierName), getPort(), timeoutInMilliseconds);
        return submit(qosPositionKpi);
    }
    /**
     * Request QOS values from a list of PositionKPIRequests, and returns a stream Iterator of
     * predicted QOS values.
     * @param request
     * @param host
     * @param port
     * @param timeoutInMilliseconds
     * @return
     * @throws InterruptedException
     * @throws ExecutionException
     */
    public ChannelIterator<QosPositionKpiReply> getQosPositionKpi(QosPositionKpiRequest request,
                                                                  String host, int port,
                                                                  long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {

        QosPositionKpi qosPositionKpi = new QosPositionKpi(this);
        qosPositionKpi.setRequest(request, host, port, timeoutInMilliseconds);
        return qosPositionKpi.call();
    }

    /**
     * Request QOS values from a list of PositionKPIRequests, and returns an asynchronous Future
     * for a stream Iterator of predicted QOS values.
     *
     * @param request
     * @param host
     * @param port
     * @param timeoutInMilliseconds
     * @return
     * @throws InterruptedException
     * @throws ExecutionException
     */
    public Future<ChannelIterator<QosPositionKpiReply>> getQosPositionKpiFuture(QosPositionKpiRequest request,
                                                                                String host, int port,
                                                                                long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {

        QosPositionKpi qosPositionKpi = new QosPositionKpi(this);
        qosPositionKpi.setRequest(request, host, port, timeoutInMilliseconds);
        return submit(qosPositionKpi);
    }

    //
    // Network Wrappers:
    //

    /**
     * Returns if the bound Data Network for application is currently roaming or not.
     * @return
     */
    @RequiresApi(api = android.os.Build.VERSION_CODES.P)
    public boolean isRoamingData() {
        return mNetworkManager.isRoamingData();
    }

    /**
     * Returns whether Wifi is enabled on the system or not, independent of Application's network state.
     * @param context
     * @return
     */
    public boolean isWiFiEnabled(Context context) {
        WifiManager wifiManager = (WifiManager)context.getSystemService(Context.WIFI_SERVICE);
        return wifiManager.isWifiEnabled();
    }

    /**
     * Checks if the Carrier + Phone combination supports WiFiCalling. This does not return whether it is enabled.
     * If under roaming conditions, WiFi Calling may disable cellular network data interfaces needed by
     * cellular data only Distributed Matching Engine and Cloudlet network operations.
     *
     * @return
     */
    public boolean isWiFiCallingSupported(CarrierConfigManager carrierConfigManager) {
        return mNetworkManager.isWiFiCallingSupported(carrierConfigManager);
    }


    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }

    void setTokenServerURI(String tokenFollowURI) {
        mTokenServerURI = tokenFollowURI;
    }

    String getTokenServerURI() {
        return mTokenServerURI;
    }

    void setTokenServerToken(String token) {
        mTokenServerToken = token;
    }

    String getTokenServerToken() {
        return mTokenServerToken;
    }


    public boolean isSSLEnabled() {
        return isSSLEnabled;
    }

    public void setSSLEnabled(boolean SSLEnabled) {
        isSSLEnabled = SSLEnabled;
    }

    /**
     * Helper function to return a channel that handles SSL,
     * or returns a more basic ManagedChannelBuilder.
     * @param host
     * @param port
     * @return
     */
    ManagedChannel channelPicker(String host, int port) {
        if (isSSLEnabled()) {
            return OkHttpChannelBuilder // Public certs only.
                    .forAddress(host, port)
                    .build();
        } else {
            return ManagedChannelBuilder
                    .forAddress(host, port)
                    .usePlaintext()
                    .build();
        }
    }

}
