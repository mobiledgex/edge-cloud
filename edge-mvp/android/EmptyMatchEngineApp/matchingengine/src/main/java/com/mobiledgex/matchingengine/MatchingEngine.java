package com.mobiledgex.matchingengine;

import android.content.Context;
import android.content.pm.ApplicationInfo;
import android.content.pm.PackageManager;
import android.support.annotation.NonNull;
import android.telephony.NeighboringCellInfo;
import android.telephony.TelephonyManager;

import com.google.protobuf.ByteString;

import java.io.IOException;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.Callable;
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
    private final String mInitalDMEContactHost = "tdg.dme.mobiledgex.net";
    private String mCurrentNetworkOperatorName = "";
    private String host = "tdg.dme.mobiledgex.net"; // FIXME: Need CarrierName from actual SIM card to generate.
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

    public MatchingEngine() {
        threadpool = Executors.newSingleThreadExecutor();
    }
    public MatchingEngine(ExecutorService executorService) {
        threadpool = executorService;
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

    String getCurrentNetworkOperatorName() {
        return mCurrentNetworkOperatorName;
    }

    void setCurrentNetworkOperatorName(String networkOperatorName) {
        this.mCurrentNetworkOperatorName = networkOperatorName;
    }

    private void updateDmeHostAddress(String networkOperatorName) {
        setCurrentNetworkOperatorName(networkOperatorName);
        this.host = getCurrentNetworkOperatorName() + ".dme.mobiledgex.net";
    }

    /**
     * The library itself will not directly ask for permissions, the application should before use.
     * This keeps responsibilities managed clearly in one spot under the app's control.
     */
    public Match_Engine_Request createRequest(Context context, android.location.Location loc) throws SecurityException {
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
        String networkOperatorName = telManager.getNetworkOperatorName();
        // READ_PHONE_STATE or
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


        // also update MatchingEngine state.
        // FIXME: NetworkManager callback.
        updateDmeHostAddress(networkOperatorName);
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

        TelephonyManager telManager = (TelephonyManager)context.getSystemService(Context.TELEPHONY_SERVICE);
        String networkOperatorName = telManager.getNetworkOperatorName();

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
    public AppClient.Match_Engine_Status registerClient(AppClient.Match_Engine_Request request, long timeoutInMilliseconds)
            throws StatusRuntimeException, IOException {
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
    public Future<AppClient.Match_Engine_Status> registerClientFuture(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
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
    public FindCloudletResponse findCloudlet(AppClient.Match_Engine_Request request, long timeoutInMilliseconds)
            throws StatusRuntimeException {
        FindCloudlet findCloudlet = new FindCloudlet(this);
        findCloudlet.setRequest(request, timeoutInMilliseconds);
        return findCloudlet.call();
    }

    /**
     * findCloudlet finds the closest cloudlet instance as per request. Returns a Future.
     * @param request
     * @return cloudlet URI Future.
     */
    public Future<FindCloudletResponse> findCloudletFuture(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
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
     */
    public AppClient.Match_Engine_Loc_Verify verifyLocation(AppClient.Match_Engine_Request request, long timeoutInMilliseconds)
            throws StatusRuntimeException, IOException {
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
    public Future<AppClient.Match_Engine_Loc_Verify> verifyLocationFuture(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
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
    public AppClient.Match_Engine_Loc getLocation(AppClient.Match_Engine_Request request, long timeoutInMilliseconds)
            throws StatusRuntimeException {
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
    public Future<AppClient.Match_Engine_Loc> getLocationFuture(AppClient.Match_Engine_Request request, long timeoutInMilliseconds) {
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
    public AppClient.Match_Engine_Status addUserToGroup(AppClient.DynamicLocGroupAdd request, long timeoutInMilliseconds) {
        AddUserToGroup addUserToGroup = new AddUserToGroup(this);
        addUserToGroup.setRequest(request, timeoutInMilliseconds);
        return addUserToGroup.call();
    }

    public Future<AppClient.Match_Engine_Status> addUserToGroupFuture(AppClient.DynamicLocGroupAdd request, long timeoutInMilliseconds) {
        AddUserToGroup addUserToGroup = new AddUserToGroup(this);
        addUserToGroup.setRequest(request, timeoutInMilliseconds);
        return submit(addUserToGroup);
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
