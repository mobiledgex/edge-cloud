package com.mobiledgex.matchingengine;

import android.content.Context;
import android.content.pm.ApplicationInfo;
import android.content.pm.PackageManager;
import android.net.ConnectivityManager;
import android.net.wifi.WifiManager;
import android.provider.Settings;
import android.support.annotation.NonNull;
import android.support.annotation.RequiresApi;
import android.telephony.CarrierConfigManager;
import android.telephony.NeighboringCellInfo;
import android.telephony.TelephonyManager;

import com.google.protobuf.ByteString;

import java.io.IOException;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import distributed_match_engine.AppClient;
import distributed_match_engine.AppClient.Match_Engine_Request;
import distributed_match_engine.LocOuterClass.Loc;
import io.grpc.StatusRuntimeException;

import android.content.pm.PackageInfo;
import android.util.Log;


// TODO: GRPC (which needs http/2).
public class MatchingEngine {
    public static final String TAG = "MatchingEngine";
    private final String mInitialDMEContactHost = "tdg.dme.mobiledgex.net";
    private String host = mInitialDMEContactHost;
    private NetworkManager mNetworkManager;
    private int port = 50051;

    // A threadpool for all the MatchEngine API callable interfaces:
    final ExecutorService threadpool;

    // State info for engine
    private AppClient.Match_Engine_Status mStatus;
    private UUID mUUID;
    private String mSessionCookie;
    private String mTokenServerURI;
    private String mTokenServerToken;
    private AppClient.Match_Engine_Reply mMatchEngineFindCloudletReply; // FindCloudlet.
    private AppClient.Match_Engine_Status mMatchEngineStatus;
    private AppClient.Match_Engine_Loc mMatchEngineLocation;
    private AppClient.Match_Engine_Loc_Verify mMatchEngineLocationVerify;

    public MatchingEngine(Context context) {
        threadpool = Executors.newSingleThreadExecutor();
        ConnectivityManager connectivityManager = context.getSystemService(ConnectivityManager.class);
        mNetworkManager = NetworkManager.getSingleton(connectivityManager);
    }
    public MatchingEngine(Context context, ExecutorService executorService) {
        threadpool = executorService;
        ConnectivityManager connectivityManager = context.getSystemService(ConnectivityManager.class);
        mNetworkManager = NetworkManager.getSingleton(connectivityManager, threadpool);
    }

    // Application state Bundle Key.
    public static final String MEX_LOCATION_PERMISSION = "MEX_LOCATION_PERMISSION";
    private static boolean mMexLocationAllowed = false;

    public static boolean isMexLocationAllowed() {
        return mMexLocationAllowed;
    }

    public static void setMexLocationAllowed(boolean allowMexLocation) {
        mMexLocationAllowed = allowMexLocation;
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

    public UUID getUUID() {
        return mUUID;
    }

    public void setUUID(UUID uuid) {
        mUUID = uuid;
    }

    public UUID createUUID() {
        mUUID = UUID.randomUUID();
        return mUUID;
    }

    void setSessionCookie(String sessionCookie) {
        this.mSessionCookie = sessionCookie;
    }
    String getSessionCookie() {
        return this.mSessionCookie;
    }

    void setMatchEngineStatus(AppClient.Match_Engine_Status status) {
        mMatchEngineStatus = status;
    }

    void setMatchEngineLocation(AppClient.Match_Engine_Loc location) {
        mMatchEngineLocation = location;
    }

    void setMatchEngineLocationVerify(AppClient.Match_Engine_Loc_Verify locationVerify) {
        mMatchEngineLocationVerify = locationVerify;
    }

    void setFindCloudletResponse(AppClient.Match_Engine_Reply reply) {
        mMatchEngineFindCloudletReply = reply;
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


    /**
     * Creates a MatchingEngineRequest going to the Distributed Matching Engine (DME). The library
     * itself will not directly ask for permissions, the application should before use.
     * @param context
     * @param loc
     * @return
     * @throws SecurityException
     */
    public MatchingEngineRequest createRequest(Context context, android.location.Location loc) throws SecurityException {
        String dmeHost = generateDmeHostAddress(retrieveNetworkCarrierName(context));
        MatchingEngineRequest request = createRequest(context, dmeHost, getPort(), loc);
        return request;
    }

    /**
     * Creates a MatchingEngineRequest going to the Distributed Matching Engine (DME). The library
     * itself will not directly ask for permissions, the application should before use.
     * @param context
     * @param host
     * @param port
     * @param loc
     * @return
     * @throws SecurityException
     */
    public MatchingEngineRequest createRequest(Context context, String host, int port, android.location.Location loc) throws SecurityException {
        Match_Engine_Request grpcRequest = createGRPCRequest(context, retrieveNetworkCarrierName(context), loc);
        MatchingEngineRequest matchingEngineRequest = new MatchingEngineRequest(grpcRequest, host, port);
        return matchingEngineRequest;
    }

    Match_Engine_Request createGRPCRequest(Context context, String networkOperatorName, android.location.Location loc) throws SecurityException {
        if (context == null) {
            throw new IllegalArgumentException("MatchingEngine requires a working application context.");
        }

        if (!mMexLocationAllowed) {
            Log.d(TAG, "Create Request disabled. Matching engine is not configured to allow use.");
            return null;
        }

        if (loc == null) {
            throw new IllegalArgumentException("Location parameter is required.");
        }

        if (mUUID == null) {
            throw new IllegalArgumentException("UUID is not set, and is required.");
        }

        // Operator
        TelephonyManager telManager = (TelephonyManager)context.getSystemService(Context.TELEPHONY_SERVICE);
        String carrierName = retrieveNetworkCarrierName(context);
        // READ_PHONE_STATE. FIXME: May
        AppClient.IDTypes id_types = AppClient.IDTypes.IPADDR;
        String id = telManager.getLine1Number(); // NOT IMEI, if this throws a SecurityException, application must handle it.
        String mnc = telManager.getNetworkOperator();
        String mcc = telManager.getNetworkCountryIso();

        if (id == null) { // Fallback to IP:
	    // TODO: Dual SIM?
        }

        // Tower
        List<NeighboringCellInfo> neighbors = telManager.getNeighboringCellInfo();
        int lac = 0;
        int cid = 0;
        if (neighbors.size() > 0) {
            lac = neighbors.get(0).getLac();
            cid = neighbors.get(0).getCid();
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
        String versionCode = "";
        String appName = packageLabel;
        try {
            pInfo = context.getPackageManager().getPackageInfo(context.getPackageName(), 0);
            versionName = pInfo.versionName;
            versionCode = (new Integer(pInfo.versionCode)).toString();
        } catch (PackageManager.NameNotFoundException nfe) {
            nfe.printStackTrace();
            // Hard stop, or continue?
        }

        // Passed in Location (which is a callback interface)
        Loc aLoc = androidLocToMexLoc(loc);

        Match_Engine_Request request = AppClient.Match_Engine_Request.newBuilder()
                .setVer(5)
                .setIdType(id_types)
                .setUuid(mUUID.toString())
                .setId((id == null) ? "" : id)
                .setCarrierID(3l) // uint64 --> String? mnc, mcc?
                .setCarrierName(networkOperatorName.equals("") ? mnc : networkOperatorName) // Carrier Name or Mnc?
                .setTower(cid) // cid and lac (int)
                .setGpsLocation(aLoc)
                .setAppId(5011l) // uint64 --> String again. TODO: Clarify use.
                .setProtocol(ByteString.copyFromUtf8("http")) // This one is appId context sensitive.
                .setServerPort(ByteString.copyFromUtf8("1234")) // App dependent.
                .setDevName(packageLabel) // From signing certificate?
                .setAppName(appName)
                .setAppVers(versionName) // Or versionName, which is visual name?
                .setSessionCookie(mSessionCookie == null ? "" : mSessionCookie) // "" if null/unknown.
                .setVerifyLocToken(mTokenServerToken == null ? "" : mTokenServerToken)
                .build();

        return request;
    }

    public AppClient.DynamicLocGroupAdd createDynamicLocationGroupAdd(Context context,
                                                                      long groupLocationId,
                                                                      @NonNull
                                                                      AppClient.DynamicLocGroupAdd.DlgCommType type,
                                                                      @NonNull
                                                                      android.location.Location loc,
                                                                      String userData)
            throws SecurityException {

        if (context == null) {
            throw new IllegalArgumentException("MatchingEngine requires a working application context.");
        }

        if (!mMexLocationAllowed) {
            Log.d(TAG, "Create Request disabled. Matching engine is not configured to allow use.");
            return null;
        }

        if (loc == null) {
            throw new IllegalStateException("Location parameter is required.");
        }

        if (mUUID == null) {
            throw new IllegalArgumentException("UUID is not set, and is required.");
        }

        String networkOperatorName = retrieveNetworkCarrierName(context);

        TelephonyManager tm = (TelephonyManager) context.getSystemService(Context.TELEPHONY_SERVICE);
        List<NeighboringCellInfo> neighbors = tm.getNeighboringCellInfo();
        int lac = 0;
        int cid = 0;
        if (neighbors.size() > 0) {
            lac = neighbors.get(0).getLac();
            cid = neighbors.get(0).getCid();
        }

        Loc aLoc = androidLocToMexLoc(loc);

        AppClient.DynamicLocGroupAdd groupAddRequest = AppClient.DynamicLocGroupAdd.newBuilder()
                .setVer(0)
                .setIdType(AppClient.IDTypes.IPADDR)
                .setCarrierID(0)
                .setCarrierName(networkOperatorName == null ? "" : networkOperatorName)
                .setTower(cid)
                .setGpsLocation(aLoc)
                .setLgId(groupLocationId)
                .setSessionCookie(mSessionCookie)
                .setCommType(type)
                .setUserData(userData)
                .build();

        return groupAddRequest;

    }

    private Loc androidLocToMexLoc(android.location.Location loc) {
        return Loc.newBuilder()
                .setLat((loc == null) ? 0.0d : loc.getLatitude())
                .setLong((loc == null) ? 0.0d : loc.getLongitude())
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
     */
    public AppClient.Match_Engine_Status registerClient(MatchingEngineRequest request, long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, ExecutionException {
        RegisterClient registerClient = new RegisterClient(this);
        registerClient.setRequest(request, timeoutInMilliseconds);
        return registerClient.call();
    }

    /**
     * Registers device on the MatchingEngine server. Returns a Future.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<AppClient.Match_Engine_Status> registerClientFuture(MatchingEngineRequest request, long timeoutInMilliseconds) {
        RegisterClient registerClient = new RegisterClient(this);
        registerClient.setRequest(request, timeoutInMilliseconds);
        return submit(registerClient);
    }

    /**
     * findCloudlet finds the closest cloudlet instance as per request.
     * @param request
     * @return cloudlet URI.
     * @throws StatusRuntimeException
     */
    public FindCloudletResponse findCloudlet(MatchingEngineRequest request, long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, ExecutionException {
        FindCloudlet findCloudlet = new FindCloudlet(this);
        findCloudlet.setRequest(request, timeoutInMilliseconds);
        return findCloudlet.call();
    }

    /**
     * findCloudlet finds the closest cloudlet instance as per request. Returns a Future.
     * @param request
     * @return cloudlet URI Future.
     */
    public Future<FindCloudletResponse> findCloudletFuture(MatchingEngineRequest request, long timeoutInMilliseconds) {
        FindCloudlet findCloudlet = new FindCloudlet(this);
        findCloudlet.setRequest(request, timeoutInMilliseconds);
        return submit(findCloudlet);
    }


    /**
     * verifyLocationFuture validates the client submitted information against known network
     * parameters on the subscriber network side.
     * @param request
     * @return boolean validated or not.
     * @throws StatusRuntimeException
     * @throws InterruptedException
     * @throws IOException
     */
    public AppClient.Match_Engine_Loc_Verify verifyLocation(MatchingEngineRequest request, long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, IOException, ExecutionException {
        VerifyLocation verifyLocation = new VerifyLocation(this);
        verifyLocation.setRequest(request, timeoutInMilliseconds);
        return verifyLocation.call();
    }

    /**
     * verifyLocationFuture validates the client submitted information against known network
     * parameters on the subscriber network side. Returns a future.
     * @param request
     * @return Future<Boolean> validated or not.
     */
    public Future<AppClient.Match_Engine_Loc_Verify> verifyLocationFuture(MatchingEngineRequest request, long timeoutInMilliseconds) {
        VerifyLocation verifyLocation = new VerifyLocation(this);
        verifyLocation.setRequest(request, timeoutInMilliseconds);
        return submit(verifyLocation);
    }

    /**
     * getLocation returns the network verified location of this device.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     * @throws StatusRuntimeException
     */
    public AppClient.Match_Engine_Loc getLocation(MatchingEngineRequest request, long timeoutInMilliseconds)
            throws StatusRuntimeException, InterruptedException, ExecutionException {
        GetLocation getLocation = new GetLocation(this);
        getLocation.setRequest(request, timeoutInMilliseconds);
        return getLocation.call();
    }

    /**
     * getLocation returns the network verified location of this device. Returns a Future.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public Future<AppClient.Match_Engine_Loc> getLocationFuture(MatchingEngineRequest request, long timeoutInMilliseconds) {
        GetLocation getLocation = new GetLocation(this);
        getLocation.setRequest(request, timeoutInMilliseconds);
        return submit(getLocation);
    }

    /**
     * addUserToGroup is a blocking call.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public AppClient.Match_Engine_Status addUserToGroup(DynamicLocationGroupAdd request, long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {
        AddUserToGroup addUserToGroup = new AddUserToGroup(this);
        addUserToGroup.setRequest(request, timeoutInMilliseconds);
        return addUserToGroup.call();
    }

    public Future<AppClient.Match_Engine_Status> addUserToGroupFuture(DynamicLocationGroupAdd request, long timeoutInMilliseconds) {
        AddUserToGroup addUserToGroup = new AddUserToGroup(this);
        addUserToGroup.setRequest(request, timeoutInMilliseconds);
        return submit(addUserToGroup);
    }

    /**
     * Retrieve nearby Cloudlets (or AppInsts) for registered application. This is a blocking call.
     * @param request
     * @param timeoutInMilliseconds
     * @return
     */
    public AppClient.Match_Engine_Cloudlet_List getCloudletList(MatchingEngineRequest request, long timeoutInMilliseconds)
            throws InterruptedException, ExecutionException {
        GetCloudletList getCloudletList = new GetCloudletList(this);
        getCloudletList.setRequest(request, timeoutInMilliseconds);
        return getCloudletList.call();
    }

    public Future<AppClient.Match_Engine_Cloudlet_List> getCloudletListFuture(MatchingEngineRequest request, long timeoutInMilliseconds) {
        GetCloudletList getCloudletList = new GetCloudletList(this);
        getCloudletList.setRequest(request, timeoutInMilliseconds);
        return submit(getCloudletList);
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
     * cellular data only Distributed Matching Enigne and Cloudlet network operations.
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
}
